package main

import (
	"github.com/gookit/color"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
	"time"
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

	testCommand("reply it works", "it works", input, expectedOutput)
	testCommand("wtf", "Oops! Command `wtf` not found...try `help`.", input, expectedOutput)
	testCommand("add reaction :smile:", "😄", input, expectedOutput)

	// delay
	testCommand("delay 10m reply I'm delayed", "I queued the command `reply I'm delayed` for 10m0s. Use `stop timer 0` to stop the timer", input, expectedOutput)
	testCommand("stop timer 0", "Stopped timer!", input, expectedOutput)
	testCommand("stop timer 0", "invalid timer", input, expectedOutput)

	// custom commands
	testCommand("add command 'wtf' 'reply bar'", "Added command: `reply bar`. Just use `wtf` in future.", input, expectedOutput)
	testCommand("wtf", "executing command: `reply bar`\nbar", input, expectedOutput)

	time.Sleep(time.Second * 2)

	ctx.StopTheWorld()

	assert.Equal(t, expectedOutput.String(), output.String())
}

func testCommand(command string, expectedOutput string, input io.Writer, output io.Writer) {
	input.Write([]byte(command + "\n"))

	output.Write([]byte(">>>> " + command + "\n"))
	output.Write([]byte(expectedOutput + "\n"))
}
