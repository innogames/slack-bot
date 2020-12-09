package main

import (
	"bytes"
	"github.com/gookit/color"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/innogames/slack-bot/bot/util"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAll(t *testing.T) {
	input := &bytes.Buffer{}
	output := &bytes.Buffer{}

	color.Enable = false
	cfg := config.Config{}

	ctx := util.NewServerContext()

	input.Write([]byte("reply it works\n"))

	go startCli(ctx, input, output, cfg)
	time.Sleep(time.Millisecond * 200)

	ctx.StopTheWorld()

	assert.Equal(t, output.String(), "Type in your command:\n>>>> reply it works\nit works\n")
}
