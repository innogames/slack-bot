package main

import (
	"testing"
	"time"

	"github.com/gookit/color"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	input := &util.MutexBuffer{}
	output := &util.MutexBuffer{}

	color.Enable = false
	cfg := config.Config{}

	ctx := util.NewServerContext()

	expectedOutput := &util.MutexBuffer{}
	expectedOutput.Write([]byte("Type in your command:\n"))

	go startCli(ctx, input, output, cfg)
	time.Sleep(time.Millisecond * 200)

	testCommand("reply it works", "it works", input, expectedOutput)
	testCommand("wtf", "Oops! Command `wtf` not found...try `help`.", input, expectedOutput)
	testCommand("add reaction :smile:", "😄", input, expectedOutput)

	// custom commands
	testCommand("add command 'wtf' 'reply bar'", "Added command: `reply bar`. Just use `wtf` in future.", input, expectedOutput)
	testCommand("wtf", "executing command: `reply bar`\nbar", input, expectedOutput)

	ctx.StopTheWorld()

	assert.Equal(t, output.String(), expectedOutput.String())
}

func testCommand(command string, expectedOutput string, input *util.MutexBuffer, output *util.MutexBuffer) {
	input.Write([]byte(command + "\n"))
	time.Sleep(time.Millisecond * 200)

	output.Write([]byte(">>>> " + command + "\n"))
	output.Write([]byte(expectedOutput + "\n"))
}
