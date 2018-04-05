/*
Package cfg implements a library to access configuration data in a JSON file.

All hamradio tools share the same file in the same directory for configuration data: ~/.config/hamradio/conf.json
*/
package cfg

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// Configuration contains configuration data in a generic key value structure.
type Configuration map[string]interface{}

// DefaultDirectory is the place where all hamradio tools store their configuration locally.
const DefaultDirectory = "~/.config/hamradio"

// DefaultFilename is the default name of the configuration file that is used by all hamradio tools,
// relative to the configuration directory.
const DefaultFilename = "conf.json"

// Directory returns the path of the configuration directory as absolute path. If the given
// path is the empty string, the default directory is returned.
func Directory(path string) (string, error) {
	if path != "" {
		return resolvePath(path)
	}
	return resolvePath(DefaultDirectory)
}

func resolvePath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	if strings.HasPrefix(path, "~/") {
		homeDir, err := homeDir()
		if err != nil {
			return "", err
		}
		return strings.Replace(path, "~", homeDir, 1), nil
	}
	return filepath.Abs(path)
}

func homeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return usr.HomeDir, nil
}

// PrepareDirectory ensures that the given directory exists. The given path is resolved and any missing parent
// directories are created if necessary. The function returns the absolute path of the directory. If the given
// path is the empty string, the default configuration directory is used.
func PrepareDirectory(path string) (string, error) {
	absolutePath, err := Directory(path)
	if err != nil {
		return "", err
	}
	return absolutePath, os.MkdirAll(absolutePath, os.ModePerm)
}

// LoadDefault loads JSON configuration data from the default file in the default configuration directory.
func LoadDefault() (Configuration, error) {
	return Load("", "")
}

// Load loads the configuration from the given file in the given directory. If the path is the empty string, the default
// configuration directory is used. If the given filename is the empty string, the default filename is used.
func Load(path, filename string) (Configuration, error) {
	absolutePath, err := Directory(path)
	if err != nil {
		return Configuration{}, err
	}

	var absoluteFilename string
	if filename == "" {
		absoluteFilename = filepath.Join(absolutePath, DefaultFilename)
	} else {
		absoluteFilename = filepath.Join(absolutePath, filename)
	}

	file, err := os.Open(absoluteFilename)
	if err != nil {
		return Configuration{}, err
	}
	defer file.Close()

	in := bufio.NewReader(file)
	config, err := Read(in)
	if err != nil {
		return Configuration{}, err
	}
	return config, nil
}

// Read reads JSON configuration data from the given reader.
func Read(in io.Reader) (Configuration, error) {
	var buffer bytes.Buffer
	_, err := buffer.ReadFrom(in)
	if err != nil {
		return Configuration{}, err
	}

	var data interface{}
	err = json.Unmarshal(buffer.Bytes(), &data)
	if err != nil {
		return Configuration{}, err
	}

	return Configuration(data.(map[string]interface{})), nil
}
