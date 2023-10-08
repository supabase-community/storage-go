package storage_go

import (
	"encoding/json"
	"net/http"
)

// ListBuckets retrieves the details of all Storage buckets within an existing project.
func (c *Client) ListBuckets() ([]Bucket, error) {
	bucketsURL := c.clientTransport.baseUrl.String() + "/bucket"
	req, err := c.NewRequest(http.MethodGet, bucketsURL, nil)
	if err != nil {
		return nil, err
	}

	var buckets []Bucket
	_, err = c.Do(req, &buckets)
	if err != nil {
		return nil, err
	}

	return buckets, nil
}

// GetBucket retrieves the details of an existing Storage bucket.
func (c *Client) GetBucket(id string) (Bucket, error) {
	bucketURL := c.clientTransport.baseUrl.String() + "/bucket/" + id
	req, err := c.NewRequest(http.MethodGet, bucketURL, nil)
	if err != nil {
		return Bucket{}, err
	}

	var bucket Bucket
	_, err = c.Do(req, &bucket)
	if err != nil {
		return Bucket{}, err
	}

	return bucket, nil
}

// CreateBucket creates a new Storage bucket
// options.public The visibility of the bucket. Public buckets don't require an authorization token to download objects, but still require a valid token for all other operations. By default, buckets are private.
// options.fileSizeLimit The maximum file size in bytes allowed in the bucket. By default, there is no limit.
// options.allowedMimeTypes The list of allowed MIME types. By default, all MIME types are allowed.
// return newly created bucket id
func (c *Client) CreateBucket(id string, options BucketOptions) (Bucket, error) {
	createBucketURL := c.clientTransport.baseUrl.String() + "/bucket"
	bodyData := map[string]interface{}{
		"id":     id,
		"name":   id,
		"public": options.Public,
	}
	// We only set the file size limit if it's not empty
	if len(options.FileSizeLimit) > 0 {
		bodyData["file_size_limit"] = options.FileSizeLimit
	}
	// We only set the allowed mime types if it's not empty
	if len(options.AllowedMimeTypes) > 0 {
		bodyData["allowed_mime_types"] = options.AllowedMimeTypes
	}
	// jsonBody, _ := json.Marshal(bodyData)
	req, err := c.NewRequest(http.MethodPost, createBucketURL, &bodyData)
	if err != nil {
		return Bucket{}, err
	}

	var bucket Bucket
	_, err = c.Do(req, &bucket)
	if err != nil {
		return Bucket{}, err
	}

	return bucket, nil
}

// UpdateBucket creates a new Storage bucket
// options.public The visibility of the bucket. Public buckets don't require an authorization token to download objects, but still require a valid token for all other operations. By default, buckets are private.
// options.fileSizeLimit The maximum file size in bytes allowed in the bucket. By default, there is no limit.
// options.allowedMimeTypes The list of allowed MIME types. By default, all MIME types are allowed.
// return newly updated bucket id
func (c *Client) UpdateBucket(id string, options BucketOptions) (MessageResponse, error) {
	bucketURL := c.clientTransport.baseUrl.String() + "/bucket/" + id
	bodyData := map[string]interface{}{
		"id":     id,
		"name":   id,
		"public": options.Public,
	}
	// We only set the file size limit if it's not empty
	if len(options.FileSizeLimit) > 0 {
		bodyData["file_size_limit"] = options.FileSizeLimit
	}
	// We only set the allowed mime types if it's not empty
	if len(options.AllowedMimeTypes) > 0 {
		bodyData["allowed_mime_types"] = options.AllowedMimeTypes
	}
	req, err := c.NewRequest(http.MethodPut, bucketURL, &bodyData)
	if err != nil {
		return MessageResponse{}, err
	}
	var message MessageResponse
	_, err = c.Do(req, &message)
	if err != nil {
		return MessageResponse{}, err
	}

	return message, nil
}

// EmptyBucket removes all objects inside a single bucket.
func (c *Client) EmptyBucket(id string) (MessageResponse, error) {
	bucketURL := c.clientTransport.baseUrl.String() + "/bucket/" + id + "/empty"
	jsonBody, _ := json.Marshal(map[string]interface{}{})
	req, err := c.NewRequest(http.MethodPost, bucketURL, &jsonBody)
	if err != nil {
		return MessageResponse{}, err
	}

	var message MessageResponse
	_, err = c.Do(req, &message)

	return message, err
}

// DeleteBucket deletes an existing bucket. A bucket must be empty before it can be deleted.
func (c *Client) DeleteBucket(id string) (MessageResponse, error) {
	bucketURL := c.clientTransport.baseUrl.String() + "/bucket/" + id
	jsonBody, _ := json.Marshal(map[string]interface{}{})
	req, err := c.NewRequest(http.MethodDelete, bucketURL, &jsonBody)
	if err != nil {
		return MessageResponse{}, err
	}

	var message MessageResponse
	_, err = c.Do(req, &message)

	return message, err
}
