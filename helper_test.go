package main

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/magiconair/properties/assert"
	assertions "github.com/stretchr/testify/assert"
)

var (
	content      = "some content"
	testFileName = "woodle-doodle-noodle.png"
)

func Test_unmarshalResponse(t *testing.T) {
	assert := assertions.New(t)

	t.Run("body unmarshalled successfully", func(t *testing.T) {
		body := []byte(`{"id":"h45h"}`)

		var resDto FileUploadResponse
		assert.NoError(unmarshalResponse(ioutil.NopCloser(bytes.NewReader(body)), &resDto))
		assert.Equal("h45h", resDto.ID)
	})

	t.Run("failed to unmarshal body", func(t *testing.T) {
		assert.Error(unmarshalResponse(ioutil.NopCloser(bytes.NewReader(nil)), nil))
	})
}

func Test_writeMultipart(t *testing.T) {
	assert := assertions.New(t)

	t.Run("multipart write successfully", func(t *testing.T) {
		file, err := ioutil.TempFile("", "*.txt")
		assert.NoError(err)

		_, err = file.Write([]byte(content))
		assert.NoError(err)
		defer os.Remove(file.Name())

		buf, header, err := writeMultipart(file.Name())
		assert.NoError(err)
		assert.Contains(header, "multipart/form-data")
		assert.Contains(buf.String(), content)
	})

	t.Run("failed to write multipart - file not found", func(t *testing.T) {
		buf, header, err := writeMultipart("file_not_exists.txt")
		assert.Error(err)
		assert.Equal("", header)
		assert.Nil(buf)
	})
}

func Test_doPostRequest(t *testing.T) {
	assert := assertions.New(t)
	ctx := context.Background()
	contentType := "multipart/form-data"
	server := httptest.NewServer(postRequestHandler())
	defer server.Close()

	t.Run("post request was successful", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte(content))

		res, err := doPostRequest(ctx, http.DefaultClient, server.URL, buf, contentType)
		assert.NoError(err)
		defer res.Body.Close()

		bt, err := ioutil.ReadAll(res.Body)
		assert.NoError(err)
		assert.Contains(string(bt), "h45h")
	})

	t.Run("failed to do post request - can't build request", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte(content))

		res, err := doPostRequest(ctx, http.DefaultClient, " "+filemarkBase, buf, contentType)
		assert.Error(err)
		assert.Nil(res)
	})

	t.Run("failed to do post request - address not exists", func(t *testing.T) {
		buf := bytes.NewBuffer([]byte(content))

		res, err := doPostRequest(ctx, http.DefaultClient, "https://i-hope-this-url-not-exists.com", buf, contentType)
		assert.Error(err)
		assert.Nil(res)
	})
}

func Test_doGetRequest(t *testing.T) {
	assert := assertions.New(t)
	ctx := context.Background()
	server := httptest.NewServer(getRequestHandler())
	defer server.Close()

	t.Run("get request was successful", func(t *testing.T) {
		res, err := doGetRequest(ctx, http.DefaultClient, server.URL)
		assert.NoError(err)
		assert.Equal(testFileName, res.Header.Get(fileNameHeader))

		bt, err := ioutil.ReadAll(res.Body)
		assert.NoError(err)
		assert.Equal([]byte(content), bt)
	})

	t.Run("failed to do get request - can't build request", func(t *testing.T) {
		res, err := doGetRequest(ctx, http.DefaultClient, " "+filemarkBase)
		assert.Error(err)
		assert.Nil(res)
	})

	t.Run("failed to do get request - address not exists", func(t *testing.T) {
		res, err := doGetRequest(ctx, http.DefaultClient, "https://i-hope-this-url-not-exists.com")
		assert.Error(err)
		assert.Nil(res)
	})
}

func postRequestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, `{"id":"h45h"}`)
	}
}

func getRequestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(fileNameHeader, testFileName)
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, content)
	}
}

func Test_buildDownloadURL(t *testing.T) {
	assert.Equal(t, "https://filemark.io/api/v1/files/h45h", buildDownloadURL(filemarkBase, "h45h"))
}

func Test_buildUploadURL(t *testing.T) {
	assert.Equal(t, "https://filemark.io/api/v1/files", buildUploadURL(filemarkBase))
}
