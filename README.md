# Storage GO
Golang client for [Supabase Storage API](https://github.com/supabase/storage-api)

## Quick start
Install
```shell
go get github.com/supabase/storage-go
```

Usage
```go
package main

import (
	"fmt"

	"github.com/supabase/storage-go"
)

func main() {
	client := storage_go.NewClient("https://abc.supabase.co/storage/v1", "<service-token>", nil)
	
	// Get buckets
	fmt.Println(client.ListBuckets())
	
	// Upload a file
	resp := client.UploadFile("bucket-name", "file.txt", []byte("hello world"))
	fmt.Println(resp)
}
```