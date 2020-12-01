package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var errGiveMeAnotherTry = errors.New("give me another try")

func main() {
	if err := createApp(&http.Client{}).Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func createApp(client *http.Client) *cli.App {
	return &cli.App{Name: "filemark", Usage: "cli for filemark.io", Version: "1.0.0", Commands: commands(client)}
}

func commands(client *http.Client) []*cli.Command {
	return []*cli.Command{
		{Name: "SET", Usage: "filemark SET <file path>", Action: set(client, filemarkBase)},
		{Name: "GET", Usage: "filemark GET <file id>", Action: get(client, filemarkBase)},
	}
}

func set(client *http.Client, basePath string) cli.ActionFunc {
	return func(c *cli.Context) error {
		filePath := c.Args().First()
		if filePath == "" {
			return errors.New("<file path> is empty")
		}

		buf, contentType, err := writeMultipart(filePath)
		if err != nil {
			return err
		}

		res, err := doPostRequest(c.Context, client, buildUploadURL(basePath), buf, contentType)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusCreated {
			return errGiveMeAnotherTry
		}

		var resDto FileUploadResponse
		if err := unmarshalResponse(res.Body, &resDto); err != nil {
			return err
		}

		fmt.Println(resDto.ID)
		return nil
	}
}

func get(client *http.Client, basePath string) cli.ActionFunc {
	return func(c *cli.Context) error {
		fileID := c.Args().First()
		if fileID == "" {
			return errors.New("<file id> is empty")
		}

		res, err := doGetRequest(c.Context, client, buildDownloadURL(basePath, fileID))
		if err != nil {
			return err
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			return errGiveMeAnotherTry
		}

		fileName, err := saveFile(res)
		if err != nil {
			return err
		}

		fmt.Println(fileName)
		return nil
	}
}

func saveFile(res *http.Response) (string, error) {
	out, err := os.Create(res.Header.Get(fileNameHeader))
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err = io.Copy(out, res.Body); err != nil {
		return "", err
	}
	return out.Name(), nil
}
