package util

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var parserTestCases = []struct {
	long   string
	normal string
}{
	{"10h12s", "10h12s"},
	{"10h12sec", "10h12s"},
	{"10h12.5sec", "10h12.5s"},
	{"10h1min12sec", "10h1m12s"},
}

func TestParseDuration(t *testing.T) {
	t.Run("ParseDuration", func(t *testing.T) {
		for _, testCase := range parserTestCases {
			native, _ := time.ParseDuration(testCase.normal)

			actualFull, err := ParseDuration(testCase.normal)
			assert.Nil(t, err)
			actualShort, _ := ParseDuration(testCase.normal)
			assert.Nil(t, err)

			assert.Equal(t, native, actualFull)
			assert.Equal(t, native, actualShort)
		}
	})
}

var formatterTestCases = []struct {
	input    time.Duration
	expected string
}{
	{time.Hour * 0, "0s"},
	{time.Hour * 10, "10h0m0s"},
	{time.Hour*10 + time.Second*12, "10h0m12s"},
	{time.Hour*10 + time.Second*12 + 11*time.Millisecond, "10h0m12s"},
}

func TestFormatDuration(t *testing.T) {
	t.Run("FormatDuration", func(t *testing.T) {
		for _, testCase := range formatterTestCases {
			actualFull := FormatDuration(testCase.input)

			assert.Equal(t, testCase.expected, actualFull)
		}
	})
}

func ExampleParseDuration() {
	fmt.Println(ParseDuration("12min1sec"))
	// Output: 12m1s <nil>
}
