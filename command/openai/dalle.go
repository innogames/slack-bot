package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/innogames/slack-bot/v2/bot/matcher"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/util"
)

/*
	{
	    "model": "dall-e-3",
	    "prompt": "a white siamese cat",
	    "n": 1,
	    "size": "1024x1024"
	  }
*/
type DalleRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	N      int    `json:"n"`
	Size   string `json:"size"`
}

/*
	{
	  "created": 1700233554,
	  "data": [
	    {
	      "url": "https://XXXX"
	    }
	  ]
	}
or:
	{
	  "error": {
	    "code": "invalid_size",
	    "message": "The size is not supported by this model.",
	    "param": null,
	    "type": "invalid_request_error"
	  }
	}
*/

type DalleResponseImage struct {
	URL           string `json:"url"`
	RevisedPrompt string `json:"revised_prompt"`
}

type DalleResponse struct {
	Data  []DalleResponseImage
	Error []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
}

// bot function which is called, when the user started a new conversation with openai/chatgpt
func (c *chatGPTCommand) dalle(match matcher.Result, message msg.Message) {
	prompt := match.GetString(util.FullMatch)

	c.AddReaction(":coffee:", message)
	defer c.RemoveReaction(":coffee:", message)

	image, err := generateImage(c.cfg, prompt)
	if err != nil {
		c.ReplyError(message, err)
		return
	}

	c.SendMessage(message,
		fmt.Sprintf(
			"%s: %s",
			image.RevisedPrompt,
			image.URL,
		),
	)
}

// WIP, just a simple test
func generateImage(cfg Config, prompt string) (*DalleResponseImage, error) {
	jsonData, _ := json.Marshal(DalleRequest{
		Model:  "dall-e-3",
		Prompt: prompt,
		N:      1,
		Size:   "1024x1024",
	})
	fmt.Println(string(jsonData))

	req, err := http.NewRequest("POST", cfg.APIHost+apiDalleURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	r, _ := io.ReadAll(resp.Body)
	var response DalleResponse
	_ = json.Unmarshal(r, &response)

	if len(response.Error) > 0 {
		return nil, fmt.Errorf(response.Error[0].Message)
	}

	fmt.Println(string(r))
	fmt.Println(response)

	return &response.Data[0], nil
}
