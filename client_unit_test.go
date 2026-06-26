package storage_go

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"
)

type trackingReadCloser struct {
	*strings.Reader
	closed bool
}

type clientInfoObservedRequest struct {
	clientInfo            string
	deprecatedPlatform    string
	deprecatedRuntime     string
	deprecatedRuntimeVer  string
	deprecatedPlatformVer string
}

func (r *trackingReadCloser) Close() error {
	r.closed = true
	return nil
}

func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client := NewClient(server.URL+"/", "test-token", map[string]string{
		"x-test-header": "test-value",
	})

	return client, server
}

func decodeJSONBody[T any](t *testing.T, r *http.Request) T {
	t.Helper()

	var body T
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Errorf("decodeJSONBody: decode request body: %v", err)
		if closeErr := r.Body.Close(); closeErr != nil {
			t.Errorf("decodeJSONBody: close request body: %v", closeErr)
		}
		var zero T
		return zero
	}
	if err := r.Body.Close(); err != nil {
		t.Errorf("decodeJSONBody: close request body: %v", err)
		var zero T
		return zero
	}

	return body
}

func TestClientExposesDomainClients(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if got := client.Buckets(); got == nil {
		t.Fatal("Buckets() = nil, want *BucketClient")
	} else if got.client != client {
		t.Errorf("Buckets().client = %p, want %p", got.client, client)
	}
	if got := client.From("avatars"); got == nil {
		t.Fatalf("From(%q) = nil, want *FileClient", "avatars")
	} else {
		if got.client != client {
			t.Errorf("From(%q).client = %p, want %p", "avatars", got.client, client)
		}
		if got.bucketID != "avatars" {
			t.Errorf("From(%q).bucketID = %q, want %q", "avatars", got.bucketID, "avatars")
		}
	}
	if got := client.Analytics(); got == nil {
		t.Fatal("Analytics() = nil, want *AnalyticsClient")
	} else if got.client != client {
		t.Errorf("Analytics().client = %p, want %p", got.client, client)
	}
	if got := client.Vectors(); got == nil {
		t.Fatal("Vectors() = nil, want *VectorClient")
	} else if got.client != client {
		t.Errorf("Vectors().client = %p, want %p", got.client, client)
	}
}

func TestClientAddsAuthorizationAndCustomHeaders(t *testing.T) {
	defaultClient := http.DefaultClient
	http.DefaultClient = &http.Client{Timeout: time.Second}
	t.Cleanup(func() {
		http.DefaultClient = defaultClient
	})

	type observedRequest struct {
		path          string
		authorization string
		customHeader  string
	}

	observed := make(chan observedRequest, 1)
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		observed <- observedRequest{
			path:          r.URL.EscapedPath(),
			authorization: r.Header.Get("authorization"),
			customHeader:  r.Header.Get("x-test-header"),
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	})

	if got, want := client.session.Timeout, time.Second; got != want {
		t.Errorf("NewClient(%q).session.Timeout = %v, want %v", "server URL", got, want)
	}

	if _, err := client.ListBuckets(); err != nil {
		t.Fatalf("ListBuckets returned error: %v", err)
	}

	var got observedRequest
	select {
	case got = <-observed:
	default:
		t.Fatal("ListBuckets did not call handler")
	}

	if got.path != "/bucket" {
		t.Errorf("ListBuckets request path = %q, want %q", got.path, "/bucket")
	}
	if want := "Bearer test-token"; got.authorization != want {
		t.Errorf("authorization header = %q, want %q", got.authorization, want)
	}
	if want := "test-value"; got.customHeader != want {
		t.Errorf("x-test-header header = %q, want %q", got.customHeader, want)
	}
}

func TestClientSetsStructuredClientInfoHeader(t *testing.T) {
	observed := make(chan clientInfoObservedRequest, 1)
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		observed <- clientInfoObservedRequest{
			clientInfo:            r.Header.Get("x-client-info"),
			deprecatedPlatform:    r.Header.Get("x-supabase-client-platform"),
			deprecatedRuntime:     r.Header.Get("x-supabase-client-runtime"),
			deprecatedRuntimeVer:  r.Header.Get("x-supabase-client-runtime-version"),
			deprecatedPlatformVer: r.Header.Get("x-supabase-client-platform-version"),
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	})

	if _, err := client.ListBuckets(); err != nil {
		t.Fatalf("ListBuckets() returned error: %v", err)
	}

	got := takeStructuredHeaderRequest(t, observed)
	params := parseClientInfoHeader(t, got.clientInfo)
	if got.clientInfo == "" {
		t.Fatal("X-Client-Info header = empty, want structured value")
	}
	if token := strings.Split(got.clientInfo, ";")[0]; token != "storage-go/"+version {
		t.Errorf("X-Client-Info library token = %q, want %q", token, "storage-go/"+version)
	}
	if got, want := params["platform"], runtime.GOOS; got != want {
		t.Errorf("X-Client-Info platform = %q, want %q", got, want)
	}
	if got, want := params["runtime"], "go"; got != want {
		t.Errorf("X-Client-Info runtime = %q, want %q", got, want)
	}
	if got, want := params["runtime-version"], strings.TrimPrefix(runtime.Version(), "go"); got != want {
		t.Errorf("X-Client-Info runtime-version = %q, want %q", got, want)
	}
	for _, key := range []string{"platform-version", "framework", "framework-version"} {
		if value, ok := params[key]; ok {
			t.Errorf("X-Client-Info %s = %q, want omitted", key, value)
		}
	}
	if got.deprecatedPlatform != "" || got.deprecatedRuntime != "" || got.deprecatedRuntimeVer != "" || got.deprecatedPlatformVer != "" {
		t.Fatalf("deprecated Supabase client metadata headers = %#v, want all empty", got)
	}
}

func TestClientPreservesUserSuppliedClientInfoHeader(t *testing.T) {
	observed := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observed <- r.Header.Get("x-client-info")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(server.Close)

	want := "custom-sdk/9.9.9; framework=myapp"
	client := NewClient(server.URL, "test-token", map[string]string{"X-Client-Info": want})
	if _, err := client.ListBuckets(); err != nil {
		t.Fatalf("ListBuckets() returned error: %v", err)
	}

	select {
	case got := <-observed:
		if got != want {
			t.Fatalf("X-Client-Info header = %q, want %q", got, want)
		}
	default:
		t.Fatal("ListBuckets() did not call handler")
	}
}

func takeStructuredHeaderRequest(t *testing.T, observed <-chan clientInfoObservedRequest) clientInfoObservedRequest {
	t.Helper()

	select {
	case got := <-observed:
		return got
	default:
		t.Fatal("ListBuckets() did not call handler")
		return clientInfoObservedRequest{}
	}
}

func parseClientInfoHeader(t *testing.T, header string) map[string]string {
	t.Helper()

	parts := strings.Split(header, ";")
	params := map[string]string{}
	for _, part := range parts[1:] {
		part = strings.TrimSpace(part)
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			t.Fatalf("X-Client-Info parameter %q is not key=value", part)
		}
		params[key] = value
	}

	return params
}

func TestClientSendsJSONRequestBody(t *testing.T) {
	type requestBody struct {
		Name   string `json:"name"`
		Public bool   `json:"public"`
	}
	type observedRequest struct {
		method string
		path   string
		body   requestBody
	}

	observed := make(chan observedRequest, 1)
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		observed <- observedRequest{
			method: r.Method,
			path:   r.URL.EscapedPath(),
			body:   decodeJSONBody[requestBody](t, r),
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	wantBody := requestBody{Name: "avatars", Public: true}
	req, err := client.NewRequest(http.MethodPost, "/bucket", wantBody)
	if err != nil {
		t.Fatalf("NewRequest(%q, %q, %#v) returned error: %v", http.MethodPost, "/bucket", wantBody, err)
	}

	var out struct {
		OK bool `json:"ok"`
	}
	if _, err := client.Do(req, &out); err != nil {
		t.Fatalf("Do(NewRequest(%q, %q, %#v)) returned error: %v", http.MethodPost, "/bucket", wantBody, err)
	}
	if !out.OK {
		t.Errorf("Do(NewRequest(%q, %q, %#v)) decoded OK = false, want true", http.MethodPost, "/bucket", wantBody)
	}

	var got observedRequest
	select {
	case got = <-observed:
	default:
		t.Fatal("Do(NewRequest) did not call handler")
	}

	if got.method != http.MethodPost {
		t.Errorf("NewRequest(%q, %q, %#v) sent method = %q, want %q", http.MethodPost, "/bucket", wantBody, got.method, http.MethodPost)
	}
	if got.path != "/bucket" {
		t.Errorf("NewRequest(%q, %q, %#v) sent path = %q, want %q", http.MethodPost, "/bucket", wantBody, got.path, "/bucket")
	}
	if got.body != wantBody {
		t.Errorf("NewRequest(%q, %q, %#v) sent body = %#v, want %#v", http.MethodPost, "/bucket", wantBody, got.body, wantBody)
	}
}

func TestClientPreservesBasePathForRelativeRequests(t *testing.T) {
	observed := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observed <- r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	t.Cleanup(server.Close)

	client := NewClient(server.URL+"/storage/v1", "test-token", nil)
	if _, err := client.ListBuckets(); err != nil {
		t.Fatalf("ListBuckets returned error: %v", err)
	}

	select {
	case got := <-observed:
		if want := "/storage/v1/bucket"; got != want {
			t.Fatalf("request path = %q, want %q", got, want)
		}
	default:
		t.Fatal("ListBuckets did not call handler")
	}
}

func TestClientDecodesStorageError(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"statusCode":"400","error":"invalid_request","message":"invalid bucket"}`))
	})

	req, err := client.NewRequest(http.MethodGet, "/bucket/bad")
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	var out Bucket
	_, err = client.Do(req, &out)
	if err == nil {
		t.Fatal("Do returned nil error, want StorageError")
	}

	storageErr, ok := err.(*StorageError)
	if !ok {
		t.Fatalf("error type = %T, want *StorageError", err)
	}
	if got, want := storageErr.StatusCode, http.StatusBadRequest; got != want {
		t.Fatalf("status code = %d, want %d", got, want)
	}
	if got, want := storageErr.ErrorCode, "invalid_request"; got != want {
		t.Fatalf("error code = %q, want %q", got, want)
	}
	if got, want := storageErr.Message, "invalid bucket"; got != want {
		t.Fatalf("message = %q, want %q", got, want)
	}
}

func TestClientClosesStorageErrorBody(t *testing.T) {
	body := &trackingReadCloser{Reader: strings.NewReader(`{"message":"invalid bucket"}`)}

	err := checkForError(&http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       body,
	})
	if err == nil {
		t.Fatal("checkForError returned nil, want error")
	}
	if !body.closed {
		t.Fatal("checkForError did not close response body")
	}
}
