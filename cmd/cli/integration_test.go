//go:build !windows
// +build !windows

package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"syscall"
	"testing"
	"time"

	"github.com/gookit/color"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/tester"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	input := &util.MutexBuffer{}
	output := &util.MutexBuffer{}

	color.Enable = false
	cfg := config.Config{}

	lock := mocks.LockInternalMessages()
	defer lock.Unlock()

	ctx := util.NewServerContext()

	expectedOutput := &util.MutexBuffer{}
	expectedOutput.Write([]byte("Type in your command:\n"))

	go startCli(ctx, input, output, cfg)
	time.Sleep(time.Millisecond * 100)

	testCommand("reply it works", "it works", input, expectedOutput)
	testCommand("wtf", "❓\nOops! Command wtf not found...try help.\n<"+tester.FakeServerURL+"command?command=help|Help!>\n", input, expectedOutput)
	testCommand("add reaction :smile:", "😄", input, expectedOutput)
	testCommand("add link EXAMPLE https://example.com", "<https://example.com|EXAMPLE>\n", input, expectedOutput)
	testCommand("add button \"text\" \"reply test\"", "<"+tester.FakeServerURL+"command?command=reply test|text>\n", input, expectedOutput)

	// delay
	testCommand("delay 10m reply I'm delayed", "I queued the command reply I'm delayed for 10m0s. Use stop timer 0 to stop the timer\n<"+tester.FakeServerURL+"command?command=stop timer 0|Stop timer!>\n", input, expectedOutput)
	testCommand("stop timer 0", "Stopped timer!", input, expectedOutput)
	testCommand("stop timer 0", "invalid timer", input, expectedOutput)

	// custom commands
	testCommand("add command 'wtf' 'reply bar'", "Added command: reply bar. Just use wtf in future.", input, expectedOutput)
	testCommand("wtf", "executing command: reply bar\nbar", input, expectedOutput)
	testCommand("add command 'timer-test' 'delay 100ms quiet reply foo; delay 1ms quiet reply bar'", "Added command: delay 100ms quiet reply foo; delay 1ms quiet reply bar. Just use timer-test in future.", input, expectedOutput)
	testCommand("timer-test", "executing command: delay 100ms quiet reply foo; delay 1ms quiet reply bar\nbar\nfoo", input, expectedOutput)

	time.Sleep(time.Second * 1)

	testURL := tester.FakeServerURL + "command?command=reply%20X"
	r, err := http.Get(testURL) //nolint:gosec
	assert.Nil(t, err)

	resp, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	assert.Equal(t, "Executed command 'reply X'. You can close the browser and go back to the terminal.", string(resp))
	expectedOutput.Write([]byte("Clicked link with message: reply X\n"))
	expectedOutput.Write([]byte("X\n"))

	time.Sleep(time.Millisecond * 500)

	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	assert.Equal(t, expectedOutput.String(), output.String())
}

func testCommand(command string, expectedOutput string, input io.Writer, output io.Writer) {
	input.Write([]byte(command + "\n"))
	output.Write([]byte(expectedOutput + "\n"))
}
