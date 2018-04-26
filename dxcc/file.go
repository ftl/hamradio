package dxcc

import (
	"io"

	"github.com/ftl/localcopy"
)

// LoadLocal loads a cty.dat file from the local file system.
func LoadLocal(localFilename string) (*Prefixes, error) {
	prefixes, err := localcopy.LoadLocal(localFilename, func(r io.Reader) (interface{}, error) {
		return Read(r)
	})
	if err != nil {
		return nil, err
	}
	return prefixes.(*Prefixes), nil
}

// LoadRemote loads a cty.dat file from a remote URL.
func LoadRemote(remoteURL string) (*Prefixes, error) {
	prefixes, err := localcopy.LoadRemote(remoteURL, func(r io.Reader) (interface{}, error) {
		return Read(r)
	})
	if err != nil {
		return nil, err
	}
	return prefixes.(*Prefixes), nil
}

// Download downloads a cty.dat file from a remote URL and stores it locally.
func Download(remoteURL, localFilename string) error {
	return localcopy.Download(remoteURL, localFilename, func(r io.Reader) (interface{}, error) {
		return Read(r)
	})
}

// Update updates the local copy of a cty.dat file from the given remote URL,
// but only if an update is needed.
func Update(remoteURL, localFilename string) (bool, error) {
	return localcopy.Update(remoteURL, localFilename, func(r io.Reader) (interface{}, error) {
		return Read(r)
	})
}
