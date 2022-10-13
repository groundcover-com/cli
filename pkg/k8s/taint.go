package k8s

import (
	"encoding/json"
	"strings"

	"golang.org/x/exp/maps"
	v1 "k8s.io/api/core/v1"
)

const (
	BUILTIN_TAINTS_PREFIX = "node.kubernetes.io"
)

type TolerationManager struct {
	TaintedNodes []*IncompatibleNode
}

func (manager TolerationManager) GetTaints() ([]string, error) {
	var err error

	taintsSet := make(map[string]struct{})

	for _, taintedNode := range manager.TaintedNodes {
		for _, taint := range taintedNode.Taints {
			if isBuiltinTaint(taint) {
				continue
			}

			var taintMarshaled string
			if taintMarshaled, err = manager.marshalTaint(taint); err != nil {
				return nil, err
			}

			if _, exists := taintsSet[taintMarshaled]; !exists {
				taintsSet[taintMarshaled] = struct{}{}
			}
		}
	}

	return maps.Keys(taintsSet), nil
}

func (manager TolerationManager) GetTolerations(allowedTaints []string) ([]v1.Toleration, error) {
	var err error
	var tolerations []v1.Toleration

	for _, taintMarshaled := range allowedTaints {
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

func (manager TolerationManager) GetTolerableNodes(allowedTaints []string) ([]*NodeSummary, error) {
	var err error
	var tolerableNodes []*NodeSummary

	if len(allowedTaints) == 0 {
		return tolerableNodes, nil
	}

	for _, taintedNode := range manager.TaintedNodes {
		var incompatibleNode bool
		for _, taint := range taintedNode.Taints {
			if isBuiltinTaint(taint) {
				continue
			}

			var taintMarshaled string
			if taintMarshaled, err = manager.marshalTaint(taint); err != nil {
				return nil, err
			}

			for _, allowedTaint := range allowedTaints {
				if taintMarshaled != allowedTaint {
					incompatibleNode = true
					break
				}
			}
		}

		if incompatibleNode {
			continue
		}

		tolerableNodes = append(tolerableNodes, taintedNode.NodeSummary)
	}

	return tolerableNodes, nil
}

func (validator TolerationManager) marshalTaint(taint v1.Taint) (string, error) {
	var err error

	var jsonByte []byte
	if jsonByte, err = json.Marshal(taint); err != nil {
		return "", err
	}

	return string(jsonByte), nil
}

func isBuiltinTaint(taint v1.Taint) bool {
	return strings.HasPrefix(taint.Key, BUILTIN_TAINTS_PREFIX)
}
