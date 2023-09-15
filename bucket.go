package storage_go

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) ListBuckets() ([]Bucket, error) {
	res, err := c.session.Get(c.clientTransport.baseUrl.String() + "/bucket")
	if err != nil {
		return []Bucket{}, err
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return []Bucket{}, err
	}

	var data []Bucket
	err = json.Unmarshal(body, &data)
	if err != nil {
		return []Bucket{}, err
	}

	var respError BucketResponseError
	_ = json.Unmarshal(body, &respError)
	if respError.Errors != "" {
		return []Bucket{}, respError
	}

	return data, nil
}

func (c *Client) GetBucket(id string) (Bucket, error) {
	res, err := c.session.Get(c.clientTransport.baseUrl.String() + "/bucket/" + id)
	if err != nil {
		return Bucket{}, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Bucket{}, err
	}
	var data Bucket
	err = json.Unmarshal(body, &data)
	if err != nil {
		return Bucket{}, err
	}

	var respError BucketResponseError
	_ = json.Unmarshal(body, &respError)
	if respError.Errors != "" {
		return Bucket{}, respError
	}

	return data, nil
}

func (c *Client) CreateBucket(id string, options BucketOptions) (Bucket, error) {
	bodyData := map[string]interface{}{
		"id":     id,
		"name":   id,
		"public": options.Public,
	}
	// We only set the file size limit if it's not empty
	if len(options.FileSizeLimit) > 0 {
		bodyData["file_size_limit"] = options.FileSizeLimit
	}
	// We only set the allowed mime types if it's not empty
	if len(options.AllowedMimeTypes) > 0 {
		bodyData["allowed_mime_types"] = options.AllowedMimeTypes
	}
	jsonBody, err := json.Marshal(bodyData)
	if err != nil {
		return Bucket{}, err
	}

	res, err := c.session.Post(c.clientTransport.baseUrl.String()+"/bucket",
		"application/json",
		bytes.NewBuffer(jsonBody))
	if err != nil {
		return Bucket{}, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Bucket{}, err
	}

	var data Bucket
	var errResp BucketResponseError
	err = json.Unmarshal(body, &data)
	if err != nil {
		return Bucket{}, err
	}

	data.Public = options.Public
	_ = json.Unmarshal(body, &errResp)
	if errResp.Errors != "" {
		return Bucket{}, errResp
	}

	return data, nil
}

func (c *Client) UpdateBucket(id string, options BucketOptions) (MessageResponse, error) {
	bodyData := map[string]interface{}{
		"id":     id,
		"name":   id,
		"public": options.Public,
	}
	// We only set the file size limit if it's not empty
	if len(options.FileSizeLimit) > 0 {
		bodyData["file_size_limit"] = options.FileSizeLimit
	}
	// We only set the allowed mime types if it's not empty
	if len(options.AllowedMimeTypes) > 0 {
		bodyData["allowed_mime_types"] = options.AllowedMimeTypes
	}
	jsonBody, err := json.Marshal(bodyData)
	if err != nil {
		return MessageResponse{}, err
	}
	request, err := http.NewRequest(http.MethodPut, c.clientTransport.baseUrl.String()+"/bucket/"+id, bytes.NewBuffer(jsonBody))
	if err != nil {
		return MessageResponse{}, err
	}

	res, err := c.session.Do(request)
	if err != nil {
		return MessageResponse{}, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return MessageResponse{}, err
	}

	var data MessageResponse
	var errResp BucketResponseError
	err = json.Unmarshal(body, &data)
	if err != nil {
		return MessageResponse{}, err
	}

	_ = json.Unmarshal(body, &errResp)
	if errResp.Errors != "" {
		return MessageResponse{}, err
	}

	return data, nil
}

func (c *Client) EmptyBucket(id string) (MessageResponse, error) {
	jsonBody, err := json.Marshal(map[string]interface{}{})
	if err != nil {
		return MessageResponse{}, err
	}

	res, err := c.session.Post(c.clientTransport.baseUrl.String()+"/bucket/"+id+"/empty", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return MessageResponse{}, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return MessageResponse{}, err
	}

	var data MessageResponse
	var errResp BucketResponseError
	err = json.Unmarshal(body, &data)
	if err != nil {
		return MessageResponse{}, err
	}
	_ = json.Unmarshal(body, &errResp)
	if errResp.Errors != "" {
		return data, errResp
	}

	return data, nil
}

func (c *Client) DeleteBucket(id string) (MessageResponse, error) {
	jsonBody, err := json.Marshal(map[string]interface{}{})
	if err != nil {
		return MessageResponse{}, err
	}

	request, err := http.NewRequest(http.MethodDelete, c.clientTransport.baseUrl.String()+"/bucket/"+id, bytes.NewBuffer(jsonBody))
	if err != nil {
		return MessageResponse{}, err
	}

	res, err := c.session.Do(request)
	if err != nil {
		return MessageResponse{}, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return MessageResponse{}, err
	}

	var data MessageResponse
	var errResp BucketResponseError
	err = json.Unmarshal(body, &data)
	if err != nil {
		return MessageResponse{}, err
	}

	_ = json.Unmarshal(body, &errResp)
	if errResp.Errors != "" {
		return data, errResp
	}

	return data, nil
}

type MessageResponse struct {
	Message string `json:"message"`
}

type BucketResponseError struct {
	Errors     string `json:"error"`
	Message    string `json:"message"`
	StatusCode uint16 `json:"statusCode"`
}

func (b BucketResponseError) Error() string {
	return fmt.Sprintf("status %d: err %s: message %s", b.StatusCode, b.Errors, b.Message)
}

type Bucket struct {
	Id               string   `json:"id"`
	Name             string   `json:"name"`
	Owner            string   `json:"owner"`
	Public           bool     `json:"public"`
	FileSizeLimit    string   `json:"file_size_limit"`
	AllowedMimeTypes []string `json:"allowed_mine_types"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
}

type BucketOptions struct {
	Public           bool
	FileSizeLimit    string
	AllowedMimeTypes []string
}
