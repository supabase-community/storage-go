package test

import (
	"fmt"
	"testing"

	"github.com/supabase-community/storage-go"
)

func TestBucketListAll(t *testing.T) {
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	resp, err := c.ListBuckets()
	fmt.Println(resp, err)
}

func TestBucketFetchById(t *testing.T) {
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	fmt.Println(c.GetBucket("test"))
}

func TestBucketCreate(t *testing.T) {
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	fmt.Println(c.CreateBucket("test", storage_go.BucketOptions{
		Public: true,
	}))
}

func TestBucketUpdate(t *testing.T) {
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	_, _ = c.UpdateBucket("test", storage_go.BucketOptions{
		Public: false,
	})

	bucket, _ := c.GetBucket("test")

	if bucket.Public {
		t.Errorf("Should have been private bucket after updating")
	}
}

func TestEmptyBucket(t *testing.T) {
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	fmt.Println(c.EmptyBucket("test"))
}

func TestDeleteBucket(t *testing.T) {
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	fmt.Println(c.DeleteBucket("test"))
}
