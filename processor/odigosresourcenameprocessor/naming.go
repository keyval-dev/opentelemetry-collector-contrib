package odigosresourcenameprocessor

import (
	"context"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ContainerDetails struct {
	PodName         string
	PodNamespace    string
	ContainersInPod int
	ContainerName   string
}

type NameStrategy interface {
	GetName(containerDetails *ContainerDetails) string
}

type NameFromOwner struct {
	kc     kubernetes.Interface
	logger *zap.Logger
}

func (n *NameFromOwner) GetName(containerDetails *ContainerDetails) string {
	if containerDetails.ContainersInPod > 1 {
		return containerDetails.ContainerName
	}

	name, err := n.getNameByOwner(containerDetails)
	if err != nil {
		n.logger.Error("Failed to get name by owner, using pod name", zap.Error(err))
		return containerDetails.PodName
	}

	return name
}

func (n *NameFromOwner) getNameByOwner(containerDetails *ContainerDetails) (string, error) {
	pod, err := n.kc.CoreV1().Pods(containerDetails.PodNamespace).
		Get(context.Background(), containerDetails.PodName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	ownerRefs := pod.GetOwnerReferences()
	for _, ownerRef := range ownerRefs {
		if ownerRef.Kind == "ReplicaSet" {
			rs, err := n.kc.AppsV1().ReplicaSets(pod.Namespace).Get(context.Background(), ownerRef.Name, metav1.GetOptions{})
			if err != nil {
				return "", err
			}

			ownerRefs = rs.GetOwnerReferences()
			for _, ownerRef := range ownerRefs {
				if ownerRef.Kind == "Deployment" {
					return ownerRef.Name, nil
				}
			}
		} else if ownerRef.Kind == "StatefulSet" || ownerRef.Kind == "DaemonSet" {
			return ownerRef.Name, nil
		}
	}

	return containerDetails.PodName, nil
}
