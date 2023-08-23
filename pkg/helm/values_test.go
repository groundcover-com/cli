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
	defer suite.file.Close()

	if _, err = suite.file.Write(fileData); err != nil {
		suite.T().Fatal(err)
	}

	suite.ValuesFile = suite.file.Name()
}

func (suite *HelmValuesTestSuite) TearDownSuite() {
	suite.server.Close()
	os.Remove(suite.file.Name())
}

func TestHelmValuesTestSuite(t *testing.T) {
	suite.Run(t, &HelmValuesTestSuite{})
}

func (suite *HelmValuesTestSuite) TestOrderChartValuesOverrideSuccess() {
	//prepare
	templateValues := helm.TemplateValues{}
	fileData, err := yaml.Marshal(map[string]interface{}{"file": "override"})
	suite.NoError(err)

	file, err := os.CreateTemp("", "values")
	suite.NoError(err)
	defer file.Close()
	defer os.Remove(file.Name())

	_, err = file.Write(fileData)
	suite.NoError(err)

	overridePaths := []string{suite.ValuesFile, file.Name()}

	//act
	chartValues, err := helm.GetChartValuesOverrides(overridePaths, &templateValues)
	suite.NoError(err)

	// assert

	expected := map[string]interface{}{
		"file": "override",
	}

	suite.Equal(expected, chartValues)
}

func (suite *HelmValuesTestSuite) TestMultiPathsChartValuesOverrideSuccess() {
	//prepare
	templateValues := helm.TemplateValues{}
	overridePaths := []string{suite.ValuesUrl, suite.ValuesFile}

	//act
	chartValues, err := helm.GetChartValuesOverrides(overridePaths, &templateValues)
	suite.NoError(err)

	// assert

	expected := map[string]interface{}{
		"file": "value",
		"url":  "value",
	}

	suite.Equal(expected, chartValues)
}

func (suite *HelmValuesTestSuite) TestUrlChartValuesOverrideSuccess() {
	//prepare
	templateValues := helm.TemplateValues{}
	overridePaths := []string{suite.ValuesUrl}

	//act
	chartValues, err := helm.GetChartValuesOverrides(overridePaths, &templateValues)
	suite.NoError(err)

	// assert

	expected := map[string]interface{}{
		"url": "value",
	}

	suite.Equal(expected, chartValues)
}

func (suite *HelmValuesTestSuite) TestFileChartValuesOverrideSuccess() {
	//prepare
	templateValues := helm.TemplateValues{}
	overridePaths := []string{suite.ValuesFile}

	//act
	chartValues, err := helm.GetChartValuesOverrides(overridePaths, &templateValues)
	suite.NoError(err)

	// assert
	expected := map[string]interface{}{
		"file": "value",
	}

	suite.Equal(expected, chartValues)
}
