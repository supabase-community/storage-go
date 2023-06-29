package storage_go

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
)

func (c *Client) ListBuckets() ([]Bucket, BucketResponseError) {
	res, err := c.session.Get(c.clientTransport.baseUrl.String() + "/bucket")
	if err != nil {
		panic(err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(res.Body)

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var data []Bucket
	err = json.Unmarshal(body, &data)

	var respError BucketResponseError
	err = json.Unmarshal(body, &respError)

	return data, respError
}

func (c *Client) GetBucket(id string) (Bucket, BucketResponseError) {
	res, err := c.session.Get(c.clientTransport.baseUrl.String() + "/bucket/" + id)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	var data Bucket
	var error_ BucketResponseError
	err = json.Unmarshal(body, &data)
	err = json.Unmarshal(body, &error_)

	return data, error_
}

func (c *Client) CreateBucket(id string, options BucketOptions) (Bucket, BucketResponseError) {
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
	jsonBody, _ := json.Marshal(bodyData)
	res, err := c.session.Post(c.clientTransport.baseUrl.String()+"/bucket",
		"application/json",
		bytes.NewBuffer(jsonBody))
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	var data Bucket
	var error_ BucketResponseError
	err = json.Unmarshal(body, &data)
	data.Public = options.Public
	err = json.Unmarshal(body, &error_)

	return data, error_
}

func (c *Client) UpdateBucket(id string, options BucketOptions) (MessageResponse, BucketResponseError) {
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
	jsonBody, _ := json.Marshal(bodyData)
	request, err := http.NewRequest(http.MethodPut, c.clientTransport.baseUrl.String()+"/bucket/"+id, bytes.NewBuffer(jsonBody))
	res, err := c.session.Do(request)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	var data MessageResponse
	var error_ BucketResponseError
	err = json.Unmarshal(body, &data)
	err = json.Unmarshal(body, &error_)

	return data, error_
}

func (c *Client) EmptyBucket(id string) (MessageResponse, BucketResponseError) {
	jsonBody, _ := json.Marshal(map[string]interface{}{})
	res, err := c.session.Post(c.clientTransport.baseUrl.String()+"/bucket/"+id+"/empty", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	var data MessageResponse
	var error_ BucketResponseError
	err = json.Unmarshal(body, &data)
	err = json.Unmarshal(body, &error_)

	return data, error_
}

func (c *Client) DeleteBucket(id string) (MessageResponse, BucketResponseError) {
	jsonBody, _ := json.Marshal(map[string]interface{}{})
	request, err := http.NewRequest(http.MethodDelete, c.clientTransport.baseUrl.String()+"/bucket/"+id, bytes.NewBuffer(jsonBody))
	res, err := c.session.Do(request)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	var data MessageResponse
	var error_ BucketResponseError
	err = json.Unmarshal(body, &data)
	err = json.Unmarshal(body, &error_)

	return data, error_
}

type MessageResponse struct {
	Message string `json:"message"`
}

type BucketResponseError struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	StatusCode uint16 `json:"statusCode"`
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
