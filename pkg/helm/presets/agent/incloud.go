package agent

import "fmt"

func InCloudPreset(environmentId string) map[string]interface{} {
	valuesOverride := make(map[string]interface{})

	valuesOverride["backend"] = map[string]interface{}{
		"enabled": false,
	}

	// Replace placeholder with environmentId using fmt.Sprintf
	groundcoverURL := fmt.Sprintf("https://api-otel-http.%s.platform.grcv.io", environmentId)

	valuesOverride["global"] = map[string]interface{}{
		"logs": map[string]interface{}{
			"overrideUrl": groundcoverURL,
		},
		"otlp": map[string]interface{}{
			"overrideHttpURL": groundcoverURL,
			"overrideGrpcURL": fmt.Sprintf("api-otel-grpc.%s.platform.grcv.io:443", environmentId),
		},
	}

	valuesOverride["shepherd"] = map[string]interface{}{
		"overrideGrpcURL": fmt.Sprintf("api-shepherd-grpc.%s.platform.grcv.io:443", environmentId),
		"overrideHttpURL": fmt.Sprintf("https://api-shepherd-http.%s.platform.grcv.io", environmentId),
	}

	valuesOverride["custom-metrics"] = map[string]interface{}{
		"enabled": true,
		"remoteWriteUrls": []string{
			fmt.Sprintf("https://api-victoria-grpc.%s.platform.grcv.io/api/v1/write?apikey=$(API_KEY)", environmentId),
		},
		"extraArgs": map[string]interface{}{
			"remoteWrite.tlsInsecureSkipVerify": true,
		},
	}

	return valuesOverride
}
