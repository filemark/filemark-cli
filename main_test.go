package main

import (
	"flag"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	assertions "github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func Test_createApp(t *testing.T) {
	assert := assertions.New(t)

	actual := createApp(http.DefaultClient)
	assert.Equal("filemark", actual.Name)
	assert.Equal("1.0.0", actual.Version)
	assert.Equal("cli for filemark.io", actual.Usage)
	assert.Len(actual.Commands, 2)
}

func Test_commands(t *testing.T) {
	assert := assertions.New(t)

	actual := commands(http.DefaultClient)
	assert.Len(actual, 2)
	assert.Equal("SET", actual[0].Name)
	assert.Equal("filemark SET <file path>", actual[0].Usage)
	assert.NotNil(actual[0].Action)

	assert.Equal("GET", actual[1].Name)
	assert.Equal("filemark GET <file id>", actual[1].Usage)
	assert.NotNil(actual[1].Action)
}

func Test_set(t *testing.T) {
	assert := assertions.New(t)
	server := httptest.NewServer(postRequestHandler())
	defer server.Close()

	t.Run("SET command successfully finish", func(t *testing.T) {
		file, err := ioutil.TempFile("", "*.txt")
		assert.NoError(err)

		_, err = file.Write([]byte(content))
		assert.NoError(err)
		defer os.Remove(file.Name())

		flagSet := flag.NewFlagSet("SET", flag.ContinueOnError)
		assert.NoError(flagSet.Parse([]string{file.Name()}))

		assert.NoError(set(http.DefaultClient, server.URL)(cli.NewContext(nil, flagSet, nil)))
	})

	t.Run("SET command finish with error - empty file path", func(t *testing.T) {
		flagSet := flag.NewFlagSet("SET", flag.ContinueOnError)
		assert.Error(set(http.DefaultClient, server.URL)(cli.NewContext(nil, flagSet, nil)))
	})

	t.Run("SET command finish with error - can't read file", func(t *testing.T) {
		flagSet := flag.NewFlagSet("SET", flag.ContinueOnError)
		assert.NoError(flagSet.Parse([]string{"test.txt"}))

		assert.Error(set(http.DefaultClient, server.URL)(cli.NewContext(nil, flagSet, nil)))
	})

	t.Run("SET command finish with error - can't call HOST", func(t *testing.T) {
		file, err := ioutil.TempFile("", "*.txt")
		assert.NoError(err)

		_, err = file.Write([]byte(content))
		assert.NoError(err)
		defer os.Remove(file.Name())

		flagSet := flag.NewFlagSet("SET", flag.ContinueOnError)
		assert.NoError(flagSet.Parse([]string{file.Name()}))

		assert.Error(set(http.DefaultClient, " ")(cli.NewContext(nil, flagSet, nil)))
	})
}

func Test_get(t *testing.T) {
	assert := assertions.New(t)
	server := httptest.NewServer(getRequestHandler())
	fileID := "h45h"
	defer server.Close()

	t.Run("GET command successfully finished", func(t *testing.T) {
		flagSet := flag.NewFlagSet("GET", flag.ContinueOnError)
		assert.NoError(flagSet.Parse([]string{fileID}))

		assert.Nil(get(http.DefaultClient, server.URL)(cli.NewContext(nil, flagSet, nil)))
		os.Remove(testFileName)
	})

	t.Run("GET command finish with error - empty file id", func(t *testing.T) {
		flagSet := flag.NewFlagSet("SET", flag.ContinueOnError)

		assert.Error(get(http.DefaultClient, server.URL)(cli.NewContext(nil, flagSet, nil)))
	})

	t.Run("GET command finish with error - can't call HOST", func(t *testing.T) {
		flagSet := flag.NewFlagSet("SET", flag.ContinueOnError)
		assert.NoError(flagSet.Parse([]string{fileID}))
		assert.Error(get(http.DefaultClient, "")(cli.NewContext(nil, flagSet, nil)))
	})
}
