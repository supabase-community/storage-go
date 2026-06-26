package storage_go

import (
	"net/http"
	"testing"
)

func TestAnalyticsClientCreateBucketUsesAnalyticsBucketType(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Errorf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/bucket"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		body := decodeJSONBody[map[string]any](t, r)
		if got, want := body["type"], "ANALYTICS"; got != want {
			t.Errorf("type = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"events","name":"events","type":"ANALYTICS"}`))
	})

	bucket, err := client.Analytics().CreateBucket("events")
	if err != nil {
		t.Fatalf("CreateBucket returned error: %v", err)
	}
	if got, want := bucket.Type, BucketTypeAnalytics; got != want {
		t.Fatalf("type = %q, want %q", got, want)
	}
}

func TestAnalyticsClientListAndDeleteBuckets(t *testing.T) {
	requests := make(chan string, 2)
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests <- r.Method + " " + r.URL.Path + "?" + r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		if r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`[{"id":"events","name":"events","type":"ANALYTICS"}]`))
			return
		}
		_, _ = w.Write([]byte(`{"message":"ok"}`))
	})

	if _, err := client.Analytics().ListBuckets(); err != nil {
		t.Fatalf("ListBuckets returned error: %v", err)
	}
	if err := client.Analytics().DeleteBucket("events"); err != nil {
		t.Fatalf("DeleteBucket returned error: %v", err)
	}

	wants := []string{
		"GET /bucket?type=ANALYTICS",
		"DELETE /bucket/events?",
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

func TestAnalyticsClientFromReturnsCatalogClient(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	catalog := client.Analytics().From("events")
	if catalog == nil {
		t.Fatal("From returned nil")
	}
	if catalog.bucket != "events" {
		t.Fatalf("bucket = %q, want %q", catalog.bucket, "events")
	}
}
