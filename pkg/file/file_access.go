package file

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/jaconi-io/secret-file-provider/pkg/templates"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
)

// Name will return either the filename of a single file to contain the secret information or the directory path, where
// all files should be stored in.
func Name(secret *corev1.Secret) (string, error) {
	return templates.Render(viper.GetString(env.SecretFileNamePattern), secret)
}

// ReadAll secret contents of all existing files for the secret.
func ReadAll(filename string) (map[interface{}]interface{}, error) {
	if viper.GetBool(env.SecretFileSingle) {
		return readMultipleFiles(filename)
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	content := make(map[interface{}]interface{})
	err = yaml.NewDecoder(f).Decode(content)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func readMultipleFiles(dir string) (map[interface{}]interface{}, error) {
	result := make(map[interface{}]interface{})

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		path := filepath.Join(dir, file.Name())
		bytes, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		result[file.Name()] = string(bytes)
	}

	return result, nil
}

// WriteAll content either into a single file with the given identifier or into multiple ones under a directory with
// the given name.
func WriteAll(filename string, content map[interface{}]interface{}) error {
	if viper.GetBool(env.SecretFileSingle) {
		return writeMultipleFiles(filename, content)
	}

	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	err = yaml.NewEncoder(f).Encode(content)
	if err != nil {
		return fmt.Errorf("invalid secret content for %s: %w", filename, err)
	}

	return nil
}

func writeMultipleFiles(filename string, content map[interface{}]interface{}) error {
	// TODO handle delete file case!
	if err := os.MkdirAll(filename, os.ModePerm); err != nil {
		return err
	}

	for k, v := range content {
		file := fmt.Sprintf("%v", k)
		content := fmt.Sprintf("%v", v)
		err := os.WriteFile(filepath.Join(filename, file), []byte(content), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}
