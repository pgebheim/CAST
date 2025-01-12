package shared

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

const (
	baseUrl = "https://api.pinata.cloud"
)

type IpfsClient struct {
	BaseURL    string
	apiKey     string
	apiSecret  string
	HTTPClient *http.Client
}

type ipfsErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ipfsSuccessResponse struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

type Pin struct {
	IpfsHash    string    `json:"IpfsHash"`
	PinSize     int       `json:"PinSize"`
	Timestamp   time.Time `json:"Timestamp"`
	IsDuplicate bool      `json:"isDuplicate"`
}

func NewIpfsClient(apiKey string, apiSecret string) *IpfsClient {
	return &IpfsClient{
		BaseURL:   baseUrl,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		HTTPClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (c *IpfsClient) sendRequest(req *http.Request, v interface{}) error {
	req.Header.Set("pinata_api_key", "c63e755ac2650e32aeb4")
	req.Header.Set("pinata_secret_api_key", "12e61330988a819b5771416e8698a6b9e2550b5be99fad79b9213192b7e97073")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		var errRes ipfsErrorResponse
		if err = json.NewDecoder(res.Body).Decode(&errRes); err == nil {
			return errors.New(errRes.Message)
		}
		return fmt.Errorf("unknown error, status code: %d", res.StatusCode)
	}

	fullResponse := ipfsSuccessResponse{
		Data: v,
	}

	if err = json.NewDecoder(res.Body).Decode(&fullResponse.Data); err != nil {
		return err
	}

	return nil
}

func (c *IpfsClient) PinJson(data interface{}) (*Pin, error) {
	url := c.BaseURL + "/pinning/pinJSONToIPFS"
	json_data, err := json.Marshal(data)

	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(json_data))
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	res := Pin{}

	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *IpfsClient) PinFile(file multipart.File, fileName string) (*Pin, error) {
	url := c.BaseURL + "/pinning/pinFileToIPFS"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", fileName)
	io.Copy(part, file)
	writer.Close()

	req, _ := http.NewRequest("POST", url, body)
	req.Header.Add("Content-Type", writer.FormDataContentType())

	res := Pin{}

	if err := c.sendRequest(req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
