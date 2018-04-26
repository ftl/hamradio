package scp

import (
	"io"

	"github.com/ftl/localcopy"
)

// LoadLocal loads the database from a file in the local filesystem.
func LoadLocal(localFilename string) (*Database, error) {
	database, err := localcopy.LoadLocal(localFilename, func(r io.Reader) (interface{}, error) {
		return Read(r)
	})
	if err != nil {
		return nil, err
	}
	return database.(*Database), nil
}

// LoadRemote loads the database file from a remote URL.
func LoadRemote(remoteURL string) (*Database, error) {
	database, err := localcopy.LoadRemote(remoteURL, func(r io.Reader) (interface{}, error) {
		return Read(r)
	})
	if err != nil {
		return nil, err
	}
	return database.(*Database), nil
}

// Download downloads the database file from a remote URL and stores it locally.
func Download(remoteURL, localFilename string) error {
	return localcopy.Download(remoteURL, localFilename, func(r io.Reader) (interface{}, error) {
		return Read(r)
	})
}

// Update updates the local copy of the database file from the given remote URL,
// but only if an update is needed.
func Update(remoteURL, localFilename string) (bool, error) {
	return localcopy.Update(remoteURL, localFilename, func(r io.Reader) (interface{}, error) {
		return Read(r)
	})
}
