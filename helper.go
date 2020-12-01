package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"path/filepath"
)

const (
	filesPath         = "/api/v1/files"
	filemarkBase      = "https://filemark.io"
	multipartFileName = "file"
	fileNameHeader    = "File-Name"
)

type FileUploadResponse struct {
	ID string `json:"id"`
}

func writeMultipart(filePath string) (*bytes.Buffer, string, error) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, "", err
	}

	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	defer w.Close()

	part, err := w.CreateFormFile(multipartFileName, filepath.Base(filePath))
	if err != nil {
		return nil, "", err
	}

	if _, err = part.Write(file); err != nil {
		return nil, "", err
	}

	return buf, w.FormDataContentType(), nil
}

func doPostRequest(ctx context.Context, client *http.Client, url string, buf io.Reader, contentType string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func doGetRequest(ctx context.Context, client *http.Client, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func unmarshalResponse(body io.Reader, dest interface{}) error {
	buf, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(buf, dest); err != nil {
		return err
	}
	return nil
}

func buildDownloadURL(base, id string) string {
	return base + filesPath + "/" + id
}

func buildUploadURL(base string) string {
	return base + filesPath
}
