package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const APIEndpoint = "https://api.remove.bg/v1.0/removebg"

//go:generate counterfeiter . ClientInterface
type ClientInterface interface {
	RemoveFromFile(inputPath string, apiKey string, params map[string]string) ([]byte, error)
}

type Client struct {
	HTTPClient http.Client
}

func (c Client) RemoveFromFile(inputPath string, apiKey string, params map[string]string) ([]byte, error) {
	request, err := buildRequest(APIEndpoint, apiKey, params, inputPath)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.HTTPClient.Do(request)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println(resp.StatusCode)
		fmt.Println(resp.Header)
		return nil, errors.New("unable to process image")
	}

	return ioutil.ReadAll(resp.Body)
}

func buildRequest(uri string, apiKey string, params map[string]string, inputPath string) (*http.Request, error) {
	file, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("image_file", filepath.Base(inputPath))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Add("X-Api-Key", apiKey)

	return req, err
}