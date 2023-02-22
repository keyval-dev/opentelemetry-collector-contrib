package odigosresourcenameprocessor

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"os"
	"sync"
)

const (
	nodeNameEnvVar = "NODE_NAME"
)

var (
	ErrNoDeviceFound = errors.New("no device found")
)

type NameResolver struct {
	kc            kubernetes.Interface
	logger        *zap.Logger
	kubelet       *kubeletClient
	mu            sync.RWMutex
	devicesToPods map[string]string
}

func (n *NameResolver) Resolve(deviceID string) (string, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	name, ok := n.devicesToPods[deviceID]
	if !ok {
		return "", ErrNoDeviceFound
	}

	return name, nil
}

func (n *NameResolver) updateDevicesToPods() error {
	allocations, err := n.kubelet.GetAllocations()
	if err != nil {
		return err
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	n.devicesToPods = allocations
	return nil
}

func (n *NameResolver) Start() error {
	n.logger.Info("Starting NameResolver ...")
	nn, ok := os.LookupEnv(nodeNameEnvVar)
	if !ok {
		return fmt.Errorf("env var %s is not set", nodeNameEnvVar)
	}

	w, err := n.kc.CoreV1().Pods("").Watch(context.Background(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nn),
	})
	if err != nil {
		n.logger.Error("Error watching pods", zap.Error(err))
		return err
	}

	go func() {
		for event := range w.ResultChan() {
			pod := event.Object.(*corev1.Pod)
			if event.Type == watch.Modified && pod.Status.Phase == corev1.PodRunning {
				if err := n.updateDevicesToPods(); err != nil {
					n.logger.Error("Error updating devices to pods", zap.Error(err))
				}
			}
		}
	}()

	return nil
}
