package openai

import (
	"strings"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/slack-go/slack"
)

// messageChunk holds a reference to a slack message and tracks its content
type messageChunk struct {
	messageRef string
	content    strings.Builder
}

// messageChunker manages splitting long responses into multiple Slack messages
type messageChunker struct {
	chunks       []messageChunk
	chunkSizeMax int
	slackClient  bot.BaseCommand
	message      msg.Ref
	noThread     bool // if true, don't use thread timestamps
}

// newMessageChunker creates a new message chunker with the first message initialized
func newMessageChunker(slackClient bot.BaseCommand, message msg.Ref, initialRef string, chunkSize int, noThread bool) *messageChunker {
	return &messageChunker{
		chunks: []messageChunk{
			{
				messageRef: initialRef,
				content:    strings.Builder{},
			},
		},
		chunkSizeMax: chunkSize,
		slackClient:  slackClient,
		message:      message,
		noThread:     noThread,
	}
}

// appendContent adds content and creates new chunks as needed
func (mc *messageChunker) appendContent(delta string) {
	remainingDelta := delta

	for len(remainingDelta) > 0 {
		currentChunk := &mc.chunks[len(mc.chunks)-1]
		availableSpace := mc.chunkSizeMax - currentChunk.content.Len()

		// If current chunk is full or delta won't fit, create a new chunk
		if availableSpace <= 0 {
			msgOptions := []slack.MsgOption{}
			if !mc.noThread || mc.message.GetThread() != "" {
				msgOptions = append(msgOptions, slack.MsgOptionTS(mc.message.GetTimestamp()))
			}
			newRef := mc.slackClient.SendMessage(
				mc.message,
				":bulb: thinking...",
				msgOptions...,
			)
			mc.chunks = append(mc.chunks, messageChunk{
				messageRef: newRef,
				content:    strings.Builder{},
			})
			currentChunk = &mc.chunks[len(mc.chunks)-1]
			availableSpace = mc.chunkSizeMax
		}

		// Write as much as we can to the current chunk
		if len(remainingDelta) <= availableSpace {
			// Delta fits entirely in current chunk
			currentChunk.content.WriteString(remainingDelta)
			remainingDelta = ""
		} else {
			// Delta is larger than available space, split it
			currentChunk.content.WriteString(remainingDelta[:availableSpace])
			remainingDelta = remainingDelta[availableSpace:]
		}
	}
}

// updateMessages updates all chunks in Slack
func (mc *messageChunker) updateMessages() {
	for i := range mc.chunks {
		chunk := &mc.chunks[i]
		if chunk.content.Len() > 0 {
			msgOptions := []slack.MsgOption{slack.MsgOptionUpdate(chunk.messageRef)}
			if !mc.noThread || mc.message.GetThread() != "" {
				msgOptions = append(msgOptions, slack.MsgOptionTS(mc.message.GetTimestamp()))
			}
			mc.slackClient.SendMessage(
				mc.message,
				chunk.content.String(),
				msgOptions...,
			)
		}
	}
}

// getFullText returns the complete text across all chunks
func (mc *messageChunker) getFullText() string {
	var fullText strings.Builder
	for i := range mc.chunks {
		fullText.WriteString(mc.chunks[i].content.String())
	}
	return fullText.String()
}

// getTotalLength returns the total length of all chunks
func (mc *messageChunker) getTotalLength() int {
	total := 0
	for i := range mc.chunks {
		total += mc.chunks[i].content.Len()
	}
	return total
}
