package k8s_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"groundcover.com/pkg/k8s"
	"k8s.io/client-go/kubernetes/fake"
)

type KubeAuthTestSuite struct {
	suite.Suite
	KubeClient k8s.Client
}

func (suite *KubeAuthTestSuite) SetupTest() {
	suite.KubeClient = k8s.Client{
		Interface: fake.NewSimpleClientset(),
	}
}

func (suite *KubeAuthTestSuite) TearDownSuite() {}

func TestKubeAuthTestSuite(t *testing.T) {
	suite.Run(t, &KubeAuthTestSuite{})
}

func TestValidateAwsCliVersionSupported(t *testing.T) {
	testCases := []struct {
		name     string
		version  string
		expected error
	}{
		{
			name:     "bad format version, no spaces",
			version:  "bad",
			expected: fmt.Errorf("unknown aws cli version: \"bad\""),
		},
		{
			name:     "bad format version, first path no slash",
			version:  "bad version",
			expected: fmt.Errorf("unknown aws cli version: \"bad version\""),
		},
		{
			name:     "aws cli version 1.18.0",
			version:  "aws-cli/1.18.0 Python/3.7.4 Darwin/19.4.0 botocore/1.17.0",
			expected: fmt.Errorf("aws-cli version is unsupported (1.18.0 < 1.23.9)"),
		},
		{
			name:     "aws cli version 1.23.9",
			version:  "aws-cli/1.23.9 Python/3.7.4 Darwin/19.4.0 botocore/1.23.9",
			expected: nil,
		},
		{
			name:     "aws cli version 1.23.10",
			version:  "aws-cli/1.23.10 Python/3.7.4 Darwin/19.4.0 botocore/1.23.10",
			expected: nil,
		},
		{
			name:     "aws cli version 2.0.0",
			version:  "aws-cli/2.0.0 Python/3.7.4 Darwin/19.4.0 botocore/2.0.0dev0",
			expected: fmt.Errorf("aws-cli version is unsupported (2.0.0 < 2.7.0)"),
		},
		{
			name:     "aws cli version 2.7.0",
			version:  "aws-cli/2.7.0 Python/3.7.4 Darwin/19.4.0 botocore/2.7.0",
			expected: nil,
		},
		{
			name:     "aws cli version 2.7.1",
			version:  "aws-cli/2.7.1 Python/3.7.4 Darwin/19.4.0 botocore/2.7.1",
			expected: nil,
		},
		{
			name:     "aws cli version 3.0.0",
			version:  "aws-cli/3.0.0 Python/3.7.4 Darwin/19.4.0 botocore/3.0.0",
			expected: fmt.Errorf("aws-cli version 3.0.0 is unsupported"),
		},
		{
			name:     "aws cli version 0.9.0",
			version:  "aws-cli/0.9.0 Python/3.7.4 Darwin/19.4.0 botocore/0.9.0",
			expected: fmt.Errorf("aws-cli version 0.9.0 is unsupported"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// act
			version, err := k8s.DefaultAwsCliVersionValidator.Parse(tc.version)

			if err == nil {
				err = k8s.DefaultAwsCliVersionValidator.Validate(version)
			}

			// assert
			assert.Equal(t, tc.expected, err)
		})
	}
}
