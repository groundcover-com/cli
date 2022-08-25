package helm_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
	"groundcover.com/pkg/helm"
)

type HelmValuesTestSuite struct {
	suite.Suite
	ValuesUrl   string
	ValuesFile  string
	OverrideUrl string
	file        *os.File
	server      *httptest.Server
}

func (suite *HelmValuesTestSuite) SetupSuite() {
	var err error

	var urlData []byte
	if urlData, err = yaml.Marshal(map[string]interface{}{"url": "value"}); err != nil {
		suite.T().Fatal(err)
	}

	suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/values.yaml" {
			_, err = w.Write(urlData)
			suite.NoError(err)
		}
	}))

	suite.ValuesUrl = fmt.Sprintf("%s/values.yaml", suite.server.URL)
	suite.OverrideUrl = fmt.Sprintf("%s/override.yaml", suite.server.URL)

	var fileData []byte
	if fileData, err = yaml.Marshal(map[string]interface{}{"file": "value"}); err != nil {
		suite.T().Fatal(err)
	}

	if suite.file, err = os.CreateTemp("", "values"); err != nil {
		suite.T().Fatal(err)
	}

	if _, err = suite.file.Write(fileData); err != nil {
		suite.T().Fatal(err)
	}

	suite.ValuesFile = suite.file.Name()
}

func (suite *HelmValuesTestSuite) TearDownSuite() {
	suite.server.Close()
	suite.file.Close()
	os.Remove(suite.file.Name())
}

func TestHelmValuesTestSuite(t *testing.T) {
	suite.Run(t, &HelmValuesTestSuite{})
}

func (suite *HelmValuesTestSuite) TestOrderChartValuesOverrideSuccess() {
	//prepare
	fileData, err := yaml.Marshal(map[string]interface{}{"file": "override"})
	suite.NoError(err)

	file, err := os.CreateTemp("", "values")
	suite.NoError(err)

	_, err = file.Write(fileData)
	suite.NoError(err)

	chartValues := make(map[string]interface{})
	overridePaths := []string{suite.ValuesFile, file.Name()}

	//act

	valuesOverride, err := helm.LoadChartValuesOverrides(&chartValues, overridePaths)
	suite.NoError(err)

	// assert

	expected := map[string]interface{}{
		"file": "override",
	}

	suite.Equal(expected, chartValues)
	suite.Equal(expected, valuesOverride)
}

func (suite *HelmValuesTestSuite) TestMultiPathsChartValuesOverrideSuccess() {
	//prepare

	chartValues := make(map[string]interface{})
	overridePaths := []string{suite.ValuesUrl, suite.ValuesFile}

	//act

	valuesOverride, err := helm.LoadChartValuesOverrides(&chartValues, overridePaths)
	suite.NoError(err)

	// assert

	expected := map[string]interface{}{
		"file": "value",
		"url":  "value",
	}

	suite.Equal(expected, chartValues)
	suite.Equal(expected, valuesOverride)
}

func (suite *HelmValuesTestSuite) TestUrlChartValuesOverrideSuccess() {
	//prepare
	chartValues := make(map[string]interface{})
	overridePaths := []string{suite.ValuesUrl}

	//act

	valuesOverride, err := helm.LoadChartValuesOverrides(&chartValues, overridePaths)
	suite.NoError(err)

	// assert

	expected := map[string]interface{}{
		"url": "value",
	}

	suite.Equal(expected, chartValues)
	suite.Equal(expected, valuesOverride)
}

func (suite *HelmValuesTestSuite) TestFileChartValuesOverrideSuccess() {
	//prepare

	chartValues := make(map[string]interface{})
	overridePaths := []string{suite.ValuesFile}

	//act

	valuesOverride, err := helm.LoadChartValuesOverrides(&chartValues, overridePaths)
	suite.NoError(err)

	// assert

	expected := map[string]interface{}{
		"file": "value",
	}

	suite.Equal(expected, chartValues)
	suite.Equal(expected, valuesOverride)
}
