package ai

import "strings"

const (
	EmojiNeutral   = "<:eve:1310908312144969760>"
	EmojiHappy     = "<:eveHappy:1310908313457659934>"
	EmojiLove      = "<:eveLove:1310908314820935690>"
	EmojiNervous   = "<:eveNervous:1310908316980875285>"
	EmojiSleep     = "<:eveSleep:1310908318247813121>"
	EmojiSurprised = "<:eveSurprised:1310908319958958120>"
	EmojiUnhappy   = "<:eveUnhappy:1310908322009976883>"
)

var emojis = []string{
	EmojiNeutral,
	EmojiHappy,
	EmojiLove,
	EmojiNervous,
	EmojiSleep,
	EmojiSurprised,
	EmojiUnhappy,
}

func emojiList() string { return strings.Join(emojis, " ") }
