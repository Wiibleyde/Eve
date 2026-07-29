package ui

import "github.com/disgoorg/disgo/discord"

func Texts(components []discord.LayoutComponent) []string {
	var texts []string
	collect := func(c discord.Component) {
		if text, ok := c.(discord.TextDisplayComponent); ok {
			texts = append(texts, text.Content)
		}
	}
	for _, component := range components {
		collect(component)
		if iter, ok := component.(discord.ComponentIter); ok {
			for sub := range iter.SubComponents() {
				collect(sub)
			}
		}
	}
	return texts
}
