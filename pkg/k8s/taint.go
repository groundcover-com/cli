package k8s

import (
	"encoding/json"
	"strings"

	v1 "k8s.io/api/core/v1"
)

const (
	BUILD_IN_TAINTS_PREFIX = "node.kubernetes.io"
)

var (
	taintsSet = make(map[string]struct{})
)

func isBuildInTaint(taint v1.Taint) bool {
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

func unmarshalTaintToToleration(taintString string) (v1.Toleration, error) {
	var err error

	toleration := v1.Toleration{
		Operator: "Equal",
	}

	if err = json.Unmarshal([]byte(taintString), &toleration); err != nil {
		return toleration, err
	}

	return toleration, nil
}
