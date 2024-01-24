package storage_go

import (
	"bufio"
	"io"
	"net/http"
	"net/url"
	"regexp"
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
	path := removeEmptyFolderName(bucketId + "/" + relativePath)
	uploadURL := c.clientTransport.baseUrl.String() + "/object/" + path

	// Check on file options
	if len(options) > 0 {
		if options[0].CacheControl != nil {
			c.clientTransport.header.Set("cache-control", *options[0].CacheControl)
		}
		if options[0].ContentType != nil {
			c.clientTransport.header.Set("content-type", *options[0].ContentType)
		}
		if options[0].Upsert != nil {
			c.clientTransport.header.Set("x-upsert", strconv.FormatBool(*options[0].Upsert))
		}
	}
	method := http.MethodPost
	if update {
		method = http.MethodPut
	}
	bodyData := bufio.NewReader(data)
	req, err := http.NewRequest(method, uploadURL, bodyData)
	if err != nil {
		return FileUploadResponse{}, err
	}

	var response FileUploadResponse
	_, err = c.Do(req, &response)

	// set content-type back to default after request
	c.clientTransport.header.Set("content-type", "application/json")
	
	if err != nil {
		return FileUploadResponse{}, err
	}

	return response, nil
}

// UpdateFile will replace an existing file at the specified path.
// bucketId string The bucket id
// relativePath path The file path, including the file name. Should be of the format `folder/subfolder/filename.png`
// data io.Reader The file data
func (c *Client) UpdateFile(bucketId string, relativePath string, data io.Reader, fileOptions ...FileOptions) (FileUploadResponse, error) {
	return c.UploadOrUpdateFile(bucketId, relativePath, data, true, fileOptions...)
}

// UploadFile will upload file to an existing bucket at the specified path.
// bucketId string The bucket id
// relativePath path The file path, including the file name. Should be of the format `folder/subfolder/filename.png`
// data io.Reader The file data
func (c *Client) UploadFile(bucketId string, relativePath string, data io.Reader, fileOptions ...FileOptions) (FileUploadResponse, error) {
	return c.UploadOrUpdateFile(bucketId, relativePath, data, false, fileOptions...)
}

// MoveFile will move an existing file to new path in the same bucket.
// bucketId string The bucket id
// sourceKey path The file path, including the file name. Should be of the format `folder/subfolder/filename.png`
// destinationKey path The file path, including the file name. Should be of the format `folder/subfolder/new-filename.png`
func (c *Client) MoveFile(bucketId string, sourceKey string, destinationKey string) (FileUploadResponse, error) {
	jsonBody := map[string]interface{}{
		"bucketId":       bucketId,
		"sourceKey":      sourceKey,
		"destinationKey": destinationKey,
	}

	moveURL := c.clientTransport.baseUrl.String() + "/object/move"
	req, err := c.NewRequest(http.MethodPost, moveURL, &jsonBody)
	if err != nil {
		return FileUploadResponse{}, err
	}

	var response FileUploadResponse
	_, err = c.Do(req, &response)
	if err != nil {
		return FileUploadResponse{}, err
	}

	return response, err
}

// CreateSignedUrl create a signed URL. Use a signed URL to share a file for a fixed amount of time.
// bucketId string The bucket id
// filePath path The file path, including the file name. Should be of the format `folder/subfolder/filename.png`
// expiresIn int The number of seconds before the signed URL expires. Defaults to 60 seconds.
func (c *Client) CreateSignedUrl(bucketId string, filePath string, expiresIn int) (SignedUrlResponse, error) {
	signedURL := c.clientTransport.baseUrl.String() + "/object/sign/" + bucketId + "/" + filePath
	jsonBody := map[string]interface{}{
		"expiresIn": expiresIn,
	}

	req, err := c.NewRequest(http.MethodPost, signedURL, &jsonBody)
	if err != nil {
		return SignedUrlResponse{}, err
	}

	var response SignedUrlResponse
	_, err = c.Do(req, &response)
	if err != nil {
		return SignedUrlResponse{}, err
	}

	response.SignedURL = c.clientTransport.baseUrl.String() + response.SignedURL

	return response, nil
}

// CreateSignedUploadUrl create a signed URL for uploading a file. Use a signed URL to upload a file directly to a bucket.
// bucketId string The bucket id
// filePath path The file path, including the file name. Should be of the format `folder/subfolder/filename.png`
func (c *Client) CreateSignedUploadUrl(bucketId string, filePath string) (SignedUploadUrlResponse, error) {
	signUploadURL := c.clientTransport.baseUrl.String() + "/object/upload/sign/" + bucketId + "/" + filePath
	emptyBody := struct{}{}

	req, err := c.NewRequest(http.MethodPost, signUploadURL, &emptyBody)
	if err != nil {
		return SignedUploadUrlResponse{}, err
	}

	var response SignedUploadUrlResponse
	_, err = c.Do(req, &response)
	if err != nil {
		return SignedUploadUrlResponse{}, err
	}

	return response, err
}

// UploadToSignedUrl upload a file to a signed URL.
// filePath string The file path, including the file name. Should be of the format `folder/subfolder/filename.png`
// fileBody io.Reader The file data
func (c *Client) UploadToSignedUrl(filePath string, fileBody io.Reader) (*UploadToSignedUrlResponse, error) {
	c.clientTransport.header.Set("cache-control", defaultFileCacheControl)
	c.clientTransport.header.Set("content-type", defaultFileContentType)
	c.clientTransport.header.Set("x-upsert", strconv.FormatBool(defaultFileUpsert))

	bodyRequest := bufio.NewReader(fileBody)
	path := removeEmptyFolderName(filePath)
	uploadToSignedURL := c.clientTransport.baseUrl.String() + path

	req, err := http.NewRequest(http.MethodPut, uploadToSignedURL, bodyRequest)
	if err != nil {
		return nil, err
	}

	var response UploadToSignedUrlResponse
	_, err = c.Do(req, &response)
	if err != nil {
		return nil, err
	}

	return &response, err
}

// GetPublicUrl use to to get the URL for an asset in a public bucket. If you do not want to use this function, you can construct the public URL by concatenating the bucket URL with the path to the asset.
// bucketId string The bucket id
// filePath path The file path, including the file name. Should be of the format `folder/subfolder/filename.png`
// urlOptions UrlOptions The URL options
func (c *Client) GetPublicUrl(bucketId string, filePath string, urlOptions ...UrlOptions) SignedUrlResponse {
	var response SignedUrlResponse
	renderPath := "object"
	var options UrlOptions
	if len(urlOptions) > 0 {
		options = urlOptions[0]
		if options.Transform != nil {
			renderPath = "render/image"
		}
	}
	urlStr := c.clientTransport.baseUrl.String() + "/" + renderPath + "/public/" + bucketId + "/" + filePath
	response.SignedURL = buildUrlWithOption(urlStr, options)

	return response
}

// RemoveFile remove a file from an existing bucket.
// bucketId string The bucket id.
// paths []string The file paths, including the file name. Should be of the format `folder/subfolder/filename.png`
func (c *Client) RemoveFile(bucketId string, paths []string) ([]FileUploadResponse, error) {
	removeURL := c.clientTransport.baseUrl.String() + "/object/" + bucketId
	jsonBody := map[string]interface{}{
		"prefixes": paths,
	}

	req, err := c.NewRequest(http.MethodDelete, removeURL, &jsonBody)
	if err != nil {
		return []FileUploadResponse{}, err
	}

	var response []FileUploadResponse
	_, err = c.Do(req, &response)
	if err != nil {
		return []FileUploadResponse{}, err
	}

	return response, err
}

// ListFiles list files in an existing bucket.
// bucketId string The bucket id.
// queryPath string The file path, including the file name. Should be of the format `folder/subfolder/filename.png`
// options FileSearchOptions The file search options
func (c *Client) ListFiles(bucketId string, queryPath string, options FileSearchOptions) ([]FileObject, error) {
	if options.Offset == 0 {
		options.Offset = defaultOffset
	}

	if options.Limit == 0 {
		options.Limit = defaultLimit
	}

	if options.SortByOptions.Order == "" {
		options.SortByOptions.Order = defaultSortOrder
	}

	if options.SortByOptions.Column == "" {
		options.SortByOptions.Column = defaultSortColumn
	}

	body := ListFileRequestBody{
		Limit:  options.Limit,
		Offset: options.Offset,
		SortByOptions: SortBy{
			Column: options.SortByOptions.Column,
			Order:  options.SortByOptions.Order,
		},
		Prefix: queryPath,
	}

	listFileURL := c.clientTransport.baseUrl.String() + "/object/list/" + bucketId
	req, err := c.NewRequest(http.MethodPost, listFileURL, &body)
	if err != nil {
		return []FileObject{}, err
	}

	var response []FileObject
	_, err = c.Do(req, &response)
	if err != nil {
		return []FileObject{}, err
	}

	return response, nil
}

// DownloadFile download a file from an existing bucket.
// bucketId string The bucket id.
// filePath string The file path, including the file name. Should be of the format `folder/subfolder/filename.png`
// urlOptions UrlOptions The URL options
func (c *Client) DownloadFile(bucketId string, filePath string, urlOptions ...UrlOptions) ([]byte, error) {
	var options UrlOptions
	renderPath := "object"
	if len(urlOptions) > 0 {
		options = urlOptions[0]
		if options.Transform != nil {
			renderPath = "render/image/authenticated"
		}
	}
	urlStr := c.clientTransport.baseUrl.String() + "/" + renderPath + "/" + bucketId + "/" + filePath
	req, err := c.NewRequest(http.MethodGet, buildUrlWithOption(urlStr, options), nil)
	if err != nil {
		return nil, err
	}

	res, err := c.Do(req, nil)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	return body, err
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
	if options.Download == true {
		signedURLQuery.Add("download", strconv.FormatBool(options.Download))
	}

	signedURL.RawQuery = signedURLQuery.Encode()
	return signedURL.String()
}

// removeEmptyFolderName replaces occurances of double slashes (//)  with a single slash /
// returns a path string with all double slashes replaced with single slash /
func removeEmptyFolderName(filePath string) string {
	return regexp.MustCompile(`\/\/`).ReplaceAllString(filePath, "/")
}
