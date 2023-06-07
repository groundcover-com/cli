package agent

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestInCloudPreset(t *testing.T) {
	environmentId := "example-environment"

	valuesOverride := InCloudPreset(environmentId)

	expected := map[string]interface{}{
		"backend": map[string]interface{}{
			"enabled": false,
		},
		"global": map[string]interface{}{
			"logs": map[string]interface{}{
				"overrideUrl": "https://api-otel-http.example-environment.platform.grcv.io",
			},
			"otlp": map[string]interface{}{
				"overrideHttpURL": "https://api-otel-http.example-environment.platform.grcv.io",
				"overrideGrpcURL": "api-otel-grpc.example-environment.platform.grcv.io:443",
			},
		},
		"shepherd": map[string]interface{}{
			"overrideGrpcURL": "api-shepherd-grpc.example-environment.platform.grcv.io:443",
			"overrideHttpURL": "https://api-shepherd-http.example-environment.platform.grcv.io",
		},
		"custom-metrics": map[string]interface{}{
			"enabled": true,
			"remoteWriteUrls": []string{
				"https://api-victoria-grpc.example-environment.platform.grcv.io/api/v1/write?apikey=$(API_KEY)",
			},
			"extraArgs": map[string]interface{}{
				"remoteWrite.tlsInsecureSkipVerify": true,
			},
		},
	}

	if diff := cmp.Diff(expected, valuesOverride); diff != "" {
		t.Errorf("valuesOverride does not match the expected map:\n%s", diff)
	}
}
