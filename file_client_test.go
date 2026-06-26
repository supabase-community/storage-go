package storage_go

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func stringPtr(s string) *string {
	return &s
}

func TestFileClientUploadUsesBucketObjectPath(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Errorf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/object/avatars/path/file.txt"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("content-type"), "text/plain"; got != want {
			t.Errorf("content-type = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"key":"avatars/path/file.txt","path":"path/file.txt","bucketId":"avatars"}`))
	})

	got, err := client.From("avatars").Upload(
		"path/file.txt",
		strings.NewReader("hello"),
		FileOptions{ContentType: stringPtr("text/plain")},
	)
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if got.Path != "path/file.txt" {
		t.Fatalf("path = %q, want %q", got.Path, "path/file.txt")
	}
}

func TestFileClientUpdateUsesPutObjectPath(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPut; got != want {
			t.Errorf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/object/avatars/path/file.txt"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"key":"avatars/path/file.txt"}`))
	})

	if _, err := client.From("avatars").Update("path/file.txt", strings.NewReader("hello")); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
}

func TestFileClientDownloadReturnsBytes(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Errorf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/object/avatars/path/file.txt"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	})

	got, err := client.From("avatars").Download("path/file.txt")
	if err != nil {
		t.Fatalf("Download returned error: %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("download body = %q, want %q", string(got), "hello")
	}
}

func TestFileClientListUsesListEndpoint(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Errorf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/object/list/avatars"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		body := decodeJSONBody[ListFileRequestBody](t, r)
		if got, want := body.Prefix, "folder"; got != want {
			t.Errorf("prefix = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"name":"file.txt"}]`))
	})

	files, err := client.From("avatars").List("folder", FileSearchOptions{})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if got, want := len(files), 1; got != want {
		t.Fatalf("file count = %d, want %d", got, want)
	}
}

func TestFileClientRemoveUsesObjectBucketEndpoint(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodDelete; got != want {
			t.Errorf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/object/avatars"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		body := decodeJSONBody[map[string][]string](t, r)
		if got, want := body["prefixes"][0], "path/file.txt"; got != want {
			t.Errorf("prefix = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"name":"file.txt"}]`))
	})

	if _, err := client.From("avatars").Remove([]string{"path/file.txt"}); err != nil {
		t.Fatalf("Remove returned error: %v", err)
	}
}

func TestCompatibilityUploadFileDelegatesToFileClient(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/object/avatars/path/file.txt"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"key":"avatars/path/file.txt","path":"path/file.txt","bucketId":"avatars"}`))
	})

	if _, err := client.UploadFile("avatars", "path/file.txt", strings.NewReader("hello")); err != nil {
		t.Fatalf("UploadFile returned error: %v", err)
	}
}

func TestFileClientUploadSendsBody(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read upload body: %v", err)
		}
		if err := r.Body.Close(); err != nil {
			t.Errorf("close upload body: %v", err)
		}
		if got, want := string(body), "hello"; got != want {
			t.Errorf("body = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"key":"avatars/path/file.txt"}`))
	})

	if _, err := client.From("avatars").Upload("path/file.txt", strings.NewReader("hello")); err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
}

func TestFileClientCopyUsesCopyEndpoint(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Errorf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/object/copy"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		body := decodeJSONBody[map[string]string](t, r)
		if got, want := body["bucketId"], "avatars"; got != want {
			t.Errorf("bucketId = %q, want %q", got, want)
		}
		if got, want := body["sourceKey"], "old.txt"; got != want {
			t.Errorf("sourceKey = %q, want %q", got, want)
		}
		if got, want := body["destinationKey"], "new.txt"; got != want {
			t.Errorf("destinationKey = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"key":"avatars/new.txt"}`))
	})

	if _, err := client.From("avatars").Copy("old.txt", "new.txt"); err != nil {
		t.Fatalf("Copy returned error: %v", err)
	}
}

func TestFileClientCreateSignedURLsUsesBucketEndpoint(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Errorf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/object/sign/avatars"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		body := decodeJSONBody[struct {
			ExpiresIn int      `json:"expiresIn"`
			Paths     []string `json:"paths"`
		}](t, r)
		if got, want := body.ExpiresIn, 60; got != want {
			t.Errorf("expiresIn = %d, want %d", got, want)
		}
		if got, want := body.Paths[0], "a.txt"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"signedURL":"/object/sign/avatars/a.txt?token=token"}]`))
	})

	urls, err := client.From("avatars").CreateSignedURLs([]string{"a.txt"}, 60)
	if err != nil {
		t.Fatalf("CreateSignedURLs returned error: %v", err)
	}
	if got, want := len(urls), 1; got != want {
		t.Fatalf("signed URL count = %d, want %d", got, want)
	}
}

func TestFileClientInfoAndExistsUseObjectEndpoints(t *testing.T) {
	requests := make(chan string, 2)
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests <- r.Method + " " + r.URL.Path
		w.WriteHeader(http.StatusOK)
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"name":"file.txt"}`))
		}
	})

	if _, err := client.From("avatars").Info("file.txt"); err != nil {
		t.Fatalf("Info returned error: %v", err)
	}
	exists, err := client.From("avatars").Exists("file.txt")
	if err != nil {
		t.Fatalf("Exists returned error: %v", err)
	}
	if !exists {
		t.Fatal("Exists returned false, want true")
	}

	wants := []string{
		"GET /object/info/avatars/file.txt",
		"HEAD /object/avatars/file.txt",
	}
	for _, want := range wants {
		select {
		case got := <-requests:
			if got != want {
				t.Fatalf("request = %q, want %q", got, want)
			}
		default:
			t.Fatalf("missing request %q", want)
		}
	}
}

func TestFileClientListV2UsesListV2Endpoint(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Errorf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/object/list-v2/avatars"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"name":"file.txt","type":"file"}]`))
	})

	results, err := client.From("avatars").ListV2("", SearchV2Options{})
	if err != nil {
		t.Fatalf("ListV2 returned error: %v", err)
	}
	if got, want := len(results), 1; got != want {
		t.Fatalf("result count = %d, want %d", got, want)
	}
}
