# Storage GO
Golang client for [Supabase Storage API](https://github.com/supabase/storage-api)

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
	"os"

	"github.com/supabase-community/storage-go"
)

func main() {
	client := storage_go.NewClient("https://abc.supabase.co/storage/v1", "<service-token>", nil)

	// Get buckets
	fmt.Println(client.ListBuckets())

	// Upload a file

	file, err := os.Open("dummy.txt")
	if err != nil {
		panic(err)
	}

	resp := client.UploadFile("bucket-name", "file.txt", file)
	fmt.Println(resp)
}
```
