package main

import (
	"bytes"
	"context"
	"github.com/gookit/color"
	"github.com/innogames/slack-bot/bot/config"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestAll(t *testing.T) {
	input := &bytes.Buffer{}
	output := &bytes.Buffer{}

	color.Enable = false
	cfg := config.Config{}

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	input.Write([]byte("reply it works\n"))

	go startCli(ctx, wg, input, output, cfg, false)
	time.Sleep(time.Millisecond * 200)

	cancel()
	wg.Wait()

	assert.Equal(t, output.String(), "Type in your command:\n>>>> reply it works\nit works\n")
}
