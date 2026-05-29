package router

import (
	"Eve/internal/logger"

	"fmt"
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

type CommandHandler func(e *events.ApplicationCommandInteractionCreate)
type ComponentHandler func(e *events.ComponentInteractionCreate)
type ModalHandler func(e *events.ModalSubmitInteractionCreate)

type prefixEntry[H any] struct {
	prefix  string
	handler H
}

type Router struct {
	commands       map[string]CommandHandler
	buttons        map[string]ComponentHandler
	buttonPrefixes []prefixEntry[ComponentHandler]
	modalPrefixes  []prefixEntry[ModalHandler]
}

func New() *Router {
	return &Router{
		commands: make(map[string]CommandHandler),
		buttons:  make(map[string]ComponentHandler),
	}
}

func (r *Router) OnCommand(name string, h CommandHandler) { r.commands[name] = h }
func (r *Router) OnButton(id string, h ComponentHandler)  { r.buttons[id] = h }
func (r *Router) OnButtonPrefix(prefix string, h ComponentHandler) {
	r.buttonPrefixes = append(r.buttonPrefixes, prefixEntry[ComponentHandler]{prefix, h})
}
func (r *Router) OnModal(prefix string, h ModalHandler) {
	r.modalPrefixes = append(r.modalPrefixes, prefixEntry[ModalHandler]{prefix, h})
}

func (r *Router) Attach(client *bot.Client) {
	client.AddEventListeners(
		bot.NewListenerFunc(func(e *events.ApplicationCommandInteractionCreate) {
			data := e.SlashCommandInteractionData()
			logger.Event("Command interaction",
				"path", data.CommandPath(),
				"user", e.User().ID.String(),
				"guild", formatGuildID(e.GuildID()),
				"args", formatOptions(data.Options),
			)
			name := data.CommandName()
			if h, ok := r.commands[name]; ok {
				h(e)
			}
		}),
		bot.NewListenerFunc(func(e *events.ComponentInteractionCreate) {
			id := e.Data.CustomID()
			logger.Event("Component interaction",
				"id", id,
				"user", e.User().ID.String(),
				"guild", formatGuildID(e.GuildID()),
			)
			if h, ok := r.buttons[id]; ok {
				h(e)
				return
			}
			for _, pe := range r.buttonPrefixes {
				if strings.HasPrefix(id, pe.prefix) {
					pe.handler(e)
					return
				}
			}
		}),
		bot.NewListenerFunc(func(e *events.ModalSubmitInteractionCreate) {
			id := e.Data.CustomID
			for _, pe := range r.modalPrefixes {
				if strings.HasPrefix(id, pe.prefix) {
					pe.handler(e)
					return
				}
			}
		}),
	)
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
