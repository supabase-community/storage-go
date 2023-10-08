package storage_go

type StorageError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (e *StorageError) Error() string {
	return e.Message
}

func NewStorageError(err error, statusCode int) StorageError {
	return StorageError{
		Status:  statusCode,
		Message: err.Error(),
	}
}
