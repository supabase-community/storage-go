package storage_go

import (
	"bufio"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Upload uploads a file to the scoped bucket.
func (f *FileClient) Upload(path string, body io.Reader, options ...FileOptions) (FileUploadResponse, error) {
	return f.uploadOrUpdate(http.MethodPost, path, body, options...)
}

// Update replaces an existing file in the scoped bucket.
func (f *FileClient) Update(path string, body io.Reader, options ...FileOptions) (FileUploadResponse, error) {
	return f.uploadOrUpdate(http.MethodPut, path, body, options...)
}

// Download downloads a file from the scoped bucket.
func (f *FileClient) Download(path string, options ...UrlOptions) ([]byte, error) {
	renderPath := "object"
	var option UrlOptions
	if len(options) > 0 {
		option = options[0]
		if option.Transform != nil {
			renderPath = "render/image/authenticated"
		}
	}

	req, err := f.client.NewRequest(
		http.MethodGet,
		buildUrlWithOption("/"+renderPath+"/"+objectPath(f.bucketID, path), option),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := f.client.Do(req, nil)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := resp.Body.Close(); err != nil {
		return nil, err
	}

	return data, nil
}

// Remove removes files from the scoped bucket.
func (f *FileClient) Remove(paths []string) ([]FileUploadResponse, error) {
	body := map[string][]string{"prefixes": paths}
	req, err := f.client.NewRequest(http.MethodDelete, "/object/"+url.PathEscape(f.bucketID), body)
	if err != nil {
		return []FileUploadResponse{}, err
	}

	var response []FileUploadResponse
	if _, err := f.client.Do(req, &response); err != nil {
		return []FileUploadResponse{}, err
	}

	return response, nil
}

// List lists files in the scoped bucket.
func (f *FileClient) List(path string, options FileSearchOptions) ([]FileObject, error) {
	body := listFileRequest(path, options)
	req, err := f.client.NewRequest(http.MethodPost, "/object/list/"+url.PathEscape(f.bucketID), body)
	if err != nil {
		return []FileObject{}, err
	}

	var response []FileObject
	if _, err := f.client.Do(req, &response); err != nil {
		return []FileObject{}, err
	}

	return response, nil
}

// Move moves a file within or across buckets.
func (f *FileClient) Move(sourceKey, destinationKey string, options ...DestinationOptions) (FileUploadResponse, error) {
	return f.moveOrCopy("/object/move", sourceKey, destinationKey, options...)
}

// Copy copies a file within or across buckets.
func (f *FileClient) Copy(sourceKey, destinationKey string, options ...DestinationOptions) (FileUploadResponse, error) {
	return f.moveOrCopy("/object/copy", sourceKey, destinationKey, options...)
}

// CreateSignedURL creates a signed download URL for a file.
func (f *FileClient) CreateSignedURL(path string, expiresIn int, options ...UrlOptions) (SignedUrlResponse, error) {
	body := map[string]any{"expiresIn": expiresIn}
	if len(options) > 0 {
		if options[0].Transform != nil {
			body["transform"] = options[0].Transform
		}
		if options[0].Download {
			body["download"] = true
		}
	}

	req, err := f.client.NewRequest(http.MethodPost, "/object/sign/"+objectPath(f.bucketID, path), body)
	if err != nil {
		return SignedUrlResponse{}, err
	}

	var response SignedUrlResponse
	if _, err := f.client.Do(req, &response); err != nil {
		return SignedUrlResponse{}, err
	}
	response.SignedURL = f.absoluteURL(response.SignedURL)

	return response, nil
}

// CreateSignedURLs creates signed download URLs for multiple files.
func (f *FileClient) CreateSignedURLs(paths []string, expiresIn int, options ...UrlOptions) ([]SignedUrlResponse, error) {
	body := map[string]any{
		"expiresIn": expiresIn,
		"paths":     paths,
	}
	if len(options) > 0 {
		if options[0].Transform != nil {
			body["transform"] = options[0].Transform
		}
		if options[0].Download {
			body["download"] = true
		}
	}

	req, err := f.client.NewRequest(http.MethodPost, "/object/sign/"+url.PathEscape(f.bucketID), body)
	if err != nil {
		return nil, err
	}

	var response []SignedUrlResponse
	if _, err := f.client.Do(req, &response); err != nil {
		return nil, err
	}
	for i := range response {
		response[i].SignedURL = f.absoluteURL(response[i].SignedURL)
	}

	return response, nil
}

// CreateSignedUploadURL creates a signed upload URL for a file.
func (f *FileClient) CreateSignedUploadURL(path string, options ...FileOptions) (SignedUploadUrlResponse, error) {
	body := map[string]any{}
	if len(options) > 0 && options[0].Upsert != nil {
		body["upsert"] = *options[0].Upsert
	}
	req, err := f.client.NewRequest(http.MethodPost, "/object/upload/sign/"+objectPath(f.bucketID, path), body)
	if err != nil {
		return SignedUploadUrlResponse{}, err
	}

	var response SignedUploadUrlResponse
	if _, err := f.client.Do(req, &response); err != nil {
		return SignedUploadUrlResponse{}, err
	}
	response.Url = f.absoluteURL(response.Url)

	return response, nil
}

// UploadToSignedURL uploads a file body to a signed upload URL.
func (f *FileClient) UploadToSignedURL(signedURL string, body io.Reader, options ...FileOptions) (UploadToSignedUrlResponse, error) {
	req, err := http.NewRequest(http.MethodPut, signedURL, bufio.NewReader(body))
	if err != nil {
		return UploadToSignedUrlResponse{}, err
	}
	applyFileOptions(req, options...)

	var response UploadToSignedUrlResponse
	if _, err := f.client.Do(req, &response); err != nil {
		return UploadToSignedUrlResponse{}, err
	}

	return response, nil
}

// Info retrieves metadata for a file.
func (f *FileClient) Info(path string) (FileObject, error) {
	req, err := f.client.NewRequest(http.MethodGet, "/object/info/"+objectPath(f.bucketID, path), nil)
	if err != nil {
		return FileObject{}, err
	}

	var response FileObject
	if _, err := f.client.Do(req, &response); err != nil {
		return FileObject{}, err
	}

	return response, nil
}

// Exists reports whether a file exists.
func (f *FileClient) Exists(path string) (bool, error) {
	req, err := f.client.NewRequest(http.MethodHead, "/object/"+objectPath(f.bucketID, path), nil)
	if err != nil {
		return false, err
	}

	resp, err := f.client.session.Do(req)
	if err != nil {
		return false, err
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent:
		if resp.Body != nil {
			if err := resp.Body.Close(); err != nil {
				return false, err
			}
		}
		return true, nil
	case http.StatusNotFound:
		if resp.Body != nil {
			if err := resp.Body.Close(); err != nil {
				return false, err
			}
		}
		return false, nil
	default:
		return false, checkForError(resp)
	}
}

// PublicURL returns the public URL for a file.
func (f *FileClient) PublicURL(path string, options ...UrlOptions) SignedUrlResponse {
	renderPath := "object"
	var option UrlOptions
	if len(options) > 0 {
		option = options[0]
		if option.Transform != nil {
			renderPath = "render/image"
		}
	}

	base := f.client.clientTransport.baseUrl.String()
	return SignedUrlResponse{
		SignedURL: buildUrlWithOption(base+"/"+renderPath+"/public/"+objectPath(f.bucketID, path), option),
	}
}

// ListV2 lists files and folders using the v2 API.
func (f *FileClient) ListV2(path string, options SearchV2Options) ([]SearchV2Result, error) {
	if options.Prefix == "" {
		options.Prefix = path
	}
	req, err := f.client.NewRequest(http.MethodPost, "/object/list-v2/"+url.PathEscape(f.bucketID), options)
	if err != nil {
		return nil, err
	}

	var response []SearchV2Result
	if _, err := f.client.Do(req, &response); err != nil {
		return nil, err
	}

	return response, nil
}

func (f *FileClient) uploadOrUpdate(
	method string,
	path string,
	body io.Reader,
	options ...FileOptions,
) (FileUploadResponse, error) {
	req, err := http.NewRequest(method, "/object/"+objectPath(f.bucketID, path), bufio.NewReader(body))
	if err != nil {
		return FileUploadResponse{}, err
	}
	applyFileOptions(req, options...)

	var response FileUploadResponse
	if _, err := f.client.Do(req, &response); err != nil {
		return FileUploadResponse{}, err
	}

	return response, nil
}

func (f *FileClient) moveOrCopy(
	path string,
	sourceKey string,
	destinationKey string,
	options ...DestinationOptions,
) (FileUploadResponse, error) {
	body := map[string]any{
		"bucketId":       f.bucketID,
		"sourceKey":      sourceKey,
		"destinationKey": destinationKey,
	}
	if len(options) > 0 && options[0].DestinationBucket != "" {
		body["destinationBucket"] = options[0].DestinationBucket
	}

	req, err := f.client.NewRequest(http.MethodPost, path, body)
	if err != nil {
		return FileUploadResponse{}, err
	}

	var response FileUploadResponse
	if _, err := f.client.Do(req, &response); err != nil {
		return FileUploadResponse{}, err
	}

	return response, nil
}

func (f *FileClient) absoluteURL(path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return f.client.clientTransport.baseUrl.String() + path
}

func applyFileOptions(req *http.Request, options ...FileOptions) {
	cacheControl := defaultFileCacheControl
	contentType := defaultFileContentType
	upsert := defaultFileUpsert

	if len(options) > 0 {
		option := options[0]
		if option.CacheControl != nil {
			cacheControl = *option.CacheControl
		}
		if option.ContentType != nil {
			contentType = *option.ContentType
		}
		if option.Upsert != nil {
			upsert = *option.Upsert
		}
	}

	req.Header.Set("cache-control", cacheControl)
	req.Header.Set("content-type", contentType)
	req.Header.Set("x-upsert", strconv.FormatBool(upsert))
}

func listFileRequest(path string, options FileSearchOptions) ListFileRequestBody {
	if options.Limit == 0 {
		options.Limit = defaultLimit
	}
	if options.Offset == 0 {
		options.Offset = defaultOffset
	}
	if options.SortByOptions.Order == "" {
		options.SortByOptions.Order = defaultSortOrder
	}
	if options.SortByOptions.Column == "" {
		options.SortByOptions.Column = defaultSortColumn
	}

	return ListFileRequestBody{
		Limit:  options.Limit,
		Offset: options.Offset,
		SortByOptions: SortBy{
			Column: options.SortByOptions.Column,
			Order:  options.SortByOptions.Order,
		},
		Prefix: path,
	}
}

func objectPath(bucketID string, path string) string {
	parts := []string{url.PathEscape(bucketID)}
	for _, part := range strings.Split(strings.TrimLeft(path, "/"), "/") {
		if part == "" {
			continue
		}
		parts = append(parts, url.PathEscape(part))
	}

	return strings.Join(parts, "/")
}
