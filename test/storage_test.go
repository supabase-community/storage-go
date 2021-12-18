package test

import (
	"fmt"
	storage_go "storage-go"
	"testing"
)

func TestBucketListAll(t *testing.T) {
	c := storage_go.NewClient("https://abc.supabase.co/storage/v1", "", map[string]string{})
	c.ListBuckets()
}

func TestBucketFetchById(t *testing.T) {
	c := storage_go.NewClient("https://abc.supabase.co/storage/v1", "", map[string]string{})
	fmt.Println(c.GetBucket("shield"))
}

func TestBucketCreate(t *testing.T) {
	c := storage_go.NewClient("https://abc.supabase.co/storage/v1", "", map[string]string{})
	fmt.Println(c.CreateBucket("test1", storage_go.BucketOptions{
		Public: true,
	}))
}

func TestBucketUpdate(t *testing.T) {
	c := storage_go.NewClient("https://abc.supabase.co/storage/v1", "", map[string]string{})
	c.UpdateBucket("test1", storage_go.BucketOptions{
		Public: false,
	})

	bucket, _ := c.GetBucket("test1")

	if bucket.Public {
		t.Errorf("Should have been private bucket after updating")
	}
}
