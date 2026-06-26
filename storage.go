package storage_go

import (
	"io"
	"net/url"
	"strconv"
)

const (
	defaultLimit            = 100
	defaultOffset           = 0
	defaultFileCacheControl = "3600"
	defaultFileContentType  = "text/plain;charset=UTF-8"
	defaultFileUpsert       = false
	defaultSortColumn       = "name"
	defaultSortOrder        = "asc"
)

func (c *Client) UploadOrUpdateFile(
	bucketId string,
	relativePath string,
	data io.Reader,
	update bool,
	options ...FileOptions,
) (FileUploadResponse, error) {
	if update {
		return c.From(bucketId).Update(relativePath, data, options...)
	}

	return c.From(bucketId).Upload(relativePath, data, options...)
}

// UpdateFile will replace an existing file at the specified path.
// bucketId string The bucket id
// relativePath path The file path, including the file name. Should be of the format `folder/subfolder/filename.png`
// data io.Reader The file data
func (c *Client) UpdateFile(bucketId string, relativePath string, data io.Reader, fileOptions ...FileOptions) (FileUploadResponse, error) {
	return c.From(bucketId).Update(relativePath, data, fileOptions...)
}

// UploadFile will upload file to an existing bucket at the specified path.
// bucketId string The bucket id
// relativePath path The file path, including the file name. Should be of the format `folder/subfolder/filename.png`
// data io.Reader The file data
func (c *Client) UploadFile(bucketId string, relativePath string, data io.Reader, fileOptions ...FileOptions) (FileUploadResponse, error) {
	return c.From(bucketId).Upload(relativePath, data, fileOptions...)
}

// MoveFile will move an existing file to new path in the same bucket.
// bucketId string The bucket id
// sourceKey path The file path, including the file name. Should be of the format `folder/subfolder/filename.png`
// destinationKey path The file path, including the file name. Should be of the format `folder/subfolder/new-filename.png`
func (c *Client) MoveFile(bucketId string, sourceKey string, destinationKey string) (FileUploadResponse, error) {
	return c.From(bucketId).Move(sourceKey, destinationKey)
}

// CreateSignedUrl create a signed URL. Use a signed URL to share a file for a fixed amount of time.
// bucketId string The bucket id
// filePath path The file path, including the file name. Should be of the format `folder/subfolder/filename.png`
// expiresIn int The number of seconds before the signed URL expires. Defaults to 60 seconds.
func (c *Client) CreateSignedUrl(bucketId string, filePath string, expiresIn int) (SignedUrlResponse, error) {
	return c.From(bucketId).CreateSignedURL(filePath, expiresIn)
}

// CreateSignedUploadUrl create a signed URL for uploading a file. Use a signed URL to upload a file directly to a bucket.
// bucketId string The bucket id
// filePath path The file path, including the file name. Should be of the format `folder/subfolder/filename.png`
func (c *Client) CreateSignedUploadUrl(bucketId string, filePath string) (SignedUploadUrlResponse, error) {
	return c.From(bucketId).CreateSignedUploadURL(filePath)
}

// UploadToSignedUrl upload a file to a signed URL.
// filePath string The file path, including the file name. Should be of the format `folder/subfolder/filename.png`
// fileBody io.Reader The file data
func (c *Client) UploadToSignedUrl(filePath string, fileBody io.Reader) (*UploadToSignedUrlResponse, error) {
	response, err := c.From("").UploadToSignedURL(filePath, fileBody)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetPublicUrl use to to get the URL for an asset in a public bucket. If you do not want to use this function, you can construct the public URL by concatenating the bucket URL with the path to the asset.
// bucketId string The bucket id
// filePath path The file path, including the file name. Should be of the format `folder/subfolder/filename.png`
// urlOptions UrlOptions The URL options
func (c *Client) GetPublicUrl(bucketId string, filePath string, urlOptions ...UrlOptions) SignedUrlResponse {
	return c.From(bucketId).PublicURL(filePath, urlOptions...)
}

// RemoveFile remove a file from an existing bucket.
// bucketId string The bucket id.
// paths []string The file paths, including the file name. Should be of the format `folder/subfolder/filename.png`
func (c *Client) RemoveFile(bucketId string, paths []string) ([]FileUploadResponse, error) {
	return c.From(bucketId).Remove(paths)
}

// ListFiles list files in an existing bucket.
// bucketId string The bucket id.
// queryPath string The file path, including the file name. Should be of the format `folder/subfolder/filename.png`
// options FileSearchOptions The file search options
func (c *Client) ListFiles(bucketId string, queryPath string, options FileSearchOptions) ([]FileObject, error) {
	return c.From(bucketId).List(queryPath, options)
}

// DownloadFile download a file from an existing bucket.
// bucketId string The bucket id.
// filePath string The file path, including the file name. Should be of the format `folder/subfolder/filename.png`
// urlOptions UrlOptions The URL options
func (c *Client) DownloadFile(bucketId string, filePath string, urlOptions ...UrlOptions) ([]byte, error) {
	return c.From(bucketId).Download(filePath, urlOptions...)
}

// buildUrlWithOption will base on current url and option to build a new url
func buildUrlWithOption(urlStr string, options UrlOptions) string {
	signedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	signedURLQuery := signedURL.Query()

	if options.Transform != nil {
		if options.Transform.Height > 0 {
			signedURLQuery.Add("height", strconv.Itoa(options.Transform.Height))
		}
		if options.Transform.Width > 0 {
			signedURLQuery.Add("width", strconv.Itoa(options.Transform.Width))
		}
		// Default: origin
		if options.Transform.Format != "" {
			signedURLQuery.Add("format", options.Transform.Format)
		}
		// Default: 80
		if options.Transform.Quality > 0 {
			signedURLQuery.Add("quality", strconv.Itoa(options.Transform.Quality))
		}
		if options.Transform.Resize != "" && (options.Transform.Resize == "conver" || options.Transform.Resize == "contain" || options.Transform.Resize == "fill") {
			signedURLQuery.Add("resize", options.Transform.Resize)
		}
	}
	// Default on server is false
	if options.Download {
		signedURLQuery.Add("download", strconv.FormatBool(options.Download))
	}

	signedURL.RawQuery = signedURLQuery.Encode()
	return signedURL.String()
}
