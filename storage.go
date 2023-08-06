package storage_go

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
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

func (c *Client) UploadOrUpdateFile(bucketId string, relativePath string, data io.Reader, update bool) FileUploadResponse {
	c.clientTransport.header.Set("cache-control", defaultFileCacheControl)
	if c.clientTransport.header == nil {
		c.clientTransport.header.Set("content-type", defaultFileContentType)
	}
	c.clientTransport.header.Set("x-upsert", strconv.FormatBool(defaultFileUpsert))
	body := bufio.NewReader(data)
	_path := removeEmptyFolderName(bucketId + "/" + relativePath)

	var (
		res *http.Response
		err error
	)

	if update {
		var request *http.Request
		request, err = http.NewRequest(http.MethodPut, c.clientTransport.baseUrl.String()+"/object/"+_path, body)
		res, err = c.session.Do(request)
	} else {
		var request *http.Request
		request, err = http.NewRequest(http.MethodPost, c.clientTransport.baseUrl.String()+"/object/"+_path, body)
		res, err = c.session.Do(request)
	}
	if err != nil {
		panic(err)
	}

	body_, err := ioutil.ReadAll(res.Body)
	var response FileUploadResponse
	err = json.Unmarshal(body_, &response)

	return response
}

func (c *Client) UpdateFile(bucketId string, relativePath string, data io.Reader) FileUploadResponse {
	return c.UploadOrUpdateFile(bucketId, relativePath, data, true)
}

func (c *Client) UploadFile(bucketId string, relativePath string, data io.Reader) FileUploadResponse {
	return c.UploadOrUpdateFile(bucketId, relativePath, data, false)
}

func (c *Client) MoveFile(bucketId string, sourceKey string, destinationKey string) FileUploadResponse {
	jsonBody, _ := json.Marshal(map[string]interface{}{
		"bucketId":       bucketId,
		"sourceKey":      sourceKey,
		"destinationKey": destinationKey,
	})

	request, err := http.NewRequest(
		http.MethodPost,
		c.clientTransport.baseUrl.String()+"/object/move",
		bytes.NewBuffer(jsonBody))

	res, err := c.session.Do(request)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	var response FileUploadResponse
	err = json.Unmarshal(body, &response)

	return response
}

func (c *Client) CreateSignedUrl(bucketId string, filePath string, expiresIn int) SignedUrlResponse {
	jsonBody, _ := json.Marshal(map[string]interface{}{
		"expiresIn": expiresIn,
	})

	request, err := http.NewRequest(
		http.MethodPost,
		c.clientTransport.baseUrl.String()+"/object/sign/"+bucketId+"/"+filePath,
		bytes.NewBuffer(jsonBody))

	res, err := c.session.Do(request)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	var response SignedUrlResponse
	err = json.Unmarshal(body, &response)
	response.SignedURL = c.clientTransport.baseUrl.String() + response.SignedURL

	return response
}

func (c *Client) GetPublicUrl(bucketId string, filePath string) SignedUrlResponse {
	var response SignedUrlResponse

	response.SignedURL = c.clientTransport.baseUrl.String() + "/object/public/" + bucketId + "/" + filePath

	return response
}

func (c *Client) RemoveFile(bucketId string, paths []string) FileUploadResponse {
	jsonBody, _ := json.Marshal(map[string]interface{}{
		"prefixes": paths,
	})

	request, err := http.NewRequest(
		http.MethodDelete,
		c.clientTransport.baseUrl.String()+"/object/"+bucketId,
		bytes.NewBuffer(jsonBody))

	res, err := c.session.Do(request)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	var response FileUploadResponse
	err = json.Unmarshal(body, &response)
	response.Data = body

	return response
}

func (c *Client) ListFiles(bucketId string, queryPath string, options FileSearchOptions) []FileObject {
	if options.Offset == 0 {
		options.Offset = defaultOffset
	}

	if options.Limit == 0 {
		options.Limit = defaultLimit
	}

	if options.SortByOptions.Order == "" {
		options.SortByOptions.Order = defaultSortOrder
	}

	if options.SortByOptions.Column == "" {
		options.SortByOptions.Column = defaultSortColumn
	}

	body_ := ListFileRequestBody{
		Limit:  options.Limit,
		Offset: options.Offset,
		SortByOptions: SortBy{
			Column: options.SortByOptions.Column,
			Order:  options.SortByOptions.Order,
		},
		Prefix: queryPath,
	}
	jsonBody, _ := json.Marshal(body_)

	request, err := http.NewRequest(
		http.MethodPost,
		c.clientTransport.baseUrl.String()+"/object/list/"+bucketId,
		bytes.NewBuffer(jsonBody))

	res, err := c.session.Do(request)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	var response []FileObject

	err = json.Unmarshal(body, &response)

	return response
}

// removeEmptyFolderName replaces occurances of double slashes (//)  with a single slash /
// returns a path string with all double slashes replaced with single slash /
func removeEmptyFolderName(filePath string) string {
	return regexp.MustCompile(`\/\/`).ReplaceAllString(filePath, "/")
}

type SortBy struct {
	Column string `json:"column"`
	Order  string `json:"order"`
}

type FileUploadResponse struct {
	Key     string `json:"Key"`
	Message string `json:"message"`
	Data    []byte
}

type SignedUrlResponse struct {
	SignedURL string `json:"signedURL"`
}

type FileSearchOptions struct {
	Limit         int    `json:"limit"`
	Offset        int    `json:"offset"`
	SortByOptions SortBy `json:"sortBy"`
}

type FileObject struct {
	Name           string      `json:"name"`
	BucketId       string      `json:"bucket_id"`
	Owner          string      `json:"owner"`
	Id             string      `json:"id"`
	UpdatedAt      string      `json:"updated_at"`
	CreatedAt      string      `json:"created_at"`
	LastAccessedAt string      `json:"last_accessed_at"`
	Metadata       interface{} `json:"metadata"`
	Buckets        Bucket
}

type ListFileRequestBody struct {
	Limit         int    `json:"limit"`
	Offset        int    `json:"offset"`
	SortByOptions SortBy `json:"sortBy"`
	Prefix        string `json:"prefix"`
}
