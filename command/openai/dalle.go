package openai

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	log "github.com/sirupsen/logrus"
)

// bot function to generate images with Dall-E
func (c *chatGPTCommand) dalleGenerateImage(match matcher.Result, message msg.Message) {
	prompt := match.GetString(util.FullMatch)

	go func() {
		c.AddReaction(":coffee:", message)
		defer c.RemoveReaction(":coffee:", message)

		images, err := generateImage(c.cfg, prompt)
		if err != nil {
			c.ReplyError(message, err)
			return
		}

		text := ""
		for _, image := range images {
			text += fmt.Sprintf(
				" - %s: <%s|open image>\n",
				image.RevisedPrompt,
				image.URL,
			)
		}
		c.SendMessage(message, text)
	}()
}

func generateImage(cfg Config, prompt string) ([]DalleResponseImage, error) {
	jsonData, _ := json.Marshal(DalleRequest{
		Model:  cfg.DalleModel,
		Size:   cfg.DalleImageSize,
		N:      cfg.DalleNumberOfImages,
		Prompt: prompt,
	})

	start := time.Now()
	resp, err := doRequest(cfg, apiDalleGenerateImageURL, jsonData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response DalleResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}

	if response.Error.Message != "" {
		return nil, fmt.Errorf(response.Error.Message)
	}

	log.WithField("model", cfg.DalleModel).
		Infof("Dall-E image generation took %s", time.Since(start))

	return response.Data, nil
}
