package file

import (
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/jaconi-io/secret-file-provider/pkg/templates"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
)

// Name will return either the filename of a single file to contain the secret information
// or the directory path, where all sub-files should be stored in.
func Name(secret *corev1.Secret) string {
	return templates.Resolve(viper.GetString(env.SecretFileNamePattern), secret)
}

// ReadAll returns the secret contents of all existing files for the secret.
// Return current file contents
func ReadAll(logger *slog.Logger, filename string) map[interface{}]interface{} {
	if viper.GetBool(env.SecretFileSingle) {
		return readMultipleFiles(logger, filename)
	}
	bytes, err := os.ReadFile(filename)
	if err != nil {
		// assume file not existing
		bytes = []byte{}
	}
	content := make(map[interface{}]interface{})
	err = yaml.Unmarshal(bytes, content)
	if err != nil {
		logger.Error("failed to map content", "filename", filename, "content", string(bytes), "error", err)
	}
	return content
}

func readMultipleFiles(logger *slog.Logger, filename string) map[interface{}]interface{} {
	result := make(map[interface{}]interface{})
	files, err := ioutil.ReadDir(filename)
	if os.IsNotExist(err) {
		return result
	}
	if err != nil {
		logger.Error("failed to read content", "path", filename, "error", err)
		return result
	}
	for _, file := range files {
		fullpath := filepath.Join(filename, file.Name())
		bytes, err := os.ReadFile(fullpath)
		if err != nil {
			logger.Error("failed to read file content", "filename", filename, "error", err)
			bytes = []byte{}
		}
		result[file.Name()] = string(bytes)
	}
	return result
}

// WriteAll writes all content either into a single file with the given identifier or into multiple
// ones under a directory with the given name.
// Returns potential error
func WriteAll(logger *slog.Logger, filename string, content map[interface{}]interface{}) error {

	if viper.GetBool(env.SecretFileSingle) {
		return writeMultipleFiles(logger, filename, content)
	}

	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		logger.Error("failed to create parent directories", "path", filename, "error", err)
		os.Exit(1)
	}

	yamlData, err := yaml.Marshal(content)
	if err != nil {
		logger.Error("failed to parse secret content for writing", "path", filename, "error", err)
		// do not retry, because secret content is just invalid
		return nil
	}

	err = os.WriteFile(filename, yamlData, 0644)
	if err != nil {
		logger.Error("failed to write secret", "path", filename, "error", err)
		return err
	}
	logger.Info("successfuly written", "path", filename)
	return nil
}

func writeMultipleFiles(logger *slog.Logger, filename string, content map[interface{}]interface{}) error {
	// TODO handle delete file case!
	if err := os.MkdirAll(filename, os.ModePerm); err != nil {
		logger.Error("failed to create parent directories", "path", filename)
		os.Exit(1)
	}
	for k, v := range content {
		file := fmt.Sprintf("%v", k)
		content := fmt.Sprintf("%v", v)
		err := os.WriteFile(filepath.Join(filename, file), []byte(content), 0644)
		if err != nil {
			logger.Error("failed to write secret", "path", file, "error", err)
			return err
		}
		logger.Info("successfully written", "path", file)
	}
	return nil
}
