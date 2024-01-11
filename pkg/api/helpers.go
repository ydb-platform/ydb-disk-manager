package api

import (
	"fmt"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func getDevices() []*pluginapi.Device {
	var devs []*pluginapi.Device
	devs = append(devs, &pluginapi.Device{
		ID:     strconv.Itoa(1),
		Health: pluginapi.Healthy,
	})

	return devs
}

func deviceExists(devs []*pluginapi.Device, id string) bool {
	for _, d := range devs {
		if d.ID == id {
			return true
		}
	}
	return false
}

// dial establishes the gRPC communication with the registered device plugin.
func dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	c, err := grpc.Dial(unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)

	if err != nil {
		return nil, err
	}

	return c, nil
}

// dial establishes the gRPC communication with the registered device plugin.
func LivenessProbe(socketPath string) error {
	now := time.Now()
	for time.Since(now) <= maxDialTimeout {
		conn, err := dial(socketPath, oneTimeDialTimeout)
		if err != nil {
			klog.Errorf("error when probing grpc server: %v", err)
			time.Sleep(time.Second)
			continue
		}
		conn.Close()
		return nil
	}
	return fmt.Errorf("Server socket %s is not ready", socketPath)
}
