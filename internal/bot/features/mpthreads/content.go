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

	maxUploadBytes = 9 << 20
)

const downloadTimeout = 30 * time.Second

var httpClient = &http.Client{Timeout: downloadTimeout}

const (
	msgAttachmentTooBig  = "Pièce jointe « %s » trop lourde pour être ré-uploadée, lien temporaire (expire sous ~24 h) : %s"
	msgAttachmentFailed  = "Pièce jointe « %s » indisponible au téléchargement, lien temporaire (expire sous ~24 h) : %s"
	msgAttachmentRefused = "Pièce jointe « %s » refusée par Discord, lien temporaire (expire sous ~24 h) : %s"
)

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

func noMentions() *discord.AllowedMentions {
	return &discord.AllowedMentions{
		Parse: []discord.AllowedMentionType{},
		Roles: []snowflake.ID{},
		Users: []snowflake.ID{},
	}
}

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

func stickerLinks(stickers []discord.MessageSticker) []string {
	links := make([]string, 0, len(stickers))
	for _, sticker := range stickers {
		url := discord.Sticker{ID: sticker.ID, Name: sticker.Name, FormatType: sticker.FormatType}.URL()
		links = append(links, fmt.Sprintf("Sticker « %s » : %s", sticker.Name, url))
	}
	return links
}

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
