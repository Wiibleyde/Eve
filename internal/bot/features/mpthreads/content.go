package mpthreads

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"Eve/internal/logger"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

const (
	maxMessageLength = 2000
	maxRelayedFiles  = 10
	maxThreadName    = 100

	// Discord's default upload limit is 10 MiB per message on a guild without
	// boosts, and every re-uploaded attachment of a relayed message travels in a
	// single CreateMessage. The budget stays under that limit — as a per-file cap
	// and as a per-message total — so the request is not rejected outright.
	maxUploadBytes = 9 << 20
)

const downloadTimeout = 30 * time.Second

var httpClient = &http.Client{Timeout: downloadTimeout}

// Discord CDN links carry an expiring signature, so staff has to be told to
// fetch them right away.
const (
	msgAttachmentTooBig  = "Pièce jointe « %s » trop lourde pour être ré-uploadée, lien temporaire (expire sous ~24 h) : %s"
	msgAttachmentFailed  = "Pièce jointe « %s » indisponible au téléchargement, lien temporaire (expire sous ~24 h) : %s"
	msgAttachmentRefused = "Pièce jointe « %s » refusée par Discord, lien temporaire (expire sous ~24 h) : %s"
)

// relayAttachment is a downloaded attachment kept in memory so the same bytes
// can be re-sent across relay attempts, plus the source URL used as a fallback
// when the upload itself is refused.
type relayAttachment struct {
	filename    string
	description string
	url         string
	data        []byte
}

func (a relayAttachment) file() *discord.File {
	return discord.NewFile(a.filename, a.description, bytes.NewReader(a.data))
}

func filesOf(attachments []relayAttachment) []*discord.File {
	if len(attachments) == 0 {
		return nil
	}
	files := make([]*discord.File, 0, len(attachments))
	for _, attachment := range attachments {
		files = append(files, attachment.file())
	}
	return files
}

func linksOf(attachments []relayAttachment) []string {
	links := make([]string, 0, len(attachments))
	for _, attachment := range attachments {
		links = append(links, fmt.Sprintf(msgAttachmentRefused, attachment.filename, attachment.url))
	}
	return links
}

// noMentions neutralises every mention of a relayed message. User-authored
// content must never be able to ping @everyone once mirrored into the guild.
func noMentions() *discord.AllowedMentions {
	return &discord.AllowedMentions{
		Parse: []discord.AllowedMentionType{},
		Roles: []snowflake.ID{},
		Users: []snowflake.ID{},
	}
}

// usersOnlyMentions is used for staff → user relays: user mentions stay
// resolvable, @everyone and role pings never leave the guild.
func usersOnlyMentions() *discord.AllowedMentions {
	return &discord.AllowedMentions{
		Parse: []discord.AllowedMentionType{discord.AllowedMentionTypeUsers},
		Roles: []snowflake.ID{},
	}
}

func ptr[T any](v T) *T { return &v }

func threadName(user discord.User) string {
	name := strings.TrimSpace(user.EffectiveName())
	if name == "" {
		name = strings.TrimSpace(user.Username)
	}
	if name == "" {
		name = user.ID.String()
	}
	return truncateRunes(name, maxThreadName)
}

func truncateRunes(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}

// stickerLinks renders the stickers of a message as links. Re-sending stickers
// across guilds is unreliable (the bot must have access to the sticker), so
// they are relayed as URLs instead.
func stickerLinks(stickers []discord.MessageSticker) []string {
	links := make([]string, 0, len(stickers))
	for _, sticker := range stickers {
		url := discord.Sticker{ID: sticker.ID, Name: sticker.Name, FormatType: sticker.FormatType}.URL()
		links = append(links, fmt.Sprintf("Sticker « %s » : %s", sticker.Name, url))
	}
	return links
}

// splitContent cuts content into chunks Discord accepts, preferring line
// boundaries. An empty content yields no chunk at all.
func splitContent(content string) []string {
	content = strings.TrimRight(content, "\n")
	if strings.TrimSpace(content) == "" {
		return nil
	}

	var chunks []string
	remaining := []rune(content)
	for len(remaining) > maxMessageLength {
		cut := maxMessageLength
		for i := maxMessageLength - 1; i > maxMessageLength/2; i-- {
			if remaining[i] == '\n' {
				cut = i + 1
				break
			}
		}
		if chunk := strings.TrimRight(string(remaining[:cut]), "\n"); chunk != "" {
			chunks = append(chunks, chunk)
		}
		remaining = remaining[cut:]
	}
	if last := strings.TrimRight(string(remaining), "\n"); last != "" {
		chunks = append(chunks, last)
	}
	return chunks
}

// collectAttachments downloads what fits in the upload budget and returns the
// attachments to re-upload plus the (expiring) links of everything skipped.
func collectAttachments(attachments []discord.Attachment) (kept []relayAttachment, skipped []string) {
	total := 0
	for _, attachment := range attachments {
		if len(kept) >= maxRelayedFiles || attachment.Size > maxUploadBytes || total+attachment.Size > maxUploadBytes {
			skipped = append(skipped, fmt.Sprintf(msgAttachmentTooBig, attachment.Filename, attachment.URL))
			continue
		}

		data, err := fetchAttachment(attachment.URL)
		if err != nil {
			logger.Warn("MP threads: downloading attachment failed, relaying its link",
				"file", attachment.Filename, "error", err)
			skipped = append(skipped, fmt.Sprintf(msgAttachmentFailed, attachment.Filename, attachment.URL))
			continue
		}

		description := ""
		if attachment.Description != nil {
			description = *attachment.Description
		}
		total += len(data)
		kept = append(kept, relayAttachment{
			filename:    attachment.Filename,
			description: description,
			url:         attachment.URL,
			data:        data,
		})
	}
	return kept, skipped
}

// fetchAttachment downloads an attachment into memory. The body is fully read
// here because the REST client consumes the reader later, after the response
// would already be closed.
func fetchAttachment(url string) ([]byte, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxUploadBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxUploadBytes {
		return nil, fmt.Errorf("attachment larger than %d bytes", maxUploadBytes)
	}
	return data, nil
}

func appendLines(body *strings.Builder, lines []string) {
	for _, line := range lines {
		if body.Len() > 0 {
			body.WriteString("\n")
		}
		body.WriteString(line)
	}
}
