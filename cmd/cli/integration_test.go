// +build !windows

package main

import (
	"io"
	"syscall"
	"testing"
	"time"

	"github.com/gookit/color"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/util"
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

	testCommand("reply it works", "it works", input, expectedOutput)
	testCommand("wtf", "❓\nOops! Command `wtf` not found...try `help`.", input, expectedOutput)
	testCommand("add reaction :smile:", "😄", input, expectedOutput)

	// delay
	testCommand("delay 10m reply I'm delayed", "I queued the command `reply I'm delayed` for 10m0s. Use `stop timer 0` to stop the timer", input, expectedOutput)
	testCommand("stop timer 0", "Stopped timer!", input, expectedOutput)
	testCommand("stop timer 0", "invalid timer", input, expectedOutput)

	// custom commands
	testCommand("add command 'wtf' 'reply bar'", "Added command: `reply bar`. Just use `wtf` in future.", input, expectedOutput)
	testCommand("wtf", "executing command: `reply bar`\nbar", input, expectedOutput)

	time.Sleep(time.Second * 2)

	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	assert.Equal(t, expectedOutput.String(), output.String())
}

func testCommand(command string, expectedOutput string, input io.Writer, output io.Writer) {
	input.Write([]byte(command + "\n"))

	output.Write([]byte(">>>> " + command + "\n"))
	output.Write([]byte(expectedOutput + "\n"))
}
