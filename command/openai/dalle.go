package openai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

// bot function to generate images with Dall-E
func (c *openaiCommand) dalleGenerateImage(match matcher.Result, message msg.Message) {
	// run the generation in the background and as it could take some time
	c.AddReaction(":coffee:", message)

	go func() {
		defer c.RemoveReaction(":coffee:", message)

		prompt := match.GetString(util.FullMatch)
		images, err := generateImages(c.cfg, prompt)
		if err != nil {
			c.ReplyError(message, err)
			return
		}

		for _, image := range images {
			err := c.sendImageInSlack(image, message)
			if err != nil {
				log.Warnf("failed to download image: %s - %s", err, image.URL)
			}
		}
	}()
}

func (c *openaiCommand) sendImageInSlack(image DalleResponseImage, message msg.Message) error {
	req, err := http.NewRequest("GET", image.URL, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = c.SlackClient.UploadFile(slack.FileUploadParameters{
		Filename:        "dalle.png",
		Filetype:        "png",
		Channels:        []string{message.Channel},
		ThreadTimestamp: message.Timestamp,
		Reader:          resp.Body,
		InitialComment:  image.RevisedPrompt,
	})

	return err
}

func generateImages(cfg Config, prompt string) ([]DalleResponseImage, error) {
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
