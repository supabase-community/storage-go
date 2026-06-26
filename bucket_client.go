package storage_go

import (
	"net/http"
	"net/url"
	"strconv"
)

// ListBuckets retrieves all storage buckets for the project.
func (b *BucketClient) ListBuckets(options ...ListBucketOptions) ([]Bucket, error) {
	req, err := b.client.NewRequest(http.MethodGet, bucketListPath(options...), nil)
	if err != nil {
		return nil, err
	}

	var buckets []Bucket
	if _, err := b.client.Do(req, &buckets); err != nil {
		return nil, err
	}

	return buckets, nil
}

// GetBucket retrieves a storage bucket by id.
func (b *BucketClient) GetBucket(id string) (Bucket, error) {
	req, err := b.client.NewRequest(http.MethodGet, "/bucket/"+url.PathEscape(id), nil)
	if err != nil {
		return Bucket{}, err
	}

	var bucket Bucket
	if _, err := b.client.Do(req, &bucket); err != nil {
		return Bucket{}, err
	}

	return bucket, nil
}

// CreateBucket creates a storage bucket.
func (b *BucketClient) CreateBucket(id string, options BucketOptions) (Bucket, error) {
	req, err := b.client.NewRequest(http.MethodPost, "/bucket", bucketPayload(id, options))
	if err != nil {
		return Bucket{}, err
	}

	var bucket Bucket
	if _, err := b.client.Do(req, &bucket); err != nil {
		return Bucket{}, err
	}

	return bucket, nil
}

// UpdateBucket updates a storage bucket.
func (b *BucketClient) UpdateBucket(id string, options BucketOptions) (MessageResponse, error) {
	req, err := b.client.NewRequest(http.MethodPut, "/bucket/"+url.PathEscape(id), bucketPayload(id, options))
	if err != nil {
		return MessageResponse{}, err
	}

	var message MessageResponse
	if _, err := b.client.Do(req, &message); err != nil {
		return MessageResponse{}, err
	}

	return message, nil
}

// EmptyBucket removes every object from a storage bucket.
func (b *BucketClient) EmptyBucket(id string) (MessageResponse, error) {
	req, err := b.client.NewRequest(http.MethodPost, "/bucket/"+url.PathEscape(id)+"/empty", nil)
	if err != nil {
		return MessageResponse{}, err
	}

	var message MessageResponse
	if _, err := b.client.Do(req, &message); err != nil {
		return MessageResponse{}, err
	}

	return message, nil
}

// DeleteBucket deletes an empty storage bucket.
func (b *BucketClient) DeleteBucket(id string) (MessageResponse, error) {
	req, err := b.client.NewRequest(http.MethodDelete, "/bucket/"+url.PathEscape(id), nil)
	if err != nil {
		return MessageResponse{}, err
	}

	var message MessageResponse
	if _, err := b.client.Do(req, &message); err != nil {
		return MessageResponse{}, err
	}

	return message, nil
}

func bucketPayload(id string, options BucketOptions) map[string]any {
	body := map[string]any{
		"id":     id,
		"name":   id,
		"public": options.Public,
	}
	if options.Type != "" {
		body["type"] = options.Type
	}
	if options.FileSizeLimit != "" {
		body["file_size_limit"] = options.FileSizeLimit
	}
	if len(options.AllowedMimeTypes) > 0 {
		body["allowed_mime_types"] = options.AllowedMimeTypes
	}

	return body
}

func bucketListPath(options ...ListBucketOptions) string {
	if len(options) == 0 {
		return "/bucket"
	}

	values := url.Values{}
	option := options[0]
	if option.Limit > 0 {
		values.Set("limit", strconv.Itoa(option.Limit))
	}
	if option.Offset > 0 {
		values.Set("offset", strconv.Itoa(option.Offset))
	}
	if option.Search != "" {
		values.Set("search", option.Search)
	}
	if option.SortBy.Column != "" {
		values.Set("sortBy", option.SortBy.Column)
	}
	if option.SortBy.Order != "" {
		values.Set("order", option.SortBy.Order)
	}
	if len(values) == 0 {
		return "/bucket"
	}

	return "/bucket?" + values.Encode()
}
