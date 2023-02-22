package odigosresourcenameprocessor

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	podresourcesapi "k8s.io/kubelet/pkg/apis/podresources/v1alpha1"
	"regexp"
	"strings"
	"time"
)

var (
	socketDir  = "/var/lib/kubelet/pod-resources"
	socketPath = "unix://" + socketDir + "/kubelet.sock"

	connectionTimeout = 10 * time.Second
	ownerRegex        = regexp.MustCompile(`(?P<deployment_name>[a-z0-9]+(?:-[a-z0-9]+)*?)-[a-f0-9]{10}-[a-z0-9]+`)
)

type kubeletClient struct {
	conn *grpc.ClientConn
}

func NewKubeletClient() (*kubeletClient, error) {
	conn, err := connectToKubelet(socketPath)
	if err != nil {
		return nil, err
	}

	return &kubeletClient{
		conn: conn,
	}, nil
}

func (c *kubeletClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *kubeletClient) GetAllocations() (map[string]string, error) {
	pods, err := c.listPods()
	if err != nil {
		return nil, err
	}

	allocations := make(map[string]string)
	for _, pod := range pods.GetPodResources() {
		for _, container := range pod.Containers {
			for _, device := range container.Devices {
				for _, id := range device.DeviceIds {
					if strings.Contains(device.GetResourceName(), "odigos.io") {
						allocations[id] = calculateResourceName(pod.Name, len(pod.Containers), container.Name)
					}
				}
			}
		}
	}

	return allocations, nil
}

func calculateResourceName(podName string, containers int, containerName string) string {
	if containers > 1 {
		return containerName
	}

	match := ownerRegex.FindStringSubmatch(podName)
	if len(match) > 1 {
		return match[1]
	}

	return podName
}

func connectToKubelet(socket string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, socket, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())

	if err != nil {
		return nil, fmt.Errorf("failure connecting to %s: %v", socket, err)
	}

	return conn, nil
}

func (c *kubeletClient) listPods() (*podresourcesapi.ListPodResourcesResponse, error) {
	client := podresourcesapi.NewPodResourcesListerClient(c.conn)

	ctx, cancel := context.WithTimeout(context.Background(), connectionTimeout)
	defer cancel()

	resp, err := client.List(ctx, &podresourcesapi.ListPodResourcesRequest{})
	if err != nil {
		return nil, fmt.Errorf("failure getting pod resources %v", err)
	}

	return resp, nil
}
