package storage_go

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"strconv"
)

const (
	defaultLimit            = 100
	defaultOffset           = 0
	defaultFileCacheControl = "3600"
	defaultFileContentType  = "text/plain;charset=UTF-8"
	defaultFileUpsert       = false
	defaultSortColumn       = "name"
	defaultSortOrder        = "asc"
)

func (c *Client) Upload(bucketId string, relativePath string, data []byte) FileUploadResponse {
	c.clientTransport.header.Set("cache-control", defaultFileCacheControl)
	c.clientTransport.header.Set("content-type", defaultFileContentType)
	c.clientTransport.header.Set("x-upsert", strconv.FormatBool(defaultFileUpsert))

	body := bytes.NewBuffer(data)
	_path := bucketId + "/" + relativePath

	res, err := c.session.Post(
		c.clientTransport.baseUrl.String()+"/object/"+_path,
		defaultFileContentType,
		body)
	if err != nil {
		panic(err)
	}

	body_, err := ioutil.ReadAll(res.Body)
	var response FileUploadResponse
	err = json.Unmarshal(body_, &response)

	return response
}

type SortBy struct {
	Column string
	Order  string
}

type FileUploadResponse struct {
	Key string `json:"Key"`
}
