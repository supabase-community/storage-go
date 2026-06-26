package storage_go

import (
	"net/http"
	"testing"
)

func TestBucketClientListBucketsUsesBucketEndpoint(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/bucket"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":"avatars","name":"avatars","public":true}]`))
	})

	buckets, err := client.Buckets().ListBuckets()
	if err != nil {
		t.Fatalf("ListBuckets returned error: %v", err)
	}
	if got, want := len(buckets), 1; got != want {
		t.Fatalf("bucket count = %d, want %d", got, want)
	}
	if got, want := buckets[0].Id, "avatars"; got != want {
		t.Fatalf("bucket id = %q, want %q", got, want)
	}
}

func TestBucketClientCreateBucketUsesBucketEndpoint(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/bucket"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		body := decodeJSONBody[map[string]any](t, r)
		if got, want := body["id"], "avatars"; got != want {
			t.Fatalf("id = %q, want %q", got, want)
		}
		if got, want := body["name"], "avatars"; got != want {
			t.Fatalf("name = %q, want %q", got, want)
		}
		if got, want := body["public"], true; got != want {
			t.Fatalf("public = %v, want %v", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"avatars","name":"avatars","public":true}`))
	})

	bucket, err := client.Buckets().CreateBucket("avatars", BucketOptions{Public: true})
	if err != nil {
		t.Fatalf("CreateBucket returned error: %v", err)
	}
	if got, want := bucket.Id, "avatars"; got != want {
		t.Fatalf("bucket id = %q, want %q", got, want)
	}
}

func TestBucketClientUpdatesEmptiesAndDeletesBucket(t *testing.T) {
	requests := make(chan string, 3)
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests <- r.Method + " " + r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"ok"}`))
	})

	if _, err := client.Buckets().UpdateBucket("avatars", BucketOptions{Public: true}); err != nil {
		t.Fatalf("UpdateBucket returned error: %v", err)
	}
	if _, err := client.Buckets().EmptyBucket("avatars"); err != nil {
		t.Fatalf("EmptyBucket returned error: %v", err)
	}
	if _, err := client.Buckets().DeleteBucket("avatars"); err != nil {
		t.Fatalf("DeleteBucket returned error: %v", err)
	}

	wants := []string{
		"PUT /bucket/avatars",
		"POST /bucket/avatars/empty",
		"DELETE /bucket/avatars",
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

func TestClientListBucketsDelegatesToBucketClient(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/bucket"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	})

	if _, err := client.ListBuckets(); err != nil {
		t.Fatalf("ListBuckets returned error: %v", err)
	}
}
