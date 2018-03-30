package dxcc

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const httpTimeFormat = time.RFC1123
const httpLastModified = "Last-Modified"

var httpClient = &http.Client{
	Timeout: time.Second * 10,
}

// LoadLocal loads a cty.dat file from the local file system.
func LoadLocal(localFilename string) (*Prefixes, error) {
	file, err := os.Open(localFilename)
	if err != nil {
		return &Prefixes{}, err
	}
	defer file.Close()

	in := bufio.NewReader(file)
	prefixes, err := Read(in)
	if err != nil {
		return &Prefixes{}, err
	}
	return prefixes, nil
}

// LoadRemote loads a cty.dat file from a remote URL.
func LoadRemote(remoteURL string) (*Prefixes, error) {
	resp, err := httpClient.Get(remoteURL)
	if err != nil {
		return &Prefixes{}, err
	}
	defer resp.Body.Close()

	in := bufio.NewReader(resp.Body)
	prefixes, err := Read(in)
	if err != nil {
		return &Prefixes{}, err
	}
	return prefixes, nil
}

// Download downloads a cty.dat file from a remote URL and stores it locally.
func Download(remoteURL, localFilename string) error {
	response, err := httpClient.Get(remoteURL)
	if err != nil {
		return fmt.Errorf("failed to download cty.dat: %v", err)
	}
	defer response.Body.Close()

	os.MkdirAll(filepath.Dir(localFilename), os.ModePerm)
	localFile, err := os.Create(localFilename)
	if err != nil {
		return fmt.Errorf("failed to open local file: %v", err)
	}
	defer localFile.Close()

	_, err = io.Copy(localFile, response.Body)
	if err != nil {
		return fmt.Errorf("failed to store cty.dat locally: %v", err)
	}

	return nil
}

// NeedsUpdate checks whether the local copy of a cty.dat file needs to be
// updated from the given remote URL.
func NeedsUpdate(remoteURL, localFilename string) (bool, error) {
	response, err := httpClient.Head(remoteURL)
	if err != nil {
		return false, err
	}
	var lastModified time.Time
	if lastModifiedHeader, ok := response.Header[httpLastModified]; ok {
		if len(lastModifiedHeader) == 0 {
			return false, fmt.Errorf("Last-Modified header is empty")
		}

		lastModified, err = time.Parse(httpTimeFormat, lastModifiedHeader[0])
		if err != nil {
			return false, fmt.Errorf("cannot parse Last-Modified header: %v", err)
		}
	} else {
		return false, fmt.Errorf("response does not contain a Last-Modified header")
	}

	localFileInfo, err := os.Stat(localFilename)
	if os.IsNotExist(err) {
		return true, nil
	} else if err != nil {
		return false, err
	}

	return lastModified.After(localFileInfo.ModTime()), nil
}

// Update updates the local copy of a cty.dat file from the given remote URL,
// but only if an update is needed.
func Update(remoteURL, localFilename string) (bool, error) {
	needsUpdate, err := NeedsUpdate(remoteURL, localFilename)
	if err != nil {
		return false, err
	}

	if !needsUpdate {
		return false, nil
	}
	return true, Download(remoteURL, localFilename)
}
