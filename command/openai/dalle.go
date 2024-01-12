package openai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/stats"
	"github.com/innogames/slack-bot/v2/bot/util"
	log "github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
)

// bot function to generate images with Dall-E
func (c *openaiCommand) dalleGenerateImage(match matcher.Result, message msg.Message) {
	// run the generation in the background and as it could take some time
	c.AddReaction(":coffee:", message)

	go func() {
		prompt := match.GetString(util.FullMatch)
		images, err := generateImages(c.cfg, prompt)
		c.RemoveReaction(":coffee:", message)
		if err != nil {
			c.ReplyError(message, err)
			return
		}

		// add 📤 emoji to indicate that the image is being uploaded which can take some time via slack
		c.AddReaction(":outbox_tray:", message)
		defer c.RemoveReaction(":outbox_tray:", message)

		startTime := time.Now()
		for _, image := range images {
			err := c.sendImageInSlack(image, message.MessageRef)
			if err != nil {
				c.ReplyError(
					message,
					fmt.Errorf("failed to download image: %s %s", image.URL, err),
				)
			}
		}
		log.Infof("Uploading %d images took %s", len(images), util.FormatDuration(time.Since(startTime)))
	}()
}

func (c *openaiCommand) sendImageInSlack(image DalleResponseImage, message msg.Ref) error {
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
		Channels:        []string{message.GetChannel()},
		ThreadTimestamp: message.GetTimestamp(),
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
	resp, err := doRequest(cfg, "POST", apiDalleGenerateImageURL, jsonData)
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
		Infof("Dall-E image generation took %s", util.FormatDuration(time.Since(start)))
	stats.Increase("openai_dalle_images", len(response.Data))

	return response.Data, nil
}
