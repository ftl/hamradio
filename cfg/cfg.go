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

// Key names
type Key string

// Some commonly used parameters.
const (
	MyCall    Key = "my.call"
	MyLocator Key = "my.locator"
)

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

// Exists returns true if the given file exists in the given path. If the path is the empty string, the default
// configuration directory is used. If the given filename is the empty string, the default filename is used.
func Exists(path, filename string) bool {
	absolutePath, err := Directory(path)
	if err != nil {
		return false
	}

	var absoluteFilename string
	if filename == "" {
		absoluteFilename = filepath.Join(absolutePath, DefaultFilename)
	} else {
		absoluteFilename = filepath.Join(absolutePath, filename)
	}

	_, err = os.Stat(absoluteFilename)
	return err == nil
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

// LoadJSON loads JSON data from the given file in the given path and unmarshals it into the given data object.
// If the path is the empty string, the default configuration directory is used. If the given filename is the
// empty string, the default filename is used.
func LoadJSON(path, filename string, data interface{}) error {
	absolutePath, err := Directory(path)
	if err != nil {
		return err
	}

	var absoluteFilename string
	if filename == "" {
		absoluteFilename = filepath.Join(absolutePath, DefaultFilename)
	} else {
		absoluteFilename = filepath.Join(absolutePath, filename)
	}

	file, err := os.Open(absoluteFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	in := bufio.NewReader(file)
	var buffer bytes.Buffer
	_, err = buffer.ReadFrom(in)
	if err != nil {
		return err
	}

	return json.Unmarshal(buffer.Bytes(), &data)
}

// SaveJSON saves the given JSON data to the given file in the given path.
// If the path is the empty string, the default configuration directory is used. If the given filename is the
// empty string, the default filename is used.
func SaveJSON(path, filename string, data interface{}) error {
	absolutePath, err := Directory(path)
	if err != nil {
		return err
	}

	var absoluteFilename string
	if filename == "" {
		absoluteFilename = filepath.Join(absolutePath, DefaultFilename)
	} else {
		absoluteFilename = filepath.Join(absolutePath, filename)
	}

	buf, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	file, err := os.Create(absoluteFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(buf)
	return err
}

// Get retrieves the value at the given path in the configuration data. If the key path
// cannot be found, the given default value is returned.
func (config Configuration) Get(key Key, defaultValue interface{}) interface{} {
	elements := strings.Split(string(key), ".")
	path := elements[:len(elements)-1]
	nodeName := elements[len(elements)-1]
	currentNode := config
	for _, element := range path {
		nextNode := currentNode[element]
		switch nextNode := nextNode.(type) {
		case map[string]interface{}:
			currentNode = nextNode
		default:
			return defaultValue
		}
	}
	if value, exists := currentNode[nodeName]; exists {
		return value
	}
	return defaultValue
}

// GetStrings retrieves the value at the given path as string slice. If the key path
// cannot be found, the given default value is returned.
func (config Configuration) GetStrings(key Key, defaultValue []string) []string {
	rawValues := config.Get(key, nil)
	if rawValues == nil {
		return defaultValue
	}
	values := rawValues.([]interface{})

	result := make([]string, len(values))
	for i, value := range values {
		result[i] = value.(string)
	}
	return result
}

// GetSlice retrieves the value at the given path as array. It iterates over its elements and calls the given callback for each element.
func (config Configuration) GetSlice(key Key, readElement func(int, map[string]interface{})) {
	rawValues := config.Get(key, nil)
	if rawValues == nil {
		return
	}
	values := rawValues.([]interface{})

	for i, rawValue := range values {
		value, ok := rawValue.(map[string]interface{})
		if !ok {
			continue
		}
		readElement(i, value)
	}
}
