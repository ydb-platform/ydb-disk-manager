package api

import (
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"k8s.io/klog/v2"

	"github.com/ydb-platform/ydb-disk-manager/internal/hostdev"
	"github.com/ydb-platform/ydb-disk-manager/proto/locks"
	pb "github.com/ydb-platform/ydb-disk-manager/proto/locks"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	envDisablePreStartContainer  = "DP_DISABLE_PRE_START_CONTAINER"
	defaultKubeletConnectTimeout = 5 * time.Second
	maxDialTimeout               = 10 * time.Second
	oneTimeDialTimeout           = 3 * time.Second
	partLabelPath                = "/dev/disk/by-partlabel"
	AnnotationName               = "io.kubernetes.kubelet.device-plugin"
	ResourceName                 = "ydb-disk-manager/hostdev"
	DeviceName                   = "disk-manager-hostdev"
	SocketPath                   = pluginapi.DevicePluginPath + DeviceName + ".sock"
)

// DevicePlugin implements the Kubernetes device plugin API
type DevicePlugin struct {
	devs   []*pluginapi.Device
	stop   chan interface{}
	health chan *pluginapi.Device
	server *grpc.Server

	disks *hostdev.DiskManager
}

// NewDevicePlugin returns an initialized DevicePlugin
func NewDevicePlugin(hostDisks *hostdev.DiskManager) *DevicePlugin {
	return &DevicePlugin{
		devs:   getDevices(),
		stop:   make(chan interface{}),
		health: make(chan *pluginapi.Device),

		disks: hostDisks,
	}
}

// Start the gRPC server of the device plugin
func (plugin *DevicePlugin) Start() error {
	klog.V(0).Infof("Removing file %s", SocketPath)
	if err := os.Remove(SocketPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	klog.V(0).Info("Releasing locks from host disks")
	if err := plugin.disks.ReleaseLocks(); err != nil {
		return err
	}

	klog.V(0).Infof("Create listen unix socket %s", SocketPath)
	sock, err := net.Listen("unix", SocketPath)
	if err != nil {
		return err
	}

	klog.V(0).Info("Starting GRPC server")
	plugin.server = grpc.NewServer([]grpc.ServerOption{}...)
	pluginapi.RegisterDevicePluginServer(plugin.server, plugin)
	pb.RegisterLocksServer(plugin.server, plugin)

	go plugin.server.Serve(sock)

	return nil
}

// Stop the gRPC server
func (plugin *DevicePlugin) Stop() error {
	klog.V(0).Infof("Stopping device %s", ResourceName)
	if plugin.server == nil {
		return nil
	}

	klog.V(0).Info("Releasing locks from host disks")
	if err := plugin.disks.ReleaseLocks(); err != nil {
		return err
	}

	klog.V(0).Infof("Stop server with socket %s", SocketPath)
	plugin.server.Stop()
	plugin.server = nil
	close(plugin.stop)

	klog.V(0).Infof("Remove file %s", SocketPath)
	if err := os.Remove(SocketPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// Register the device plugin for the given resourceName with Kubelet.
func (plugin *DevicePlugin) Register(kubeletEndpoint string) error {
	conn, err := dial(kubeletEndpoint, defaultKubeletConnectTimeout)
	if err != nil {
		return err
	}

	defer func() {
		err := conn.Close()
		if err != nil {
			klog.Error(err)
		}
	}()

	client := pluginapi.NewRegistrationClient(conn)
	opts := &pluginapi.DevicePluginOptions{PreStartRequired: true}
	disablePreStartContainer := strings.ToLower(os.Getenv(envDisablePreStartContainer))
	if disablePreStartContainer == "true" {
		opts = &pluginapi.DevicePluginOptions{PreStartRequired: false}
	}
	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(SocketPath),
		ResourceName: ResourceName,
		Options:      opts,
	}

	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return err
	}
	return nil
}

// ListAndWatch lists devices and update that list according to the health status
func (plugin *DevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	if err := s.Send(&pluginapi.ListAndWatchResponse{Devices: plugin.devs}); err != nil {
		klog.Errorf("ListAndWatch %s error: cannot update device states: %v\n", ResourceName, err)
		plugin.Stop()
		return err
	}
	for {
		select {
		case <-s.Context().Done():
			plugin.Stop()
			if err := s.Context().Err(); err != nil {
				klog.Errorf("Connection closed unexpectedly by client: %v", err)
				return err
			}
			return nil
		case <-plugin.stop:
			return nil
		case d := <-plugin.health:
			klog.Errorf("%s device health changed", ResourceName)
			d.Health = pluginapi.Unhealthy
			if err := s.Send(&pluginapi.ListAndWatchResponse{Devices: plugin.devs}); err != nil {
				klog.Errorf("ListAndWatch %s error: cannot update device states: %v\n", ResourceName, err)
				plugin.Stop()
				return err
			}
		}
	}
}

// Allocate which return list of devices.
func (plugin *DevicePlugin) Allocate(ctx context.Context, reqs *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	devs := plugin.devs
	responses := pluginapi.AllocateResponse{}
	var devices []*pluginapi.DeviceSpec

	// This is the important part - we push all the disk devices we know about
	// instead of the meta-device 'ydb-disk-manager/hostdev'
	for _, diskPath := range plugin.disks.DiskFilenames {
		devices = append(devices,
			&pluginapi.DeviceSpec{
				ContainerPath: diskPath,
				HostPath:      diskPath,
				Permissions:   "rw",
			},
		)
	}

	for _, req := range reqs.ContainerRequests {
		response := pluginapi.ContainerAllocateResponse{
			Devices: devices,
			Mounts: []*pluginapi.Mount{{
				ContainerPath: partLabelPath,
				HostPath:      partLabelPath,
				ReadOnly:      false,
			}},
			Annotations: map[string]string{AnnotationName: DeviceName},
		}

		for _, id := range req.DevicesIDs {
			if !deviceExists(devs, id) {
				return nil, fmt.Errorf("invalid allocation request: unknown device: %s", id)
			}
		}

		responses.ContainerResponses = append(responses.ContainerResponses, &response)
	}

	return &responses, nil
}

func (plugin *DevicePlugin) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	klog.V(2).Info("Executing PreStartContainer device plugin hook")
	if err := plugin.disks.SetLocks(); err != nil {
		return nil, err
	}
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (plugin *DevicePlugin) GetPreferredAllocation(context.Context, *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return &pluginapi.PreferredAllocationResponse{}, nil
}

// Serve starts the gRPC server and register the device plugin to Kubelet
func (plugin *DevicePlugin) Serve() error {
	err := plugin.Start()
	if err != nil {
		klog.Errorf("Could not start device plugin: %s", err)
		return err
	}
	klog.V(0).Infof("Starting to serve on: %s", SocketPath)

	err = plugin.Register(pluginapi.KubeletSocket)
	if err != nil {
		klog.Errorf("Could not register device plugin: %s", err)
		plugin.Stop()
		return err
	}
	klog.V(0).Info("Registered device plugin with Kubelet")

	return nil
}

func (plugin *DevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	disablePreStartContainer := strings.ToLower(os.Getenv(envDisablePreStartContainer))
	if disablePreStartContainer == "true" {
		return &pluginapi.DevicePluginOptions{PreStartRequired: false}, nil
	}
	return &pluginapi.DevicePluginOptions{PreStartRequired: true}, nil
}

func (plugin *DevicePlugin) SetLocks(context.Context, *locks.LocksRequest) (*locks.LocksResponse, error) {
	klog.V(2).Info("Received grpc request to make SetLocks")
	if err := plugin.disks.SetLocks(); err != nil {
		return nil, err
	}
	return &locks.LocksResponse{}, nil
}

func (plugin *DevicePlugin) ReleaseLocks(context.Context, *locks.LocksRequest) (*locks.LocksResponse, error) {
	klog.V(2).Info("Received grpc request to make ReleaseLocks")
	if err := plugin.disks.ReleaseLocks(); err != nil {
		return nil, err
	}
	return &locks.LocksResponse{}, nil
}
