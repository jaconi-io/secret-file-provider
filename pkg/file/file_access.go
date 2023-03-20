package file

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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

func ReadAll(logger *logrus.Entry, filename string) map[interface{}]interface{} {
	if viper.GetBool(env.SecretFileSingle) {
		result := make(map[interface{}]interface{})
		files, err := ioutil.ReadDir(filename)
		if os.IsNotExist(err) {
			return result
		}
		logger.WithError(err).Fatalf("Failed to read content of %s", filename)
		for _, file := range files {
			fullpath := filepath.Join(filename, file.Name())
			bytes, err := os.ReadFile(fullpath)
			if err != nil {
				logger.WithError(err).Errorf("Failed to read file content for %s", file.Name())
				bytes = []byte{}
			}
			result[file.Name()] = string(bytes)
		}
		return result
	}
	bytes, err := os.ReadFile(filename)
	if err != nil {
		// file not existing
		bytes = []byte{}
	}
	content := make(map[interface{}]interface{})
	err = yaml.Unmarshal(bytes, content)
	if err != nil {
		logger.WithError(err).Errorf("Failed to read %s", string(bytes))
	}
	return content
}

func WriteAll(logger *logrus.Entry, filename string, content map[interface{}]interface{}) error {

	if viper.GetBool(env.SecretFileSingle) {
		if err := os.MkdirAll(filename, os.ModePerm); err != nil {
			logger.WithError(err).Fatalf("Failed to create parent directories for %s", filename)
		}
		for k, v := range content {
			file := fmt.Sprintf("%v", k)
			content := fmt.Sprintf("%v", v)
			err := os.WriteFile(filepath.Join(filename, file), []byte(content), 0644)
			if err != nil {
				logger.WithError(err).Errorf("Failed to write secret to %s", file)
				return err
			}
			logger.Infof("Successfuly written %s", file)
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		logger.WithError(err).Fatalf("Failed to create parent directories for %s", filename)
	}

	yamlData, err := yaml.Marshal(content)
	if err != nil {
		logger.WithError(err).Errorf("Failed to parse secret content for writing into %s", filename)
		// do not retry, because secret content is just invalid
		return nil
	}

	err = os.WriteFile(filename, yamlData, 0644)
	if err != nil {
		logger.WithError(err).Errorf("Failed to write secret to %s", filename)
	} else {
		logger.Infof("Successfuly written %s", filename)
	}
	return err
}
