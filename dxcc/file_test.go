package dxcc

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestLoadLocal(t *testing.T) {
	prefixes, err := LoadLocal("./testdata/cty.dat")
	if err != nil {
		t.Errorf("loading failed: %v", err)
		t.FailNow()
	}
	if len(prefixes.items) != 5517 {
		t.Errorf("expected 5517 prefixes, but found %d", len(prefixes.items))
	}
}

func TestLoadRemote(t *testing.T) {
	testServer := httptest.NewServer(serveCtyDat)
	defer testServer.Close()

	prefixes, err := LoadRemote(testServer.URL)
	if err != nil {
		t.Errorf("loading failed: %v", err)
		t.FailNow()
	}
	if len(prefixes.items) != 5517 {
		t.Errorf("expected 5517 prefixes, but found %d", len(prefixes.items))
	}
}

func TestDownload(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "dxcc.TestDownload")
	if err != nil {
		t.Errorf("failed to create temp file: %v", err)
		t.FailNow()
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	testServer := httptest.NewServer(serveCtyDat)
	defer testServer.Close()

	err = Download(testServer.URL, tempFile.Name())
	if err != nil {
		t.Errorf("failed to download: %v", err)
		t.FailNow()
	}

	tempFileInfo, _ := os.Stat(tempFile.Name())
	testFileInfo, _ := os.Stat(testFilename)
	if tempFileInfo.Size() != testFileInfo.Size() {
		t.Errorf("expected file size %d, but got %d", testFileInfo.Size(), tempFileInfo.Size())
	}
}

func TestNeedsUpdate(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "dxcc.TestDownload")
	if err != nil {
		t.Errorf("failed to create temp file: %v", err)
		t.FailNow()
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())
	tempFileInfo, _ := os.Stat(tempFile.Name())

	timeToServe := time.Now()
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !timeToServe.IsZero() {
			w.Header().Add(httpLastModified, timeToServe.Format(httpTimeFormat))
		}
	}))
	defer testServer.Close()

	testCases := []struct {
		remoteFileTime time.Time
		needsUpdate    bool
	}{
		{time.Time{}, false},
		{tempFileInfo.ModTime().Add(-10 * time.Minute), false},
		{tempFileInfo.ModTime().Add(10 * time.Minute), true},
	}

	for _, testCase := range testCases {
		timeToServe = testCase.remoteFileTime

		needsUpdate, err := NeedsUpdate(testServer.URL, tempFile.Name())
		if timeToServe.IsZero() {
			if err == nil {
				t.Errorf("missing Last-Update header should raise an error")
			}
		} else {
			if err != nil {
				t.Errorf("failed to check for update: %v", err)
				t.FailNow()
			}
			if needsUpdate != testCase.needsUpdate {
				t.Errorf("expected needsUpdate %t, but got %t for %v", testCase.needsUpdate, needsUpdate, timeToServe.Sub(tempFileInfo.ModTime()))
			}
		}
	}
}

const testFilename = "./testdata/cty.dat"

var serveCtyDat = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, testFilename)
})
