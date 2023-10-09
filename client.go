package storage_go

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

var version = "v0.7.0"

type Client struct {
	clientError     error
	session         http.Client
	clientTransport transport
}

type transport struct {
	header  http.Header
	baseUrl url.URL
}

func (t transport) RoundTrip(request *http.Request) (*http.Response, error) {
	for headerName, values := range t.header {
		for _, val := range values {
			request.Header.Add(headerName, val)
		}
	}
	request.URL = t.baseUrl.ResolveReference(request.URL)
	return http.DefaultTransport.RoundTrip(request)
}

func NewClient(rawUrl string, token string, headers map[string]string) *Client {
	baseURL, err := url.Parse(rawUrl)
	if err != nil {
		return &Client{
			clientError: err,
		}
	}

	t := transport{
		header:  http.Header{},
		baseUrl: *baseURL,
	}

	c := Client{
		session:         http.Client{Transport: t},
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
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
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

	errorResponse := &StorageError{}

	data, err := io.ReadAll(resp.Body)
	if err == nil && data != nil {
		_ = json.Unmarshal(data, errorResponse)
	}

	return errorResponse
}
