package openai

import (
	"encoding/json"
	"fmt"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/storage"
	"github.com/slack-go/slack"
	"io"
	"time"
)

// see https://platform.openai.com/docs/assistants/how-it-works

type assistantThreadResponse struct {
	Id string `json:"id"`
}
type assistantStartRun struct {
	AssistantId string `json:"assistant_id"`
}

type run struct {
	Id             string                  `json:"id"`
	Status         string                  `json:"status"`
	ThreadId       string                  `json:"thread_id"`
	RequiredAction AssistantRequiredAction `json:"required_action"`
}

type AssistantRequiredAction struct {
	Type               string `json:"type"`
	SubmitToolsOutputs struct {
		ToolCalls []struct {
			Id       string `json:"id"`
			Type     string `json:"type"`
			Function struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			} `json:"function"`
		} `json:"tool_calls"`
	} `json:"submit_tool_outputs"`
}

type AssistantContent struct {
	Type string `json:"type"`
	Text struct {
		Value string `json:"value"`
	} `json:"text"`
}

func (c AssistantContent) GetText() string {
	return c.Text.Value
}

type AssistantStartThreads struct {
	Messages []ChatMessage `json:"messages"`
}

type AssistantChatMessage struct {
	Id          string             `json:"id"`
	Role        string             `json:"role"`
	ChatMessage []AssistantContent `json:"content"`
	RunId       string             `json:"run_id"`
}
type assistantFullResponse struct {
	Data []AssistantChatMessage `json:"data"`
}

type AssistantToolsOutput struct {
	ToolsOutput []struct {
		ToolCallId string `json:"tool_call_id"`
		Output     string `json:"output"`
	} `json:"tool_outputs"`
}

func (c *openaiCommand) callCustomGPT(messages []ChatMessage, identifier string, message msg.Ref, text string) {
	c.AddReaction(":coffee:", message)
	defer c.RemoveReaction(":coffee:", message)

	messages = append(messages, ChatMessage{
		Role:    roleUser,
		Content: text,
	})

	var threadId string
	var err error
	storage.Read("gpt-thread", identifier, &threadId)
	if threadId == "" {
		// start a new thread!
		threadId, err = createAssistantThread(c.cfg, messages)
		if err != nil {
			c.ReplyError(message, err)
			return
		}
		storage.Write("gpt-thread", identifier, threadId)
	} else {
		// attach slack messages to an existing thread
		for _, newMessage := range messages {
			// todo no API to bulk add?!
			addMessage(c.cfg, threadId, newMessage)
		}
	}

	// start the assistant and get a "run" object
	run, err := assistantRun(c.cfg, threadId)
	if err != nil {
		c.ReplyError(message, err)
		return
	}

	// wait till run is done or required more information from function calls!
	// see https://platform.openai.com/docs/assistants/how-it-works/run-lifecycle
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	for range ticker.C {
		run, err = getRun(c.cfg, run)
		if err != nil || run.Status == "failed" || run.Status == "cancelled" || run.Status == "expired" {
			c.ReplyError(message, fmt.Errorf("run failed with status %s", run.Status))
			return
		}

		if run.Status == "completed" {
			// we have the final answer!
			break
		}

		if run.Status == "requires_action" {
			// todo extract code!
			fmt.Println(run.RequiredAction)
			fmt.Println(run.RequiredAction.SubmitToolsOutputs)
			tool := run.RequiredAction.SubmitToolsOutputs.ToolCalls[0]

			var output string
			if tool.Function.Name == "dall_image" {
				// special function
				prompt := tool.Function.Arguments
				fmt.Println(prompt, "prompt")

				images, _ := generateImages(c.cfg, prompt)
				output = images[0].RevisedPrompt
				go c.sendImageInSlack(images[0], message)
			} else {
				output = "Ticket: Fix issue in feature XYZ, status = open" // todo call function
			}

			sendToolsOutput(c.cfg, run, tool.Id, output)

			// wait for new tick, as the API is handling the new information now...
			continue
		}
	}

	// todo only fetch the new messages for this run
	respMessages, _ := listMessages(c.cfg, threadId)
	for _, m := range respMessages {
		if m.RunId != run.Id {
			continue
		}
		fmt.Println(m.ChatMessage)
		if m.Role != roleAssistant {
			continue
		}

		// reply in thread
		c.SendMessage(
			message,
			m.ChatMessage[0].GetText(),
			slack.MsgOptionTS(message.GetTimestamp()),
		)
	}
}

/*
func (c *openaiCommand) assistantUploadFile(cfg Config, file slack.File) error {
	var buf bytes.Buffer
	log.Infof("Downloading message attachment file %s", file.Name)

	fmt.Println(file)

	resp, err := doRequest(cfg, "POST", apiFilesURL, []byte("jolo"))
	if err != nil {
		return nil
	}

	r, _ := io.ReadAll(resp.Body)
	fmt.Println(string(r))

	return nil
}
*/

func assistantRun(cfg Config, threadId string) (*run, error) {
	fmt.Printf("run assistant %s\n", threadId)

	assistantStartRun := assistantStartRun{
		AssistantId: cfg.CustomGPT,
	}

	req, _ := json.Marshal(assistantStartRun)
	resp, err := doRequest(cfg, "POST", apiThreadsURL+"/"+threadId+"/runs", req)
	if err != nil {
		return nil, err
	}

	run := &run{}
	err = json.NewDecoder(resp.Body).Decode(run)
	return run, err
}

func addMessage(cfg Config, threadId string, message ChatMessage) error {
	fmt.Printf("add message to thread %s: %s\n", threadId, message)

	req, _ := json.Marshal(message)
	_, err := doRequest(cfg, "POST", apiThreadsURL+"/"+threadId+"/messages", req)

	return err
}

func createAssistantThread(cfg Config, messages []ChatMessage) (string, error) {
	fmt.Println("create thread")

	req, _ := json.Marshal(AssistantStartThreads{
		Messages: messages,
	})
	fmt.Println(string(req))
	resp, err := doRequest(cfg, "POST", apiThreadsURL, req)
	if err != nil {
		return "", err
	}
	//r, _ := io.ReadAll(resp.Body)
	//fmt.Println(string(r))
	thread := assistantThreadResponse{}
	err = json.NewDecoder(resp.Body).Decode(&thread)
	if err != nil {
		return "", err
	}
	fmt.Println(thread)

	if thread.Id == "" {
		return "", fmt.Errorf("failed to create thread")
	}
	return thread.Id, nil
}

func getRun(cfg Config, oldRun *run) (*run, error) {
	fmt.Printf("get run %s %s\n", oldRun.ThreadId, oldRun.Id)
	resp, err := doRequest(cfg, "GET", apiThreadsURL+"/"+oldRun.ThreadId+"/runs/"+oldRun.Id, nil)
	if err != nil {
		return oldRun, err
	}

	r, _ := io.ReadAll(resp.Body)
	fmt.Println(string(r))

	newRun := &run{}
	err = json.Unmarshal(r, newRun)

	return newRun, err
}

func listMessages(cfg Config, threadId string) ([]AssistantChatMessage, error) {
	fmt.Printf("list messages %s \n", threadId)
	resp, err := doRequest(cfg, "GET", apiThreadsURL+"/"+threadId+"/messages", nil)
	if err != nil {
		return []AssistantChatMessage{}, err
	}

	r, _ := io.ReadAll(resp.Body)
	fmt.Println(string(r))

	messages := assistantFullResponse{}
	json.Unmarshal(r, &messages)

	return messages.Data, nil
}

func sendToolsOutput(cfg Config, run *run, callId string, output string) error {
	fmt.Printf("send tools output %s %s %s\n", run.ThreadId, run.Id, callId)

	req, _ := json.Marshal(AssistantToolsOutput{
		ToolsOutput: []struct {
			ToolCallId string `json:"tool_call_id"`
			Output     string `json:"output"`
		}{
			{
				ToolCallId: callId,
				Output:     output,
			},
		},
	})
	fmt.Println(string(req))
	resp, err := doRequest(cfg, "POST", apiThreadsURL+"/"+run.ThreadId+"/runs/"+run.Id+"/submit_tool_outputs", req)
	if err != nil {
		return err
	}

	r, _ := io.ReadAll(resp.Body)
	fmt.Println(string(r))

	return err
}
