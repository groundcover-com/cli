package k8s

import (
	"encoding/json"
	"strings"

	v1 "k8s.io/api/core/v1"
)

const (
	BUILD_IN_TAINTS_PREFIX = "node.kubernetes.io"
)

func GenerateTolerationsFromTaints(taintMarshaleds []string) ([]v1.Toleration, error) {
	var err error
	var tolerations []v1.Toleration

	for _, taintMarshaled := range taintMarshaleds {
		toleration := v1.Toleration{
			Operator: "Equal",
		}

		if err = json.Unmarshal([]byte(taintMarshaled), &toleration); err != nil {
			return nil, err
		}

		tolerations = append(tolerations, toleration)
	}

	return tolerations, nil
}

func isBuiltinTaint(taint v1.Taint) bool {
	return strings.HasPrefix(taint.Key, BUILD_IN_TAINTS_PREFIX)
}

func marshalTaint(taint v1.Taint) (string, error) {
	var err error

	var jsonByte []byte
	if jsonByte, err = json.Marshal(taint); err != nil {
		return "", err
	}

	return string(jsonByte), nil
}
