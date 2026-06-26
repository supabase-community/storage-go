# Storage GO

This library is a Golang client for the [Supabase Storage API](https://supabase.com/docs/guides/storage). It's a collection of helper functions that help you manage your buckets through the API.

## Quick start guide

Requires Go 1.22 or newer.

#### Install

```shell
go get github.com/supabase-community/storage-go
```

### Connecting to the storage backend

```go
package main

import (
	"fmt"
	"log"
	"strings"

	storage_go "github.com/supabase-community/storage-go"
)

func main() {
	storageClient := storage_go.NewClient("https://<project-reference-id>.supabase.co/storage/v1", "<project-secret-api-key>", nil)

	fileBody := strings.NewReader("hello")
	uploaded, err := storageClient.From("avatars").Upload("user.txt", fileBody)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(uploaded.Path)
}
```

Existing flat methods such as `UploadFile`, `ListBuckets`, and `CreateSignedUrl` remain available for compatibility. New code should prefer scoped clients such as `client.From("bucket")`, `client.Buckets()`, `client.Analytics()`, and `client.Vectors()`.

### Handling resources

#### Handling Storage Buckets

- Create a new Storage bucket:

```go
  result, err := storageClient.Buckets().CreateBucket("bucket-id", storage_go.BucketOptions{
    Public: true,
  })
```

- Retrieve the details of an existing Storage bucket:

```go
  result, err := storageClient.Buckets().GetBucket("bucket-id")
```

- Update a new Storage bucket:

```go
  result, err := storageClient.Buckets().UpdateBucket("bucket-id", storage_go.BucketOptions{
    Public: true,
  })
```

- Remove all objects inside a single bucket:

```go
  result, err := storageClient.Buckets().EmptyBucket("bucket-id")
```

- Delete an existing bucket (a bucket can't be deleted with existing objects inside it):

```go
  result, err := storageClient.Buckets().DeleteBucket("bucket-id")
```

- Retrieve the details of all Storage buckets within an existing project:

```go
  result, err := storageClient.Buckets().ListBuckets()
```

#### Handling Files

```go
  fileBody := ... // load your file here

  result, err := storageClient.From("test").Upload("test.txt", fileBody)
```

> Note: The `upload` method also accepts a map of optional parameters.

- Download a file from an exisiting bucket:

```go
  result, err := storageClient.From("bucket-id").Download("test.txt")
```

- List all the files within a bucket:

```go
  result, err := storageClient.From("bucket-id").List("", storage_go.FileSearchOptions{
      Limit:  10,
      Offset: 0,
      SortByOptions: storage_go.SortBy{
      Column: "",
      Order:  "",
    },
  })
```

> Note: The `list` method also accepts a map of optional parameters.

- Replace an existing file at the specified path with a new one:

```go
  fileBody := ... // load your file here

  result, err := storageClient.From("test").Update("test.txt", file)
```

- Move an existing file:

```go
  result, err := storageClient.From("test").Move("test.txt", "random/test.txt")
```

- Delete files within the same bucket:

```go
  result, err := storageClient.From("test").Remove([]string{"book.pdf"})
```

- Create signed URL to download file without requiring permissions:

```go
  const expireIn = 60

  result, err := storageClient.From("test").CreateSignedURL("test.mp4", expireIn)
```

- Retrieve URLs for assets in public buckets:

```go
  result := storageClient.From("test").PublicURL("book.pdf")
```

- Create an signed URL and upload to signed URL:

```go
  fileBody := ... // load your file here

  resp, err := storageClient.From("test").CreateSignedUploadURL("test.txt")
  res, err := storageClient.From("test").UploadToSignedURL(resp.Url, file)
```

#### Handling Analytics Buckets

```go
  bucket, err := storageClient.Analytics().CreateBucket("events")
  buckets, err := storageClient.Analytics().ListBuckets()
  err = storageClient.Analytics().DeleteBucket("events")
  catalog := storageClient.Analytics().From(bucket.Id)
  _ = catalog
  _ = buckets
```

#### Handling Vector Storage

```go
  vectorBucket, err := storageClient.Vectors().CreateBucket("embeddings")
  index, err := storageClient.Vectors().From("embeddings").CreateIndex(storage_go.CreateIndexOptions{
    Name:      "documents",
    Dimension: 1536,
  })
  matches, err := storageClient.Vectors().From("embeddings").Index("documents").QueryVectors(storage_go.QueryVectorsOptions{
    Vector: []float64{0.1, 0.2, 0.3},
    TopK:   10,
  })
  _ = vectorBucket
  _ = index
  _ = matches
```

## License

<!-- I don't know which to use, but explicitly stating the license would be a big help -->
