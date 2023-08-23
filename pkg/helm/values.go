package helm

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	textTemplate "text/template"

	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
)

//go:embed templates/*
var templatesFS embed.FS

type TemplateValues struct {
	StorageClassName string
}

func GetChartValuesOverrides(paths []string, templateValues *TemplateValues) (map[string]interface{}, error) {
	var err error

	valuesOverride := make(map[string]interface{})

	for _, path := range paths {
		if path == "" {
			continue
		}

		var data []byte
		if data, err = readValuesOverride(path, templateValues); err != nil {
			return nil, err
		}

		var currentValuesOverrides map[string]interface{}
		if err = yaml.Unmarshal(data, &currentValuesOverrides); err != nil {
			return nil, err
		}

		if err = mergo.Merge(&valuesOverride, currentValuesOverrides, mergo.WithOverride); err != nil {
			return nil, err
		}
	}

	return valuesOverride, nil
}

func readTemplateOverride(path string, templateValues *TemplateValues) ([]byte, error) {
	var err error
	var template *textTemplate.Template
	if template, err = textTemplate.ParseFS(templatesFS, path); err != nil {
		return nil, err
	}

	buffer := new(bytes.Buffer)
	if err = template.Execute(buffer, templateValues); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func readValuesOverride(path string, templateValues *TemplateValues) ([]byte, error) {
	var err error

	if strings.HasPrefix(path, "templates") {
		return readTemplateOverride(path, templateValues)
	}

	if strings.HasPrefix(path, "presets") {
		return presetsFS.ReadFile(path)
	}

	overrideUrl, err := url.ParseRequestURI(path)
	if err == nil && overrideUrl.IsAbs() {
		var response *http.Response
		if response, err = http.Get(path); err != nil {
			return nil, err
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("[%d] %s download failed", response.StatusCode, path)
		}

		return io.ReadAll(response.Body)
	}

	return os.ReadFile(path)
}
