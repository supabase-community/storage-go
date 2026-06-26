package storage_go

// ListBuckets retrieves the details of all Storage buckets within an existing project.
func (c *Client) ListBuckets() ([]Bucket, error) {
	return c.Buckets().ListBuckets()
}

// GetBucket retrieves the details of an existing Storage bucket.
func (c *Client) GetBucket(id string) (Bucket, error) {
	return c.Buckets().GetBucket(id)
}

// CreateBucket creates a new Storage bucket.
func (c *Client) CreateBucket(id string, options BucketOptions) (Bucket, error) {
	return c.Buckets().CreateBucket(id, options)
}

// UpdateBucket updates an existing Storage bucket.
func (c *Client) UpdateBucket(id string, options BucketOptions) (MessageResponse, error) {
	return c.Buckets().UpdateBucket(id, options)
}

// EmptyBucket removes all objects inside a single bucket.
func (c *Client) EmptyBucket(id string) (MessageResponse, error) {
	return c.Buckets().EmptyBucket(id)
}

// DeleteBucket deletes an existing bucket. A bucket must be empty before it can be deleted.
func (c *Client) DeleteBucket(id string) (MessageResponse, error) {
	return c.Buckets().DeleteBucket(id)
}
