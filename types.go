package storage_go

type MessageResponse struct {
	Message string `json:"message"`
}

type BucketResponseError struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	StatusCode uint16 `json:"statusCode"`
}

type Bucket struct {
	Id               string   `json:"id"`
	Name             string   `json:"name"`
	Owner            string   `json:"owner"`
	Public           bool     `json:"public"`
	FileSizeLimit    *int64   `json:"file_size_limit"`
	AllowedMimeTypes []string `json:"allowed_mine_types"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
}

// BucketOptions is used to create or update a Bucket with option
type BucketOptions struct {
	Public           bool
	FileSizeLimit    string
	AllowedMimeTypes []string
}

type SortBy struct {
	Column string `json:"column"`
	Order  string `json:"order"`
}

type FileUploadResponse struct {
	Key     string `json:"Key"`
	Message string `json:"message"`
	Data    []byte
	Code    string `json:"statusCode"`
	Error   string `json:"error"`
}

type SignedUrlResponse struct {
	SignedURL string `json:"signedURL"`
}

type FileSearchOptions struct {
	Limit         int    `json:"limit"`
	Offset        int    `json:"offset"`
	SortByOptions SortBy `json:"sortBy"`
}

type FileObject struct {
	Name           string      `json:"name"`
	BucketId       string      `json:"bucket_id"`
	Owner          string      `json:"owner"`
	Id             string      `json:"id"`
	UpdatedAt      string      `json:"updated_at"`
	CreatedAt      string      `json:"created_at"`
	LastAccessedAt string      `json:"last_accessed_at"`
	Metadata       interface{} `json:"metadata"`
	Buckets        Bucket
}

type ListFileRequestBody struct {
	Limit         int    `json:"limit"`
	Offset        int    `json:"offset"`
	SortByOptions SortBy `json:"sortBy"`
	Prefix        string `json:"prefix"`
}

type TransformOptions struct {
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Resize  string `json:"resize"`
	Format  string `json:"format"`
	Quality int    `json:"quality"`
}

type UrlOptions struct {
	Transform *TransformOptions `json:"transform"`
	Download  bool              `json:"download"`
}

type SignedUploadUrlResponse struct {
	Url string `json:"url"`
}

type UploadToSignedUrlResponse struct {
	Key string `json:"key"`
}

type FileOptions struct {
	// The number of seconds the asset is cached in the browser and in the Supabase CDN.
	// This is set in the `Cache-Control: max-age=<seconds>` header. Defaults to 3600 seconds.
	CacheControl *string
	// The `Content-Type` header value. Should be specified if using a `fileBody` that is neither `Blob` nor `File` nor `FormData`, otherwise will default to `text/plain;charset=UTF-8`.
	ContentType *string
	// The duplex option is a string parameter that enables or disables duplex streaming, allowing for both reading and writing data in the same stream. It can be passed as an option to the fetch() method.
	Duplex *string
	// When upsert is set to true, the file is overwritten if it exists. When set to false, an error is thrown if the object already exists.
	// Defaults to false.
	Upsert *bool
}
