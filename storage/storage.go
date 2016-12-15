package storage

import (
	"io"
	"log"
)

// CloudStorage interface
type CloudStorage interface {
	Save(io.Reader, string, string) (*CloudStorageFile, error)
}

// CloudStorageFile file in CloudStorageFile
type CloudStorageFile struct {
	FilePath string
	URI      string
	Size     uint64
}

// CloudStorageType storage type
type CloudStorageType int

const (
	// GoogleCloudStorage google cloud storage
	GoogleCloudStorage = iota
)

// NewCloudStorage new CloudStorage instance
func NewCloudStorage(t CloudStorageType) CloudStorage {
	var cloudStorage CloudStorage
	switch t {
	case GoogleCloudStorage:
		cloudStorage = newGoogleStorage()
	default:
		log.Fatal("storage type is not support")
	}
	return cloudStorage
}
