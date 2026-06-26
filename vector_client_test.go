package storage_go

import (
	"net/http"
	"testing"
)

func TestVectorClientCreateBucketUsesVectorActionEndpoint(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Errorf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/vector/CreateVectorBucket"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		body := decodeJSONBody[map[string]string](t, r)
		if got, want := body["vectorBucketName"], "embeddings"; got != want {
			t.Errorf("vectorBucketName = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"vectorBucket":{"name":"embeddings"}}`))
	})

	got, err := client.Vectors().CreateBucket("embeddings")
	if err != nil {
		t.Fatalf("CreateBucket returned error: %v", err)
	}
	if got.VectorBucket.Name != "embeddings" {
		t.Fatalf("bucket name = %q, want embeddings", got.VectorBucket.Name)
	}
}

func TestVectorBucketScopeCreateIndexUsesVectorActionEndpoint(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/vector/CreateIndex"; got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
		body := decodeJSONBody[map[string]any](t, r)
		if got, want := body["vectorBucketName"], "embeddings"; got != want {
			t.Errorf("vectorBucketName = %q, want %q", got, want)
		}
		if got, want := body["name"], "docs"; got != want {
			t.Errorf("name = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"vectorIndex":{"name":"docs"}}`))
	})

	if _, err := client.Vectors().From("embeddings").CreateIndex(CreateIndexOptions{Name: "docs", Dimension: 3}); err != nil {
		t.Fatalf("CreateIndex returned error: %v", err)
	}
}

func TestVectorIndexScopePutAndQueryVectorsUseVectorActionEndpoints(t *testing.T) {
	requests := make(chan string, 2)
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests <- r.URL.Path
		body := decodeJSONBody[map[string]any](t, r)
		if got, want := body["vectorBucketName"], "embeddings"; got != want {
			t.Errorf("vectorBucketName = %q, want %q", got, want)
		}
		if got, want := body["vectorIndexName"], "docs"; got != want {
			t.Errorf("vectorIndexName = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true}`))
	})

	index := client.Vectors().From("embeddings").Index("docs")
	if _, err := index.PutVectors(PutVectorsOptions{Vectors: []VectorData{{ID: "1", Vector: []float64{1, 2, 3}}}}); err != nil {
		t.Fatalf("PutVectors returned error: %v", err)
	}
	if _, err := index.QueryVectors(QueryVectorsOptions{Vector: []float64{1, 2, 3}, TopK: 1}); err != nil {
		t.Fatalf("QueryVectors returned error: %v", err)
	}

	wants := []string{"/vector/PutVectors", "/vector/QueryVectors"}
	for _, want := range wants {
		select {
		case got := <-requests:
			if got != want {
				t.Fatalf("path = %q, want %q", got, want)
			}
		default:
			t.Fatalf("missing request %q", want)
		}
	}
}
