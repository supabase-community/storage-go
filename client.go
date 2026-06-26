package storage_go

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var version = "v0.8.1"

type Client struct {
	clientError     error
	session         http.Client
	clientTransport transport
}

// FileClient provides file operations scoped to a bucket.
type FileClient struct {
	client   *Client
	bucketID string
}

// BucketClient provides bucket management operations.
type BucketClient struct {
	client *Client
}

// AnalyticsClient provides storage analytics operations.
type AnalyticsClient struct {
	client *Client
}

// VectorClient provides storage vector operations.
type VectorClient struct {
	client *Client
}

type transport struct {
	header  http.Header
	baseUrl url.URL
	base    http.RoundTripper
}

func (t transport) RoundTrip(request *http.Request) (*http.Response, error) {
	for headerName, values := range t.header {
		if request.Header.Get(headerName) != "" {
			continue
		}
		for _, val := range values {
			request.Header.Add(headerName, val)
		}
	}
	if !request.URL.IsAbs() {
		resolved := t.baseUrl
		resolved.Path = strings.TrimRight(t.baseUrl.Path, "/") + "/" + strings.TrimLeft(request.URL.Path, "/")
		resolved.RawPath = ""
		resolved.RawQuery = request.URL.RawQuery
		resolved.Fragment = request.URL.Fragment
		request.URL = &resolved
	}
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(request)
}

func NewClient(rawUrl string, token string, headers map[string]string) *Client {
	baseURL, err := url.Parse(rawUrl)
	if err != nil {
		return &Client{
			clientError: err,
		}
	}
	baseURL.Path = strings.TrimRight(baseURL.Path, "/")
	baseURL.RawPath = ""

	session := http.Client{}
	if http.DefaultClient != nil {
		session = *http.DefaultClient
	}

	t := transport{
		header:  http.Header{},
		baseUrl: *baseURL,
		base:    session.Transport,
	}
	session.Transport = t

	c := Client{
		session:         session,
		clientTransport: t,
	}

	// Set required headers
	c.clientTransport.header.Set("Accept", "application/json")
	c.clientTransport.header.Set("Content-Type", "application/json")
	c.clientTransport.header.Set("X-Client-Info", "storage-go/"+version)
	c.clientTransport.header.Set("Authorization", "Bearer "+token)

	// Optional headers [if exists]
	for key, value := range headers {
		c.clientTransport.header.Set(key, value)
	}

	return &c
}

// From returns a file client scoped to bucketID.
func (c *Client) From(bucketID string) *FileClient {
	return &FileClient{client: c, bucketID: bucketID}
}

// Buckets returns a bucket client.
func (c *Client) Buckets() *BucketClient {
	return &BucketClient{client: c}
}

// Analytics returns an analytics client.
func (c *Client) Analytics() *AnalyticsClient {
	return &AnalyticsClient{client: c}
}

// Vectors returns a vector client.
func (c *Client) Vectors() *VectorClient {
	return &VectorClient{client: c}
}

// NewRequest will create new request with method, url and body
// If body is not nil, it will be marshalled into json
func (c *Client) NewRequest(method, url string, body ...interface{}) (*http.Request, error) {
	var buf io.ReadWriter
	if len(body) > 0 && body[0] != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body[0])
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// Do will send request using the c.sessionon which it is called
// If response contains body, it will be unmarshalled into v
// If response has err, it will be returned
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.session.Do(req)
	if err != nil {
		return nil, err
	}

	err = checkForError(resp)
	if err != nil {
		return resp, err
	}

	if resp.Body != nil && v != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp, err
		}
		if err := resp.Body.Close(); err != nil {
			return resp, err
		}
		err = json.Unmarshal(body, &v)
		if err != nil {
			return resp, err
		}
	}

	return resp, nil
}

func checkForError(resp *http.Response) error {
	if c := resp.StatusCode; 200 <= c && c < 400 {
		return nil
	}

	errorResponse := &StorageError{
		Status:     resp.StatusCode,
		StatusCode: resp.StatusCode,
	}

	if resp.Body == nil {
		return errorResponse
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		errorResponse.Message = err.Error()
		return errorResponse
	}
	if err := resp.Body.Close(); err != nil {
		errorResponse.Message = err.Error()
		return errorResponse
	}
	if len(data) == 0 {
		return errorResponse
	}

	decoded, err := decodeStorageError(data)
	if err != nil {
		errorResponse.RawBody = string(data)
		return errorResponse
	}

	if decoded.StatusCode != 0 {
		errorResponse.Status = decoded.StatusCode
		errorResponse.StatusCode = decoded.StatusCode
	}
	errorResponse.ErrorCode = decoded.ErrorCode
	errorResponse.Message = decoded.Message

	return errorResponse
}

type storageErrorResponse struct {
	StatusCode json.RawMessage `json:"statusCode"`
	ErrorCode  string          `json:"error"`
	Message    string          `json:"message"`
}

func decodeStorageError(data []byte) (StorageError, error) {
	var response storageErrorResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return StorageError{}, err
	}

	return StorageError{
		StatusCode: decodeStorageErrorStatusCode(response.StatusCode),
		ErrorCode:  response.ErrorCode,
		Message:    response.Message,
	}, nil
}

func decodeStorageErrorStatusCode(data json.RawMessage) int {
	var statusCode int
	if err := json.Unmarshal(data, &statusCode); err == nil {
		return statusCode
	}

	var statusCodeString string
	if err := json.Unmarshal(data, &statusCodeString); err != nil {
		return 0
	}

	statusCode, err := strconv.Atoi(statusCodeString)
	if err != nil {
		return 0
	}

	return statusCode
}
