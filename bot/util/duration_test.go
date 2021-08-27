package util

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var parserTestCases = []struct {
	long   string
	normal string
}{
	{"480ns", "480ns"},
	{"10h12s", "10h12s"},
	{"10h12sec", "10h12s"},
	{"10h12.5sec", "10h12.5s"},
	{"10h1min12sec", "10h1m12s"},
}

func TestParseDuration(t *testing.T) {
	t.Run("ParseDuration", func(t *testing.T) {
		for _, testCase := range parserTestCases {
			native, err := time.ParseDuration(testCase.normal)
			assert.Nil(t, err)

			actualFull, err := ParseDuration(testCase.long)
			assert.Nil(t, err)
			actualShort, err := ParseDuration(testCase.normal)
			assert.Nil(t, err)

			assert.Equal(t, native.String(), actualFull.String())
			assert.Equal(t, native.String(), actualShort.String())
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
	{time.Hour*10 + time.Second*12 + 11*time.Millisecond, "10h0m12.0s"},
	{time.Hour*2 + time.Minute*5 + 25*time.Second, "2h5m25s"},
	{time.Hour * 24, "1d0s"},
	{time.Hour*26 + time.Minute*5 + 25*time.Second, "1d2h5m25s"},
	{time.Hour*1000 + 25*time.Second, "41d16h0m25s"},
	{time.Second*1 + time.Millisecond*100, "1.1s"},
	{time.Second*1 + time.Microsecond*10, "1.0s"},
	{time.Minute*1 + time.Second*1 + time.Millisecond*100, "1m1.1s"},
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
