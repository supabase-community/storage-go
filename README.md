# Storage GO

This library is a Golang client for the [Supabase Storage API](https://supabase.com/docs/guides/storage). It's a collection of helper functions that help you manage your buckets through the API.

## Quick start

Install

```shell
go get github.com/supabase-community/storage-go
```

Usage

```go
package main

import (
	"fmt"
	"log"
	"os"

	storage_go "github.com/supabase-community/storage-go"
)

func main() {
	client := storage_go.NewClient("https://<project-reference-id>.supabase.co/storage/v1", "<project-secret-api-key>", nil)

	// Create a new bucket
	bucket, berr := client.CreateBucket("bucket-id", storage_go.BucketOptions{Public: true})

	if berr.Error != "" {
    log.Fatal("error creating bucket, ", berr)
  }

	// Upload a file
	file, err := os.Open("dummy.txt")
	if err != nil {
		panic(err)
	}

	resp := client.UploadFile("bucket-name", "file.txt", file)
	fmt.Println(resp)

	// Update Bucket
	response, berr := client.UpdateBucket(bucket.Id, storage_go.BucketOptions{Public: true})
	fmt.Println(response)

	// Empty Bucket
	response, berr = client.EmptyBucket(bucket.Id)
	fmt.Println(response)

	// Delete Bucket
	response, berr = client.DeleteBucket(bucket.Id)
	fmt.Println(response)

	// Get a bucket by its id
	bucket = GetBucket("bucket-id")
	fmt.Println(bucket)

	// Get all buckets
	fmt.Println(client.ListBuckets())

}
```

> Note to self:
> Update after tagging:
> GOPROXY=proxy.golang.org go list -m github.com/supabase-community/storage-go@v0.6.8

## License

<!-- I don't know which to use, but explicitly stating the license would be a big help -->
