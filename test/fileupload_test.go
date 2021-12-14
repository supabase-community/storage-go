package test

import (
	"fmt"
	storage_go "storage-go"
	"testing"
)

func TestUpload(t *testing.T) {
	c := storage_go.NewClient("https://abc.supabase.co/storage/v1", map[string]string{})
	resp := c.Upload("test1", "test.txt", []byte("hello world"))

	fmt.Println(resp)
}
