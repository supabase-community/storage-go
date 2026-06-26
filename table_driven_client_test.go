package storage_go

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type observedHTTPRequest struct {
	method      string
	path        string
	escapedPath string
	rawQuery    string
	header      http.Header
	body        []byte
	readErr     error
	closeErr    error
}

func newRecordedClient(
	t *testing.T,
	status int,
	response string,
) (*Client, <-chan observedHTTPRequest) {
	t.Helper()

	requests := make(chan observedHTTPRequest, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, readErr := io.ReadAll(r.Body)
		closeErr := r.Body.Close()
		requests <- observedHTTPRequest{
			method:      r.Method,
			path:        r.URL.Path,
			escapedPath: r.URL.EscapedPath(),
			rawQuery:    r.URL.RawQuery,
			header:      r.Header.Clone(),
			body:        body,
			readErr:     readErr,
			closeErr:    closeErr,
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)

	return NewClient(server.URL+"/storage/v1", "test-token", nil), requests
}

func takeObservedRequest(t *testing.T, requests <-chan observedHTTPRequest) observedHTTPRequest {
	t.Helper()

	select {
	case request := <-requests:
		if request.readErr != nil {
			t.Fatalf("reading observed request body: %v", request.readErr)
		}
		if request.closeErr != nil {
			t.Fatalf("closing observed request body: %v", request.closeErr)
		}
		return request
	default:
		t.Fatal("client call did not send a request")
		return observedHTTPRequest{}
	}
}

func assertJSONFields(t *testing.T, body []byte, want map[string]any) {
	t.Helper()

	if len(want) == 0 {
		return
	}

	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatalf("json.Unmarshal(%q) returned error: %v", string(body), err)
	}
	for key, wantValue := range want {
		gotValue, ok := got[key]
		if !ok {
			t.Errorf("json body %q missing key %q", string(body), key)
			continue
		}
		assertJSONValue(t, key, gotValue, wantValue)
	}
}

func assertJSONValue(t *testing.T, key string, got any, want any) {
	t.Helper()

	switch want := want.(type) {
	case string:
		if got != want {
			t.Errorf("json body[%q] = %v, want %q", key, got, want)
		}
	case bool:
		if got != want {
			t.Errorf("json body[%q] = %v, want %v", key, got, want)
		}
	case int:
		gotNumber, ok := got.(float64)
		if !ok || gotNumber != float64(want) {
			t.Errorf("json body[%q] = %v, want %d", key, got, want)
		}
	case []string:
		gotValues, ok := got.([]any)
		if !ok {
			t.Errorf("json body[%q] = %T(%v), want []string%v", key, got, got, want)
			return
		}
		if len(gotValues) != len(want) {
			t.Errorf("json body[%q] length = %d, want %d", key, len(gotValues), len(want))
			return
		}
		for i, wantItem := range want {
			if gotValues[i] != wantItem {
				t.Errorf("json body[%q][%d] = %v, want %q", key, i, gotValues[i], wantItem)
			}
		}
	default:
		t.Fatalf("unsupported expected JSON value %T for key %q", want, key)
	}
}

func TestClientBasePathTableDrivenRequests(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    string
	}{
		{
			name:    "base path without trailing slash",
			baseURL: "/storage/v1",
			want:    "/storage/v1/bucket",
		},
		{
			name:    "base path with trailing slash",
			baseURL: "/storage/v1/",
			want:    "/storage/v1/bucket",
		},
		{
			name:    "empty base path",
			baseURL: "",
			want:    "/bucket",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requests := make(chan observedHTTPRequest, 1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requests <- observedHTTPRequest{path: r.URL.Path}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`[]`))
			}))
			t.Cleanup(server.Close)

			client := NewClient(server.URL+tt.baseURL, "test-token", nil)
			if _, err := client.ListBuckets(); err != nil {
				t.Fatalf("ListBuckets() returned error: %v", err)
			}

			got := takeObservedRequest(t, requests)
			if got.path != tt.want {
				t.Errorf("ListBuckets() request path = %q, want %q", got.path, tt.want)
			}
		})
	}
}

func TestBucketClientTableDrivenRequests(t *testing.T) {
	tests := []struct {
		name       string
		call       func(*Client) error
		response   string
		wantMethod string
		wantPath   string
		wantQuery  string
		wantBody   map[string]any
	}{
		{
			name: "list buckets",
			call: func(c *Client) error {
				_, err := c.Buckets().ListBuckets()
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/storage/v1/bucket",
		},
		{
			name: "list buckets with options",
			call: func(c *Client) error {
				_, err := c.Buckets().ListBuckets(ListBucketOptions{
					Limit:  10,
					Offset: 2,
					Search: "ava",
					SortBy: SortBy{Column: "name", Order: "desc"},
				})
				return err
			},
			wantMethod: http.MethodGet,
			wantPath:   "/storage/v1/bucket",
			wantQuery:  "limit=10&offset=2&order=desc&search=ava&sortBy=name",
		},
		{
			name: "get bucket",
			call: func(c *Client) error {
				_, err := c.Buckets().GetBucket("avatars")
				return err
			},
			response:   `{"id":"avatars","name":"avatars","public":true}`,
			wantMethod: http.MethodGet,
			wantPath:   "/storage/v1/bucket/avatars",
		},
		{
			name: "create bucket",
			call: func(c *Client) error {
				_, err := c.Buckets().CreateBucket("avatars", BucketOptions{Public: true})
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/storage/v1/bucket",
			wantBody: map[string]any{
				"id":     "avatars",
				"name":   "avatars",
				"public": true,
			},
		},
		{
			name: "update bucket",
			call: func(c *Client) error {
				_, err := c.Buckets().UpdateBucket("avatars", BucketOptions{Public: false})
				return err
			},
			wantMethod: http.MethodPut,
			wantPath:   "/storage/v1/bucket/avatars",
			wantBody: map[string]any{
				"id":     "avatars",
				"name":   "avatars",
				"public": false,
			},
		},
		{
			name: "empty bucket",
			call: func(c *Client) error {
				_, err := c.Buckets().EmptyBucket("avatars")
				return err
			},
			wantMethod: http.MethodPost,
			wantPath:   "/storage/v1/bucket/avatars/empty",
		},
		{
			name: "delete bucket",
			call: func(c *Client) error {
				_, err := c.Buckets().DeleteBucket("avatars")
				return err
			},
			wantMethod: http.MethodDelete,
			wantPath:   "/storage/v1/bucket/avatars",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := tt.response
			if response == "" {
				response = bucketResponseForMethod(tt.wantMethod)
			}
			client, requests := newRecordedClient(t, http.StatusOK, response)
			if err := tt.call(client); err != nil {
				t.Fatalf("%s returned error: %v", tt.name, err)
			}

			got := takeObservedRequest(t, requests)
			if got.method != tt.wantMethod {
				t.Errorf("%s request method = %q, want %q", tt.name, got.method, tt.wantMethod)
			}
			if got.path != tt.wantPath {
				t.Errorf("%s request path = %q, want %q", tt.name, got.path, tt.wantPath)
			}
			if got.rawQuery != tt.wantQuery {
				t.Errorf("%s request query = %q, want %q", tt.name, got.rawQuery, tt.wantQuery)
			}
			assertJSONFields(t, got.body, tt.wantBody)
		})
	}
}

func bucketResponseForMethod(method string) string {
	switch method {
	case http.MethodGet:
		return `[{"id":"avatars","name":"avatars","public":true}]`
	case http.MethodPost, http.MethodPut:
		return `{"id":"avatars","name":"avatars","public":true}`
	default:
		return `{"message":"ok"}`
	}
}

func TestFileClientTableDrivenRequests(t *testing.T) {
	tests := []struct {
		name       string
		call       func(*Client) error
		response   string
		wantMethod string
		wantPath   string
		wantBody   map[string]any
	}{
		{
			name: "upload",
			call: func(c *Client) error {
				_, err := c.From("avatars").Upload("folder/file.txt", strings.NewReader("hello"))
				return err
			},
			response:   `{"key":"avatars/folder/file.txt","path":"folder/file.txt","bucketId":"avatars"}`,
			wantMethod: http.MethodPost,
			wantPath:   "/storage/v1/object/avatars/folder/file.txt",
		},
		{
			name: "update",
			call: func(c *Client) error {
				_, err := c.From("avatars").Update("folder/file.txt", strings.NewReader("hello"))
				return err
			},
			response:   `{"key":"avatars/folder/file.txt"}`,
			wantMethod: http.MethodPut,
			wantPath:   "/storage/v1/object/avatars/folder/file.txt",
		},
		{
			name: "download",
			call: func(c *Client) error {
				_, err := c.From("avatars").Download("folder/file.txt")
				return err
			},
			response:   `hello`,
			wantMethod: http.MethodGet,
			wantPath:   "/storage/v1/object/avatars/folder/file.txt",
		},
		{
			name: "remove",
			call: func(c *Client) error {
				_, err := c.From("avatars").Remove([]string{"folder/file.txt"})
				return err
			},
			response:   `[{"name":"file.txt"}]`,
			wantMethod: http.MethodDelete,
			wantPath:   "/storage/v1/object/avatars",
			wantBody: map[string]any{
				"prefixes": []string{"folder/file.txt"},
			},
		},
		{
			name: "list",
			call: func(c *Client) error {
				_, err := c.From("avatars").List("folder", FileSearchOptions{Limit: 50})
				return err
			},
			response:   `[{"name":"file.txt"}]`,
			wantMethod: http.MethodPost,
			wantPath:   "/storage/v1/object/list/avatars",
			wantBody: map[string]any{
				"prefix": "folder",
				"limit":  50,
			},
		},
		{
			name: "move",
			call: func(c *Client) error {
				_, err := c.From("avatars").Move("old.txt", "new.txt")
				return err
			},
			response:   `{"key":"avatars/new.txt"}`,
			wantMethod: http.MethodPost,
			wantPath:   "/storage/v1/object/move",
			wantBody: map[string]any{
				"bucketId":       "avatars",
				"sourceKey":      "old.txt",
				"destinationKey": "new.txt",
			},
		},
		{
			name: "copy",
			call: func(c *Client) error {
				_, err := c.From("avatars").Copy("old.txt", "new.txt")
				return err
			},
			response:   `{"key":"avatars/new.txt"}`,
			wantMethod: http.MethodPost,
			wantPath:   "/storage/v1/object/copy",
			wantBody: map[string]any{
				"bucketId":       "avatars",
				"sourceKey":      "old.txt",
				"destinationKey": "new.txt",
			},
		},
		{
			name: "create signed url",
			call: func(c *Client) error {
				_, err := c.From("avatars").CreateSignedURL("folder/file.txt", 60)
				return err
			},
			response:   `{"signedURL":"/object/sign/avatars/folder/file.txt?token=token"}`,
			wantMethod: http.MethodPost,
			wantPath:   "/storage/v1/object/sign/avatars/folder/file.txt",
			wantBody: map[string]any{
				"expiresIn": 60,
			},
		},
		{
			name: "create signed urls",
			call: func(c *Client) error {
				_, err := c.From("avatars").CreateSignedURLs([]string{"a.txt", "b.txt"}, 60)
				return err
			},
			response:   `[{"signedURL":"/object/sign/avatars/a.txt?token=token"}]`,
			wantMethod: http.MethodPost,
			wantPath:   "/storage/v1/object/sign/avatars",
			wantBody: map[string]any{
				"expiresIn": 60,
				"paths":     []string{"a.txt", "b.txt"},
			},
		},
		{
			name: "create signed upload url",
			call: func(c *Client) error {
				_, err := c.From("avatars").CreateSignedUploadURL("folder/file.txt")
				return err
			},
			response:   `{"url":"/object/upload/sign/avatars/folder/file.txt?token=token"}`,
			wantMethod: http.MethodPost,
			wantPath:   "/storage/v1/object/upload/sign/avatars/folder/file.txt",
		},
		{
			name: "info",
			call: func(c *Client) error {
				_, err := c.From("avatars").Info("folder/file.txt")
				return err
			},
			response:   `{"name":"file.txt"}`,
			wantMethod: http.MethodGet,
			wantPath:   "/storage/v1/object/info/avatars/folder/file.txt",
		},
		{
			name: "exists",
			call: func(c *Client) error {
				_, err := c.From("avatars").Exists("folder/file.txt")
				return err
			},
			response:   ``,
			wantMethod: http.MethodHead,
			wantPath:   "/storage/v1/object/avatars/folder/file.txt",
		},
		{
			name: "list v2",
			call: func(c *Client) error {
				_, err := c.From("avatars").ListV2("folder", SearchV2Options{Limit: 10})
				return err
			},
			response:   `[{"name":"file.txt","type":"file"}]`,
			wantMethod: http.MethodPost,
			wantPath:   "/storage/v1/object/list-v2/avatars",
			wantBody: map[string]any{
				"prefix": "folder",
				"limit":  10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, requests := newRecordedClient(t, http.StatusOK, tt.response)
			if err := tt.call(client); err != nil {
				t.Fatalf("%s returned error: %v", tt.name, err)
			}

			got := takeObservedRequest(t, requests)
			if got.method != tt.wantMethod {
				t.Errorf("%s request method = %q, want %q", tt.name, got.method, tt.wantMethod)
			}
			if got.path != tt.wantPath {
				t.Errorf("%s request path = %q, want %q", tt.name, got.path, tt.wantPath)
			}
			assertJSONFields(t, got.body, tt.wantBody)
		})
	}
}

func TestAnalyticsClientTableDrivenRequests(t *testing.T) {
	tests := []struct {
		name       string
		call       func(*Client) error
		response   string
		wantMethod string
		wantPath   string
		wantQuery  string
		wantBody   map[string]any
	}{
		{
			name: "create bucket",
			call: func(c *Client) error {
				_, err := c.Analytics().CreateBucket("events")
				return err
			},
			response:   `{"id":"events","name":"events","type":"ANALYTICS"}`,
			wantMethod: http.MethodPost,
			wantPath:   "/storage/v1/bucket",
			wantBody: map[string]any{
				"id":   "events",
				"name": "events",
				"type": "ANALYTICS",
			},
		},
		{
			name: "list buckets",
			call: func(c *Client) error {
				_, err := c.Analytics().ListBuckets()
				return err
			},
			response:   `[{"id":"events","name":"events","type":"ANALYTICS"}]`,
			wantMethod: http.MethodGet,
			wantPath:   "/storage/v1/bucket",
			wantQuery:  "type=ANALYTICS",
		},
		{
			name: "delete bucket",
			call: func(c *Client) error {
				return c.Analytics().DeleteBucket("events")
			},
			response:   `{"message":"ok"}`,
			wantMethod: http.MethodDelete,
			wantPath:   "/storage/v1/bucket/events",
		},
		{
			name: "iceberg catalog request",
			call: func(c *Client) error {
				var out map[string]any
				return c.Analytics().From("events").Do(http.MethodGet, "/v1/config", nil, &out)
			},
			response:   `{}`,
			wantMethod: http.MethodGet,
			wantPath:   "/storage/v1/iceberg/events/v1/config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, requests := newRecordedClient(t, http.StatusOK, tt.response)
			if err := tt.call(client); err != nil {
				t.Fatalf("%s returned error: %v", tt.name, err)
			}

			got := takeObservedRequest(t, requests)
			if got.method != tt.wantMethod {
				t.Errorf("%s request method = %q, want %q", tt.name, got.method, tt.wantMethod)
			}
			if got.path != tt.wantPath {
				t.Errorf("%s request path = %q, want %q", tt.name, got.path, tt.wantPath)
			}
			if got.rawQuery != tt.wantQuery {
				t.Errorf("%s request query = %q, want %q", tt.name, got.rawQuery, tt.wantQuery)
			}
			assertJSONFields(t, got.body, tt.wantBody)
		})
	}
}

func TestVectorClientTableDrivenRequests(t *testing.T) {
	tests := []struct {
		name     string
		call     func(*Client) error
		response string
		wantPath string
		wantBody map[string]any
	}{
		{
			name: "create vector bucket",
			call: func(c *Client) error {
				_, err := c.Vectors().CreateBucket("embeddings")
				return err
			},
			response: `{"vectorBucket":{"name":"embeddings"}}`,
			wantPath: "/storage/v1/vector/CreateVectorBucket",
			wantBody: map[string]any{"vectorBucketName": "embeddings"},
		},
		{
			name: "get vector bucket",
			call: func(c *Client) error {
				_, err := c.Vectors().GetBucket("embeddings")
				return err
			},
			response: `{"vectorBucket":{"name":"embeddings"}}`,
			wantPath: "/storage/v1/vector/GetVectorBucket",
			wantBody: map[string]any{"vectorBucketName": "embeddings"},
		},
		{
			name: "list vector buckets",
			call: func(c *Client) error {
				_, err := c.Vectors().ListBuckets(ListVectorBucketsOptions{Prefix: "prod-", MaxResults: 5})
				return err
			},
			response: `{"vectorBuckets":[]}`,
			wantPath: "/storage/v1/vector/ListVectorBuckets",
			wantBody: map[string]any{"prefix": "prod-", "maxResults": 5},
		},
		{
			name: "delete vector bucket",
			call: func(c *Client) error {
				_, err := c.Vectors().DeleteBucket("embeddings")
				return err
			},
			response: `{"success":true}`,
			wantPath: "/storage/v1/vector/DeleteVectorBucket",
			wantBody: map[string]any{"vectorBucketName": "embeddings"},
		},
		{
			name: "create index",
			call: func(c *Client) error {
				_, err := c.Vectors().From("embeddings").CreateIndex(CreateIndexOptions{Name: "docs", Dimension: 3})
				return err
			},
			response: `{"vectorIndex":{"name":"docs"}}`,
			wantPath: "/storage/v1/vector/CreateIndex",
			wantBody: map[string]any{"vectorBucketName": "embeddings", "name": "docs", "dimension": 3},
		},
		{
			name: "get index",
			call: func(c *Client) error {
				_, err := c.Vectors().From("embeddings").GetIndex("docs")
				return err
			},
			response: `{"vectorIndex":{"name":"docs"}}`,
			wantPath: "/storage/v1/vector/GetIndex",
			wantBody: map[string]any{"vectorBucketName": "embeddings", "vectorIndexName": "docs"},
		},
		{
			name: "list indexes",
			call: func(c *Client) error {
				_, err := c.Vectors().From("embeddings").ListIndexes(ListIndexesOptions{Prefix: "doc"})
				return err
			},
			response: `{"vectorIndexes":[]}`,
			wantPath: "/storage/v1/vector/ListIndexes",
			wantBody: map[string]any{"vectorBucketName": "embeddings", "prefix": "doc"},
		},
		{
			name: "delete index",
			call: func(c *Client) error {
				_, err := c.Vectors().From("embeddings").DeleteIndex("docs")
				return err
			},
			response: `{"success":true}`,
			wantPath: "/storage/v1/vector/DeleteIndex",
			wantBody: map[string]any{"vectorBucketName": "embeddings", "vectorIndexName": "docs"},
		},
		{
			name: "put vectors",
			call: func(c *Client) error {
				_, err := c.Vectors().From("embeddings").Index("docs").PutVectors(PutVectorsOptions{Vectors: []VectorData{{ID: "1"}}})
				return err
			},
			response: `{"success":true}`,
			wantPath: "/storage/v1/vector/PutVectors",
			wantBody: map[string]any{"vectorBucketName": "embeddings", "vectorIndexName": "docs"},
		},
		{
			name: "get vectors",
			call: func(c *Client) error {
				_, err := c.Vectors().From("embeddings").Index("docs").GetVectors(GetVectorsOptions{IDs: []string{"1"}})
				return err
			},
			response: `{"vectors":[]}`,
			wantPath: "/storage/v1/vector/GetVectors",
			wantBody: map[string]any{"vectorBucketName": "embeddings", "vectorIndexName": "docs", "ids": []string{"1"}},
		},
		{
			name: "list vectors",
			call: func(c *Client) error {
				_, err := c.Vectors().From("embeddings").Index("docs").ListVectors(ListVectorsOptions{Prefix: "doc"})
				return err
			},
			response: `{"vectors":[]}`,
			wantPath: "/storage/v1/vector/ListVectors",
			wantBody: map[string]any{"vectorBucketName": "embeddings", "vectorIndexName": "docs", "prefix": "doc"},
		},
		{
			name: "query vectors",
			call: func(c *Client) error {
				_, err := c.Vectors().From("embeddings").Index("docs").QueryVectors(QueryVectorsOptions{Vector: []float64{1, 2, 3}, TopK: 3})
				return err
			},
			response: `{"matches":[]}`,
			wantPath: "/storage/v1/vector/QueryVectors",
			wantBody: map[string]any{"vectorBucketName": "embeddings", "vectorIndexName": "docs", "topK": 3},
		},
		{
			name: "delete vectors",
			call: func(c *Client) error {
				_, err := c.Vectors().From("embeddings").Index("docs").DeleteVectors(DeleteVectorsOptions{IDs: []string{"1"}})
				return err
			},
			response: `{"success":true}`,
			wantPath: "/storage/v1/vector/DeleteVectors",
			wantBody: map[string]any{"vectorBucketName": "embeddings", "vectorIndexName": "docs", "ids": []string{"1"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, requests := newRecordedClient(t, http.StatusOK, tt.response)
			if err := tt.call(client); err != nil {
				t.Fatalf("%s returned error: %v", tt.name, err)
			}

			got := takeObservedRequest(t, requests)
			if got.method != http.MethodPost {
				t.Errorf("%s request method = %q, want %q", tt.name, got.method, http.MethodPost)
			}
			if got.path != tt.wantPath {
				t.Errorf("%s request path = %q, want %q", tt.name, got.path, tt.wantPath)
			}
			assertJSONFields(t, got.body, tt.wantBody)
		})
	}
}
