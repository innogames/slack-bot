package main

import (
	"bytes"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAll(t *testing.T) {
	input := &bytes.Buffer{}
	output := &bytes.Buffer{}
	kill := make(chan os.Signal, 1)

	go startCli(input, output, kill)

	input.Write([]byte("reply it works\n"))
	time.Sleep(time.Millisecond * 100)

	assert.Equal(t, output.String(), ">>>> reply it works\n<<<< it works\n")

	kill <- syscall.Signal(1)
}
