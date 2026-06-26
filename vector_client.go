package storage_go

import (
	"encoding/json"
	"net/http"
)

type VectorBucketScope struct {
	client *Client
	bucket string
}

type VectorIndexScope struct {
	client *Client
	bucket string
	index  string
}

// From returns a vector bucket-scoped client.
func (v *VectorClient) From(bucket string) *VectorBucketScope {
	return &VectorBucketScope{client: v.client, bucket: bucket}
}

// CreateBucket creates a vector bucket.
func (v *VectorClient) CreateBucket(name string) (VectorBucketResponse, error) {
	body := map[string]any{"vectorBucketName": name}
	var response VectorBucketResponse
	err := v.postVectorAction("CreateVectorBucket", body, &response)
	return response, err
}

// GetBucket gets a vector bucket.
func (v *VectorClient) GetBucket(name string) (VectorBucketResponse, error) {
	body := map[string]any{"vectorBucketName": name}
	var response VectorBucketResponse
	err := v.postVectorAction("GetVectorBucket", body, &response)
	return response, err
}

// ListBuckets lists vector buckets.
func (v *VectorClient) ListBuckets(options ...ListVectorBucketsOptions) (ListVectorBucketsResponse, error) {
	body := any(map[string]any{})
	if len(options) > 0 {
		body = options[0]
	}
	var response ListVectorBucketsResponse
	err := v.postVectorAction("ListVectorBuckets", body, &response)
	return response, err
}

// DeleteBucket deletes a vector bucket.
func (v *VectorClient) DeleteBucket(name string) (SuccessResponse, error) {
	body := map[string]any{"vectorBucketName": name}
	var response SuccessResponse
	err := v.postVectorAction("DeleteVectorBucket", body, &response)
	return response, err
}

// CreateIndex creates an index in the scoped vector bucket.
func (v *VectorBucketScope) CreateIndex(options CreateIndexOptions) (VectorIndexResponse, error) {
	body := scopedVectorBody(options, v.bucket, "")
	var response VectorIndexResponse
	err := postVectorAction(v.client, "CreateIndex", body, &response)
	return response, err
}

// GetIndex gets an index in the scoped vector bucket.
func (v *VectorBucketScope) GetIndex(name string) (VectorIndexResponse, error) {
	body := map[string]any{
		"vectorBucketName": v.bucket,
		"vectorIndexName":  name,
	}
	var response VectorIndexResponse
	err := postVectorAction(v.client, "GetIndex", body, &response)
	return response, err
}

// ListIndexes lists indexes in the scoped vector bucket.
func (v *VectorBucketScope) ListIndexes(options ...ListIndexesOptions) (ListIndexesResponse, error) {
	body := map[string]any{"vectorBucketName": v.bucket}
	if len(options) > 0 {
		body = scopedVectorBody(options[0], v.bucket, "")
	}
	var response ListIndexesResponse
	err := postVectorAction(v.client, "ListIndexes", body, &response)
	return response, err
}

// DeleteIndex deletes an index in the scoped vector bucket.
func (v *VectorBucketScope) DeleteIndex(name string) (SuccessResponse, error) {
	body := map[string]any{
		"vectorBucketName": v.bucket,
		"vectorIndexName":  name,
	}
	var response SuccessResponse
	err := postVectorAction(v.client, "DeleteIndex", body, &response)
	return response, err
}

// Index returns an index-scoped vector data client.
func (v *VectorBucketScope) Index(name string) *VectorIndexScope {
	return &VectorIndexScope{client: v.client, bucket: v.bucket, index: name}
}

// PutVectors inserts or updates vectors in the scoped index.
func (v *VectorIndexScope) PutVectors(options PutVectorsOptions) (SuccessResponse, error) {
	body := scopedVectorBody(options, v.bucket, v.index)
	var response SuccessResponse
	err := postVectorAction(v.client, "PutVectors", body, &response)
	return response, err
}

// GetVectors retrieves vectors by id from the scoped index.
func (v *VectorIndexScope) GetVectors(options GetVectorsOptions) (GetVectorsResponse, error) {
	body := scopedVectorBody(options, v.bucket, v.index)
	var response GetVectorsResponse
	err := postVectorAction(v.client, "GetVectors", body, &response)
	return response, err
}

// ListVectors lists vectors in the scoped index.
func (v *VectorIndexScope) ListVectors(options ListVectorsOptions) (ListVectorsResponse, error) {
	body := scopedVectorBody(options, v.bucket, v.index)
	var response ListVectorsResponse
	err := postVectorAction(v.client, "ListVectors", body, &response)
	return response, err
}

// QueryVectors queries nearest vectors in the scoped index.
func (v *VectorIndexScope) QueryVectors(options QueryVectorsOptions) (QueryVectorsResponse, error) {
	body := scopedVectorBody(options, v.bucket, v.index)
	var response QueryVectorsResponse
	err := postVectorAction(v.client, "QueryVectors", body, &response)
	return response, err
}

// DeleteVectors deletes vectors from the scoped index.
func (v *VectorIndexScope) DeleteVectors(options DeleteVectorsOptions) (SuccessResponse, error) {
	body := scopedVectorBody(options, v.bucket, v.index)
	var response SuccessResponse
	err := postVectorAction(v.client, "DeleteVectors", body, &response)
	return response, err
}

func (v *VectorClient) postVectorAction(action string, body any, response any) error {
	return postVectorAction(v.client, action, body, response)
}

func postVectorAction(client *Client, action string, body any, response any) error {
	req, err := client.NewRequest(http.MethodPost, "/vector/"+action, body)
	if err != nil {
		return err
	}

	_, err = client.Do(req, response)
	return err
}

func scopedVectorBody(options any, bucket string, index string) map[string]any {
	data, err := json.Marshal(options)
	if err != nil {
		return map[string]any{}
	}

	body := map[string]any{}
	if err := json.Unmarshal(data, &body); err != nil {
		return map[string]any{}
	}
	body["vectorBucketName"] = bucket
	if index != "" {
		body["vectorIndexName"] = index
	}

	return body
}
