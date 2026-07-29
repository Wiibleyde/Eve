package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
)

const (
	AccentBase    = 0xFFFFFF
	AccentSuccess = 0x57F287
	AccentError   = 0xED4245
	AccentWarning = 0xFFA500
)

const (
	TitleSuccess = "Succès"
	TitleError   = "Erreur"
)

const MaxText = 4000

const signature = "Eve – Toujours prête à vous aider."

type Field struct {
	Name   string
	Value  string
	Inline bool
}

type Card struct {
	accent    int
	title     string
	accessory discord.SectionAccessoryComponent
	blocks    []discord.ContainerSubComponent
	footer    string
	hasFooter bool
}

func New() *Card {
	return &Card{accent: AccentBase, footer: signature, hasFooter: true}
}

func Success(message string) *Card {
	return New().Accent(AccentSuccess).Title("✅ " + TitleSuccess).Text(message)
}

func Error(message string) *Card {
	return New().Accent(AccentError).Title("⛔ " + TitleError).Text(message)
}

func Warning(title string, message string) *Card {
	return New().Accent(AccentWarning).Title("⚠️ " + title).Text(message)
}

func (c *Card) Accent(color int) *Card {
	c.accent = color
	return c
}

func (c *Card) Title(title string) *Card {
	c.title = title
	return c
}

func (c *Card) Titlef(format string, a ...any) *Card {
	return c.Title(fmt.Sprintf(format, a...))
}

func (c *Card) Text(text string) *Card {
	if strings.TrimSpace(text) == "" {
		return c
	}
	return c.block(discord.NewTextDisplay(text))
}

func (c *Card) Textf(format string, a ...any) *Card {
	return c.Text(fmt.Sprintf(format, a...))
}

func (c *Card) Heading(text string) *Card {
	return c.Text("### " + text)
}

func (c *Card) Subtext(text string) *Card {
	return c.Text("-# " + text)
}

func (c *Card) Subtextf(format string, a ...any) *Card {
	return c.Subtext(fmt.Sprintf(format, a...))
}

func (c *Card) Fields(fields ...Field) *Card {
	lines := make([]string, 0, len(fields))
	for _, f := range fields {
		if strings.TrimSpace(f.Value) == "" {
			continue
		}
		if f.Inline {
			lines = append(lines, "**"+f.Name+"** · "+f.Value)
			continue
		}
		lines = append(lines, "**"+f.Name+"**\n"+f.Value)
	}
	return c.Text(strings.Join(lines, "\n"))
}

func (c *Card) Divider() *Card {
	return c.block(discord.NewSmallSeparator())
}

func (c *Card) Space() *Card {
	return c.block(discord.NewSmallSeparator().WithDivider(false))
}

func (c *Card) Accessory(accessory discord.SectionAccessoryComponent) *Card {
	c.accessory = accessory
	return c
}

func (c *Card) Thumbnail(url string) *Card {
	if url == "" {
		return c
	}
	return c.Accessory(discord.NewThumbnail(url))
}

func (c *Card) Image(url string) *Card {
	if url == "" {
		return c
	}
	return c.Images(url)
}

func (c *Card) Images(urls ...string) *Card {
	items := make([]discord.MediaGalleryItem, 0, len(urls))
	for _, url := range urls {
		if url == "" {
			continue
		}
		items = append(items, discord.MediaGalleryItem{Media: discord.UnfurledMediaItem{URL: url}})
	}
	if len(items) == 0 {
		return c
	}
	return c.block(discord.NewMediaGallery(items...))
}

func (c *Card) Row(components ...discord.InteractiveComponent) *Card {
	if len(components) == 0 {
		return c
	}
	return c.block(discord.NewActionRow(components...))
}

func (c *Card) Footer(text string) *Card {
	c.footer = text
	c.hasFooter = true
	return c
}

func (c *Card) NoFooter() *Card {
	c.hasFooter = false
	return c
}

func (c *Card) Container() discord.ContainerComponent {
	return discord.NewContainer(c.subComponents()...).WithAccentColor(c.accent)
}

func (c *Card) Components() []discord.LayoutComponent {
	return []discord.LayoutComponent{c.Container()}
}

func (c *Card) MessageCreate() discord.MessageCreate {
	return discord.MessageCreate{
		Flags:           discord.MessageFlagIsComponentsV2,
		Components:      c.Components(),
		AllowedMentions: NoMentions(),
	}
}

func (c *Card) EphemeralCreate() discord.MessageCreate {
	msg := c.MessageCreate()
	msg.Flags = msg.Flags.Add(discord.MessageFlagEphemeral)
	return msg
}

func (c *Card) MessageUpdate() discord.MessageUpdate {
	update := discord.NewMessageUpdateV2(c.Components()...).
		ClearContent().
		ClearEmbeds()
	update.AllowedMentions = NoMentions()
	return update
}

func NoMentions() *discord.AllowedMentions {
	return &discord.AllowedMentions{Parse: []discord.AllowedMentionType{}}
}

func (c *Card) block(component discord.ContainerSubComponent) *Card {
	c.blocks = append(c.blocks, component)
	return c
}

func (c *Card) subComponents() []discord.ContainerSubComponent {
	components := make([]discord.ContainerSubComponent, 0, len(c.blocks)+3)
	blocks := c.blocks

	switch {
	case c.title != "" && c.accessory != nil:
		components = append(components, discord.NewSection(discord.NewTextDisplay("## "+c.title)).WithAccessory(c.accessory))
	case c.title != "":
		components = append(components, discord.NewTextDisplay("## "+c.title))
	case c.accessory != nil && len(blocks) > 0:
		if text, ok := blocks[0].(discord.TextDisplayComponent); ok {
			components = append(components, discord.NewSection(text).WithAccessory(c.accessory))
			blocks = blocks[1:]
		}
	}

	components = append(components, blocks...)

	if c.hasFooter && c.footer != "" {
		components = append(components,
			discord.NewSmallSeparator(),
			discord.NewTextDisplayf("-# %s • <t:%d:f>", c.footer, time.Now().Unix()),
		)
	}
	return components
}
