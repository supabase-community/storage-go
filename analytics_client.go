package storage_go

import (
	"net/http"
	"net/url"
)

// CreateBucket creates an analytics bucket.
func (a *AnalyticsClient) CreateBucket(name string) (AnalyticBucket, error) {
	options := BucketOptions{Type: BucketTypeAnalytics}
	req, err := a.client.NewRequest(http.MethodPost, "/bucket", bucketPayload(name, options))
	if err != nil {
		return AnalyticBucket{}, err
	}

	var bucket AnalyticBucket
	if _, err := a.client.Do(req, &bucket); err != nil {
		return AnalyticBucket{}, err
	}

	return bucket, nil
}

// ListBuckets lists analytics buckets.
func (a *AnalyticsClient) ListBuckets(options ...ListBucketOptions) ([]AnalyticBucket, error) {
	path := analyticsBucketListPath(options...)
	req, err := a.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var buckets []AnalyticBucket
	if _, err := a.client.Do(req, &buckets); err != nil {
		return nil, err
	}

	return buckets, nil
}

// DeleteBucket deletes an analytics bucket.
func (a *AnalyticsClient) DeleteBucket(name string) error {
	req, err := a.client.NewRequest(http.MethodDelete, "/bucket/"+url.PathEscape(name), nil)
	if err != nil {
		return err
	}

	var message MessageResponse
	_, err = a.client.Do(req, &message)
	return err
}

// From returns a thin Iceberg REST catalog client scoped to bucket.
func (a *AnalyticsClient) From(bucket string) *IcebergCatalogClient {
	return &IcebergCatalogClient{client: a.client, bucket: bucket}
}

type IcebergCatalogClient struct {
	client *Client
	bucket string
}

// Do sends a low-level Iceberg catalog request for the scoped analytics bucket.
func (i *IcebergCatalogClient) Do(method string, path string, body any, v any) error {
	if path == "" || path[0] != '/' {
		path = "/" + path
	}
	req, err := i.client.NewRequest(method, "/iceberg/"+url.PathEscape(i.bucket)+path, body)
	if err != nil {
		return err
	}

	_, err = i.client.Do(req, v)
	return err
}

func analyticsBucketListPath(options ...ListBucketOptions) string {
	path := bucketListPath(options...)
	parsed, err := url.Parse(path)
	if err != nil {
		return "/bucket?type=ANALYTICS"
	}
	values := parsed.Query()
	values.Set("type", string(BucketTypeAnalytics))
	parsed.RawQuery = values.Encode()

	return parsed.String()
}
