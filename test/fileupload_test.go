package test

import (
	"fmt"
	storage_go "github.com/supabase/storage-go"
	"testing"
)

var rawUrl = "https://xyz.supabase.co/storage/v1"

func TestUpload(t *testing.T) {
	c := storage_go.NewClient(rawUrl, "", map[string]string{})
	resp := c.UploadFile("test1", "test.txt", []byte("hello world"))
	fmt.Println(resp)

	resp = c.UploadFile("test1", "hola.txt", []byte("hello world"))
	fmt.Println(resp)
}

func TestUpdate(t *testing.T) {
	c := storage_go.NewClient(rawUrl, "", map[string]string{})
	resp := c.UpdateFile("test1", "test.txt", []byte("hello updated world"))

	fmt.Println(resp)
}

func TestMoveFile(t *testing.T) {
	c := storage_go.NewClient(rawUrl, "", map[string]string{})
	resp := c.MoveFile("test1", "test.txt", "random/test.txt")

	fmt.Println(resp)
}

func TestSignedUrl(t *testing.T) {
	c := storage_go.NewClient(rawUrl, "", map[string]string{})
	resp := c.CreateSignedUrl("test1", "file_example_MP4_480_1_5MG.mp4", 120)

	fmt.Println(resp)
}

func TestPublicUrl(t *testing.T) {
	c := storage_go.NewClient(rawUrl, "", map[string]string{})
	resp := c.GetPublicUrl("shield", "book.pdf")

	fmt.Println(resp)
}

func TestDeleteFile(t *testing.T) {
	c := storage_go.NewClient(rawUrl, "", map[string]string{})
	resp := c.RemoveFile("shield", []string{"book.pdf"})

	fmt.Println(resp)
}

func TestListFile(t *testing.T) {
	c := storage_go.NewClient(rawUrl, "", map[string]string{})
	resp := c.ListFiles("test1", "", storage_go.FileSearchOptions{
		Limit:  10,
		Offset: 0,
		SortByOptions: storage_go.SortBy{
			Column: "",
			Order:  "",
		},
	})

	fmt.Println(resp)
}
