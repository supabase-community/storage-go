package storage_go

type StorageError struct {
	Status     int    `json:"status"`
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
	ErrorCode  string `json:"error"`
	RawBody    string `json:"-"`
}

func (e *StorageError) Error() string {
	if e.Message == "" {
		return e.RawBody
	}

	return e.Message
}

func NewStorageError(err error, statusCode int) StorageError {
	return StorageError{
		Status:     statusCode,
		Message:    err.Error(),
		StatusCode: statusCode,
	}
}
