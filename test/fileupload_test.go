package test

import (
	"fmt"
	"github.com/supabase-community/storage-go"
	"os"
	"testing"
)

var rawUrl = "https://abc.supabase.co/storage/v1"
var token = ""

func TestUpload(t *testing.T) {
	file, err := os.Open("dummy.txt")
	if err != nil {
		panic(err)
	}
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	resp := c.UploadFile("test1", "test.txt", file)
	fmt.Println(resp)

	//resp = c.UploadFile("test1", "hola.txt", []byte("hello world"))
	//fmt.Println(resp)
}

func TestUpdate(t *testing.T) {
	file, err := os.Open("dummy.txt")
	if err != nil {
		panic(err)
	}
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	resp := c.UpdateFile("test1", "test.txt", file)

	fmt.Println(resp)
}

func TestMoveFile(t *testing.T) {
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	resp := c.MoveFile("test1", "test.txt", "random/test.txt")

	fmt.Println(resp)
}

func TestSignedUrl(t *testing.T) {
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	resp := c.CreateSignedUrl("test1", "file_example_MP4_480_1_5MG.mp4", 120)

	fmt.Println(resp)
}

func TestPublicUrl(t *testing.T) {
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	resp := c.GetPublicUrl("shield", "book.pdf")

	fmt.Println(resp)
}

func TestDeleteFile(t *testing.T) {
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	resp := c.RemoveFile("shield", []string{"book.pdf"})

	fmt.Println(resp)
}

func TestListFile(t *testing.T) {
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
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
