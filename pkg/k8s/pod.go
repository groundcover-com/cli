package k8s

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ContainerStatus struct {
	Name         string            `json:"name,omitempty"`
	State        v1.ContainerState `json:"state,omitempty"`
	Ready        bool              `json:"ready,omitempty"`
	RestartCount int32             `json:"restartCount,omitempty"`
	Started      *bool             `json:"started,omitempty"`
}

type PodStatus struct {
	Phase                  string
	Conditions             []v1.PodCondition `json:"conditions,omitempty"`
	Message                string            `json:"message,omitempty"`
	Reason                 string            `json:"reason,omitempty"`
	StartTime              *metav1.Time      `json:"startTime,omitempty"`
	InitContainersStatuses []ContainerStatus `json:"initContainersStatuses,omitempty"`
	ContainerStatuses      []ContainerStatus `json:"containerStatuses,omitempty"`
}

func BuildPodStatus(pod v1.Pod) PodStatus {
	initContainerStatuses := make([]ContainerStatus, 0, len(pod.Status.InitContainerStatuses))
	for _, initContainerStatus := range pod.Status.InitContainerStatuses {
		initContainerStatuses = append(initContainerStatuses, ContainerStatus{
			Name:         initContainerStatus.Name,
			State:        initContainerStatus.State,
			Ready:        initContainerStatus.Ready,
			RestartCount: initContainerStatus.RestartCount,
			Started:      initContainerStatus.Started,
		})
	}

	containerStatuses := make([]ContainerStatus, 0, len(pod.Status.ContainerStatuses))
	for _, containerStatus := range pod.Status.ContainerStatuses {
		containerStatuses = append(containerStatuses, ContainerStatus{
			Name:         containerStatus.Name,
			State:        containerStatus.State,
			Ready:        containerStatus.Ready,
			RestartCount: containerStatus.RestartCount,
			Started:      containerStatus.Started,
		})
	}

	return PodStatus{
		Phase:                  string(pod.Status.Phase),
		Conditions:             pod.Status.Conditions,
		Message:                pod.Status.Message,
		Reason:                 pod.Status.Reason,
		StartTime:              pod.Status.StartTime,
		InitContainersStatuses: initContainerStatuses,
		ContainerStatuses:      containerStatuses,
	}
}
