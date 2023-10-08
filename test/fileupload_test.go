package test

import (
	"fmt"
	"os"
	"testing"

	storage_go "github.com/supabase-community/storage-go"
)

var (
	rawUrl = "https://abc.supabase.co/storage/v1"
	token  = ""
)

func TestUpload(t *testing.T) {
	file, err := os.Open("dummy.txt")
	if err != nil {
		panic(err)
	}
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	resp, err := c.UploadFile("test", "test.txt", file)
	fmt.Println(resp, err)

	// resp, err = c.UploadFile("test", "hola.txt", []byte("hello world"))
	// fmt.Println(resp, err)
}

func TestUpdate(t *testing.T) {
	file, err := os.Open("dummy.txt")
	if err != nil {
		panic(err)
	}
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	resp, err := c.UpdateFile("test", "test.txt", file)

	fmt.Println(resp, err)
}

func TestMoveFile(t *testing.T) {
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	resp, err := c.MoveFile("test", "test.txt", "random/test.txt")

	fmt.Println(resp, err)
}

func TestSignedUrl(t *testing.T) {
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	resp, err := c.CreateSignedUrl("test", "file_example_MP4_480_1_5MG.mp4", 120)

	fmt.Println(resp, err)
}

func TestPublicUrl(t *testing.T) {
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	resp := c.GetPublicUrl("shield", "book.pdf")

	fmt.Println(resp)
}

func TestDeleteFile(t *testing.T) {
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	resp, err := c.RemoveFile("shield", []string{"book.pdf"})

	fmt.Println(resp, err)
}

func TestListFile(t *testing.T) {
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	resp, err := c.ListFiles("shield", "", storage_go.FileSearchOptions{
		Limit:  10,
		Offset: 0,
		SortByOptions: storage_go.SortBy{
			Column: "",
			Order:  "",
		},
	})

	fmt.Println(resp, err)
}

func TestCreateUploadSignedUrl(t *testing.T) {
	c := storage_go.NewClient(rawUrl, token, map[string]string{"apiKey": token})
	resp, err := c.CreateSignedUploadUrl("your-bucket-id", "book.pdf")

	fmt.Println(resp, err)
}

func TestUploadToSignedUrl(t *testing.T) {
	c := storage_go.NewClient(rawUrl, token, map[string]string{"apiKey": token})
	file, err := os.Open("dummy.txt")
	if err != nil {
		panic(err)
	}
	// resp, err := c.CreateSignedUploadUrl("test", "vu.txt")
	res, err := c.UploadToSignedUrl("your-response-url", file)
	fmt.Println(res, err)
}

func TestDownloadFile(t *testing.T) {
	c := storage_go.NewClient(rawUrl, token, map[string]string{})
	resp, err := c.DownloadFile("test", "book.pdf")
	if err != nil {
		t.Fatalf("DownloadFile failed: %v", err)
	}

	err = os.WriteFile("book.pdf", resp, 0644)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
}
