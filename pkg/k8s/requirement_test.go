package k8s_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"groundcover.com/pkg/k8s"
)

type KubeRequirementTestSuite struct {
	suite.Suite
	Stdout    *os.File
	ReadPipe  *os.File
	WritePipe *os.File
}

func (suite *KubeRequirementTestSuite) SetupSuite() {
	suite.Stdout = os.Stdout
}

func (suite *KubeRequirementTestSuite) SetupTest() {
	suite.ReadPipe, suite.WritePipe, _ = os.Pipe()
	os.Stdout = suite.WritePipe
}

func (suite *KubeRequirementTestSuite) TearDownSuite() {
	os.Stdout = suite.Stdout
}

func TestKubeRequirementTestSuite(t *testing.T) {
	suite.Run(t, &KubeRequirementTestSuite{})
}

func (suite *KubeRequirementTestSuite) TestRequirementPrintStatusNonCompatible() {
	// prepare
	requirement := k8s.Requirement{
		IsCompatible:    false,
		IsNonCompatible: true,
		Message:         "message",
		ErrorMessages: []string{
			"error-1",
			"error-2",
		},
	}

	// act
	requirement.PrintStatus()
	suite.WritePipe.Close()

	var buf bytes.Buffer
	_, err := io.Copy(&buf, suite.ReadPipe)
	suite.NoError(err)

	// assert
	expected := "✕ message\n• error-1\n• error-2\n"

	suite.Equal(expected, buf.String())
}

func (suite *KubeRequirementTestSuite) TestRequirementPrintStatusCompatible() {
	// prepare
	requirement := k8s.Requirement{
		IsCompatible:    true,
		IsNonCompatible: true,
		Message:         "message",
		ErrorMessages:   []string{},
	}

	// act
	requirement.PrintStatus()
	suite.WritePipe.Close()

	var buf bytes.Buffer
	_, err := io.Copy(&buf, suite.ReadPipe)
	suite.NoError(err)

	// assert
	expected := "✔ message\n"

	suite.Equal(expected, buf.String())
}

func (suite *KubeRequirementTestSuite) TestRequirementPrintStatusPartial() {
	// prepare
	requirement := k8s.Requirement{
		IsCompatible:    false,
		IsNonCompatible: false,
		Message:         "message",
		ErrorMessages: []string{
			"error-1",
		},
	}

	// act
	requirement.PrintStatus()
	suite.WritePipe.Close()

	var buf bytes.Buffer
	_, err := io.Copy(&buf, suite.ReadPipe)
	suite.NoError(err)

	// assert
	expected := "✋ message\n• error-1\n"

	suite.Equal(expected, buf.String())
}
