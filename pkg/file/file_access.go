package file

import (
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/jaconi-io/secret-file-provider/pkg/env"
	"github.com/jaconi-io/secret-file-provider/pkg/logger"
	"github.com/jaconi-io/secret-file-provider/pkg/templates"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
)

func Name(secret *corev1.Secret) string {

	filePattern := viper.GetString(env.SecretFileNamePattern)
	if len(filePattern) < 1 {
		logger.New(secret).Fatalf("Missing required property %s", env.SecretFileNamePattern)
	}
	return templates.Resolve(filePattern, secret)
}

// ReadAll current content plus checksum
func ReadAll(logger *logrus.Entry, filename string) (map[interface{}]interface{}, uint32) {
	if viper.GetBool(env.SecretFileSingle) {
		result := make(map[interface{}]interface{})
		files, err := ioutil.ReadDir(filename)
		if os.IsNotExist(err) {
			return result, 0
		}
		if err != nil {
			logger.WithError(err).Errorf("Failed to read content of %s", filename)
			return result, 0
		}
		readContent := make([]string, len(files))
		count := 0
		for _, file := range files {
			fullpath := filepath.Join(filename, file.Name())
			bytes, err := os.ReadFile(fullpath)
			if err != nil {
				logger.WithError(err).Errorf("Failed to read file content for %s", file.Name())
				bytes = []byte{}
			}
			result[file.Name()] = string(bytes)
			readContent[count] = string(bytes)
			count++
		}
		return result, hashStringArray(readContent)
	}
	bytes, err := os.ReadFile(filename)
	if err != nil {
		// file not existing
		bytes = []byte{}
	}
	content := make(map[interface{}]interface{})
	err = yaml.Unmarshal(bytes, content)
	if err != nil {
		logger.WithError(err).Errorf("Failed to map content of %s:  %s", filename, string(bytes))
	}
	return content, hash(bytes)
}

// WriteAll - checksum, error
func WriteAll(logger *logrus.Entry, filename string, content map[interface{}]interface{}) (uint32, error) {

	if viper.GetBool(env.SecretFileSingle) {
		// TODO handle delete file case!
		if err := os.MkdirAll(filename, os.ModePerm); err != nil {
			logger.WithError(err).Fatalf("Failed to create parent directories for %s", filename)
		}
		writtenContent := make([]string, len(content))
		count := 0
		for k, v := range content {
			file := fmt.Sprintf("%v", k)
			content := fmt.Sprintf("%v", v)
			err := os.WriteFile(filepath.Join(filename, file), []byte(content), 0644)
			if err != nil {
				logger.WithError(err).Errorf("Failed to write secret to %s", file)
				return 0, err
			}
			writtenContent[count] = content
			count++
			logger.Infof("Successfuly written %s", file)
		}
		return hashStringArray(writtenContent), nil
	}

	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		logger.WithError(err).Fatalf("Failed to create parent directories for %s", filename)
	}

	yamlData, err := yaml.Marshal(content)
	if err != nil {
		logger.WithError(err).Errorf("Failed to parse secret content for writing into %s", filename)
		// do not retry, because secret content is just invalid
		return 0, nil
	}

	err = os.WriteFile(filename, yamlData, 0644)
	if err != nil {
		logger.WithError(err).Errorf("Failed to write secret to %s", filename)
		return 0, err
	}
	logger.Infof("Successfuly written %s", filename)
	return hash(yamlData), nil
}

func hash(bytes []byte) uint32 {
	h := fnv.New32a()
	h.Write(bytes)
	return h.Sum32()
}

func hashString(s string) uint32 {
	return hash([]byte(s))
}

func hashStringArray(arr []string) uint32 {
	sort.Strings(arr)
	var value uint32 = 0
	for _, s := range arr {
		value += hashString(s)
	}
	return value
}
