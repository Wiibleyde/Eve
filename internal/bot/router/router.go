package router

import (
	"Eve/internal/bot/helpers"
	"Eve/internal/bot/maintenance"
	"Eve/internal/bot/ui"
	"Eve/internal/logger"

	"fmt"
	"runtime/debug"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

const CustomIDSeparator = ":"

const (
	MsgUnknownInteraction = "Interaction inconnue ou expirée."
	MsgHandlerError       = "Une erreur est survenue lors du traitement de cette interaction."
	MsgMaintenance        = "Le bot est actuellement en mode maintenance. Veuillez réessayer plus tard."
)

type (
	CommandHandler    func(e *events.ApplicationCommandInteractionCreate)
	ButtonHandler     func(e *events.ComponentInteractionCreate, args []string)
	SelectMenuHandler func(e *events.ComponentInteractionCreate, args []string)
	ModalHandler      func(e *events.ModalSubmitInteractionCreate, args []string)
	UserCtxHandler    func(e *events.ApplicationCommandInteractionCreate)
	MessageCtxHandler func(e *events.ApplicationCommandInteractionCreate)
)

type componentEntry struct {
	kind string
	fn   func(e *events.ComponentInteractionCreate, args []string)
}

const (
	kindButton     = "button"
	kindSelectMenu = "select_menu"
)

type Router struct {
	commands   map[string]CommandHandler
	components map[string]componentEntry
	modals     map[string]ModalHandler
	userCtx    map[string]UserCtxHandler
	messageCtx map[string]MessageCtxHandler
}

func New() *Router {
	return &Router{
		commands:   make(map[string]CommandHandler),
		components: make(map[string]componentEntry),
		modals:     make(map[string]ModalHandler),
		userCtx:    make(map[string]UserCtxHandler),
		messageCtx: make(map[string]MessageCtxHandler),
	}
}

func register[H any](m map[string]H, kind string, key string, h H) {
	if _, exists := m[key]; exists {
		logger.Warn("Duplicate interaction handler registration", "kind", kind, "key", key)
	}
	m[key] = h
}

func (r *Router) OnCommand(name string, h CommandHandler) {
	register(r.commands, "command", name, h)
}

func (r *Router) OnButton(prefix string, h ButtonHandler) {
	register(r.components, kindButton, prefix, componentEntry{kind: kindButton, fn: h})
}

func (r *Router) OnSelectMenu(prefix string, h SelectMenuHandler) {
	register(r.components, kindSelectMenu, prefix, componentEntry{kind: kindSelectMenu, fn: h})
}

func (r *Router) OnModal(prefix string, h ModalHandler) {
	register(r.modals, "modal", prefix, h)
}

func (r *Router) OnUserContextMenu(name string, h UserCtxHandler) {
	register(r.userCtx, "user_context_menu", name, h)
}

func (r *Router) OnMessageContextMenu(name string, h MessageCtxHandler) {
	register(r.messageCtx, "message_context_menu", name, h)
}

func (r *Router) Attach(client *bot.Client) {
	client.AddEventListeners(
		bot.NewListenerFunc(r.handleApplicationCommand),
		bot.NewListenerFunc(r.handleComponent),
		bot.NewListenerFunc(r.handleModal),
	)
}

func (r *Router) handleApplicationCommand(e *events.ApplicationCommandInteractionCreate) {
	switch e.Data.Type() {
	case discord.ApplicationCommandTypeSlash:
		data := e.SlashCommandInteractionData()
		name := data.CommandName()
		logger.Event("Command interaction",
			"path", data.CommandPath(),
			"user", e.User().ID.String(),
			"guild", formatGuildID(e.GuildID()),
			"args", formatOptions(data.Options),
		)
		if blockedByMaintenance(e, e.User().ID) {
			return
		}
		h, ok := r.commands[name]
		if !ok {
			unknown(e, "command", name)
			return
		}
		safeDispatch(e, "command", name, func() { h(e) })

	case discord.ApplicationCommandTypeUser:
		name := e.Data.CommandName()
		logger.Event("User context menu interaction",
			"name", name,
			"user", e.User().ID.String(),
			"guild", formatGuildID(e.GuildID()),
		)
		if blockedByMaintenance(e, e.User().ID) {
			return
		}
		h, ok := r.userCtx[name]
		if !ok {
			unknown(e, "user_context_menu", name)
			return
		}
		safeDispatch(e, "user_context_menu", name, func() { h(e) })

	case discord.ApplicationCommandTypeMessage:
		name := e.Data.CommandName()
		logger.Event("Message context menu interaction",
			"name", name,
			"user", e.User().ID.String(),
			"guild", formatGuildID(e.GuildID()),
		)
		if blockedByMaintenance(e, e.User().ID) {
			return
		}
		h, ok := r.messageCtx[name]
		if !ok {
			unknown(e, "message_context_menu", name)
			return
		}
		safeDispatch(e, "message_context_menu", name, func() { h(e) })

	default:
		logger.Debug("Unhandled application command type",
			"type", int(e.Data.Type()),
			"name", e.Data.CommandName(),
		)
	}
}

func (r *Router) handleComponent(e *events.ComponentInteractionCreate) {
	id := e.Data.CustomID()
	logger.Event("Component interaction",
		"id", id,
		"user", e.User().ID.String(),
		"guild", formatGuildID(e.GuildID()),
	)
	if blockedByMaintenance(e, e.User().ID) {
		return
	}

	key, args, ok := SplitCustomID(id)
	if !ok {
		unknown(e, "component", id)
		return
	}
	entry, ok := r.components[key]
	if !ok {
		unknown(e, "component", id)
		return
	}
	if kind, known := componentKind(e.Data.Type()); !known || kind != entry.kind {
		logger.Debug("Component type does not match the registered handler",
			"id", id, "registered", entry.kind, "type", int(e.Data.Type()),
		)
		unknown(e, "component", id)
		return
	}
	safeDispatch(e, entry.kind, key, func() { entry.fn(e, args) })
}

func (r *Router) handleModal(e *events.ModalSubmitInteractionCreate) {
	id := e.Data.CustomID
	logger.Event("Modal interaction",
		"id", id,
		"user", e.User().ID.String(),
		"guild", formatGuildID(e.GuildID()),
	)
	if blockedByMaintenance(e, e.User().ID) {
		return
	}

	key, args, ok := SplitCustomID(id)
	if !ok {
		unknown(e, "modal", id)
		return
	}
	h, ok := r.modals[key]
	if !ok {
		unknown(e, "modal", id)
		return
	}
	safeDispatch(e, "modal", key, func() { h(e, args) })
}

func componentKind(t discord.ComponentType) (string, bool) {
	switch t {
	case discord.ComponentTypeButton:
		return kindButton, true
	case discord.ComponentTypeStringSelectMenu,
		discord.ComponentTypeUserSelectMenu,
		discord.ComponentTypeRoleSelectMenu,
		discord.ComponentTypeMentionableSelectMenu,
		discord.ComponentTypeChannelSelectMenu:
		return kindSelectMenu, true
	default:
		return "", false
	}
}

func SplitCustomID(id string) (key string, args []string, ok bool) {
	parts := strings.Split(id, CustomIDSeparator)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return "", nil, false
	}
	return parts[0] + CustomIDSeparator + parts[1], parts[2:], true
}

func BuildCustomID(prefix string, args ...string) string {
	if len(args) == 0 {
		return prefix
	}
	return prefix + CustomIDSeparator + strings.Join(args, CustomIDSeparator)
}

func blockedByMaintenance(responder helpers.EphemeralResponder, userID snowflake.ID) bool {
	if !maintenance.Enabled() || helpers.IsOwner(userID) {
		return false
	}
	logger.Debug("Interaction blocked by maintenance mode", "user", userID.String())
	helpers.RespondEphemeralCard(responder, ui.Warning("Maintenance", MsgMaintenance))
	return true
}

func unknown(responder helpers.EphemeralResponder, kind string, id string) {
	logger.Debug("No handler for interaction", "kind", kind, "id", id)
	helpers.RespondEphemeralCard(responder, ui.Error(MsgUnknownInteraction))
}

func safeDispatch(responder helpers.EphemeralResponder, kind string, key string, fn func()) {
	defer func() {
		rec := recover()
		if rec == nil {
			return
		}
		logger.Error("Panic in interaction handler",
			"kind", kind,
			"key", key,
			"panic", fmt.Sprint(rec),
			"stack", strings.ReplaceAll(string(debug.Stack()), "\n", " | "),
		)
		helpers.RespondEphemeralCard(responder, ui.Error(MsgHandlerError))
	}()
	fn()
}

func formatGuildID(id *snowflake.ID) string {
	if id == nil {
		return "DM"
	}
	return id.String()
}

func formatOptions(opts map[string]discord.SlashCommandOption) string {
	if len(opts) == 0 {
		return ""
	}
	parts := make([]string, 0, len(opts))
	for name, opt := range opts {
		parts = append(parts, fmt.Sprintf("%s=%s", name, string(opt.Value)))
	}
	return strings.Join(parts, " ")
}
