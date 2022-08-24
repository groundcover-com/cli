package helm_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
	"groundcover.com/pkg/helm"
)

type HelmValuesTestSuite struct {
	suite.Suite
}

func (suite *HelmValuesTestSuite) SetupSuite() {}

func (suite *HelmValuesTestSuite) TearDownSuite() {}

func TestHelmValuesTestSuite(t *testing.T) {
	suite.Run(t, &HelmValuesTestSuite{})
}

func (suite *HelmValuesTestSuite) TestMultiPathsLoadChartValuesOverrideSuccess() {
	//prepare
	urlData, err := yaml.Marshal(map[string]interface{}{"url": uuid.New().String()})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/values.yaml" {
			_, err = w.Write(urlData)
		}
	}))
	defer server.Close()

	fileData, err := yaml.Marshal(map[string]interface{}{"file": uuid.New().String()})
	valuesFile, err := os.CreateTemp("", "values")
	suite.NoError(err)
	defer valuesFile.Close()
	defer os.Remove(valuesFile.Name())

	_, err = valuesFile.Write(fileData)
	suite.NoError(err)

	chartValues := make(map[string]interface{})
	overridePaths := []string{fmt.Sprintf("%s/values.yaml", server.URL), valuesFile.Name()}

	//act

	valuesOverride, err := helm.LoadChartValuesOverrides(&chartValues, overridePaths)
	suite.NoError(err)

	// assert

	expected := make(map[string]interface{})
	yaml.Unmarshal(urlData, &expected)
	yaml.Unmarshal(fileData, &expected)

	suite.Equal(expected, chartValues)
	suite.Equal(expected, valuesOverride)
}

func (suite *HelmValuesTestSuite) TestUrlLoadChartValuesOverrideSuccess() {
	//prepare
	urlData, err := yaml.Marshal(map[string]interface{}{"url": uuid.New().String()})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/values.yaml" {
			_, err = w.Write(urlData)
		}
	}))
	defer server.Close()

	chartValues := make(map[string]interface{})
	overridePaths := []string{fmt.Sprintf("%s/values.yaml", server.URL)}

	//act

	valuesOverride, err := helm.LoadChartValuesOverrides(&chartValues, overridePaths)
	suite.NoError(err)

	// assert

	expected := make(map[string]interface{})
	yaml.Unmarshal(urlData, &expected)

	suite.Equal(expected, chartValues)
	suite.Equal(expected, valuesOverride)
}

func (suite *HelmValuesTestSuite) TestFileLoadChartValuesOverrideSuccess() {
	//prepare
	fileData, err := yaml.Marshal(map[string]interface{}{"file": uuid.New().String()})
	valuesFile, err := os.CreateTemp("", "values")
	suite.NoError(err)
	defer valuesFile.Close()
	defer os.Remove(valuesFile.Name())

	_, err = valuesFile.Write(fileData)
	suite.NoError(err)

	chartValues := make(map[string]interface{})
	overridePaths := []string{valuesFile.Name()}

	//act

	valuesOverride, err := helm.LoadChartValuesOverrides(&chartValues, overridePaths)
	suite.NoError(err)

	// assert

	expected := make(map[string]interface{})
	yaml.Unmarshal(fileData, &expected)

	suite.Equal(expected, chartValues)
	suite.Equal(expected, valuesOverride)
}
