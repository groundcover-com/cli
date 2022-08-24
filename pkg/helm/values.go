package helm

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"k8s.io/apimachinery/pkg/util/yaml"
)

func LoadChartValuesOverrides(chartValues *map[string]interface{}, paths []string) (map[string]interface{}, error) {
	var err error

	valuesOverride := make(map[string]interface{})

	for _, path := range paths {
		var data []byte
		if data, err = readValuesOverride(path); err != nil {
			return nil, err
		}

		if err = yaml.Unmarshal(data, &valuesOverride); err != nil {
			return nil, err
		}

		if err = yaml.Unmarshal(data, chartValues); err != nil {
			return nil, err
		}
	}

	return valuesOverride, nil
}

func readValuesOverride(path string) ([]byte, error) {
	var err error

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

		return ioutil.ReadAll(response.Body)
	}

	return os.ReadFile(path)
}
