package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/ydb-platform/ydb-disk-manager/internal/hostdev"
	"github.com/ydb-platform/ydb-disk-manager/pkg/api"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type Config struct {
	DiskMatch      string        `yaml:"diskMatch"`
	HostProcPath   string        `yaml:"hostProcPath"`
	UpdateInterval time.Duration `yaml:"updateInterval"`
	DeviceCount    uint          `yaml:"deviceCount"`
}

var cfg Config
var confFileName string

func usage() {
	fmt.Fprintf(os.Stderr, "usage: ydb-disk-manager\n")
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	klog.InitFlags(nil)
	flag.StringVar(&confFileName, "config", "config/conf.yaml", "set the configuration file to use")
	flag.Usage = usage
	flag.Parse()

	defer klog.Flush()
	klog.V(0).Info("Loading ydb-disk-manager")

	// Setting up the disks to check
	klog.V(0).Infof("Reading configuration file %s", confFileName)
	yamlFile, err := ioutil.ReadFile(confFileName)
	if err != nil {
		klog.Fatalf("Reading configuration file failed with: %s", err)
	}
	cfg.DeviceCount = 1 // Default
	err = yaml.Unmarshal(yamlFile, &cfg)
	if err != nil {
		klog.Fatal("Unmarshal: %v", err)
		os.Exit(-1)
	}
	klog.V(0).Infof("Applied configuration: %v", cfg)

	klog.V(0).Info("Starting FS watcher.")
	watcher, err := newFSWatcher(pluginapi.DevicePluginPath)
	if err != nil {
		klog.Error("Failed to create FS watcher.")
		os.Exit(1)
	}
	defer watcher.Close()

	klog.V(0).Info("Starting OS watcher.")
	sigs := newOSWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	klog.V(0).Info("Starting /healthz HTTP handler on port :8080")
	http.HandleFunc("/healthz", healthHandler)
	go func() {
		if err = http.ListenAndServe(":8080", nil); err != nil {
			klog.Error("Failed to create HTTP health server")
			os.Exit(1)
		}
	}()

	runWatcherLoop(watcher, sigs)
}

func runWatcherLoop(watcher *fsnotify.Watcher, sigs chan os.Signal) {
	var diskManager *hostdev.DiskManager
	var devicePluginInstance *api.DevicePlugin
	ticker := time.NewTicker(cfg.UpdateInterval)
	tickerCh := make(chan bool)
	defer close(tickerCh)
	restart := true
	defer ticker.Stop()
	for {
		if restart {
			if devicePluginInstance != nil {
				devicePluginInstance.Stop()
			}

			diskManager = hostdev.NewDiskmanager(tickerCh)
			devicePluginInstance = api.NewDevicePlugin(diskManager, cfg.DeviceCount)

			if err := diskManager.UpdateDisks(cfg.DiskMatch, "/dev"); err != nil {
				klog.V(0).Infof("Failed to update disks: %v", err)
				time.Sleep(cfg.UpdateInterval)
				continue
			}

			if err := devicePluginInstance.Serve(); err != nil {
				klog.V(0).Info("Could not contact Kubelet, retrying. Did you enable the device plugin feature gate?")
				time.Sleep(cfg.UpdateInterval)
				continue
			}

			if err := diskManager.UpdateLocks(cfg.HostProcPath); err != nil {
				klog.V(0).Infof("Failed to update locks: %v", err)
				time.Sleep(cfg.UpdateInterval)
				continue
			}

			restart = false
		}

		select {
		case <-tickerCh:
			klog.V(2).Infof("Resetting ticker to updateInterval: %s", cfg.UpdateInterval)
			ticker.Reset(cfg.UpdateInterval)

		case event := <-watcher.Events:
			if event.Name == pluginapi.KubeletSocket && event.Op&fsnotify.Create == fsnotify.Create {
				klog.V(0).Infof("inotify: %s created, restarting.", pluginapi.KubeletSocket)
				restart = true
			}

		case <-ticker.C:
			if err := diskManager.UpdateDisks(cfg.DiskMatch, "/dev"); err != nil {
				klog.V(0).Infof("Failed to update disks: %v", err)
				restart = true
				break
			}
			if err := diskManager.UpdateLocks(cfg.HostProcPath); err != nil {
				klog.V(0).Infof("Failed to update locks: %v", err)
				restart = true
				break
			}

		case err := <-watcher.Errors:
			klog.V(0).Infof("inotify: %s", err)

		case s := <-sigs:
			switch s {
			case syscall.SIGHUP:
				klog.V(0).Info("Received SIGHUP, restarting.")
				restart = true
			default:
				klog.V(0).Infof("Received signal \"%v\", shutting down.", s)
				if devicePluginInstance != nil {
					devicePluginInstance.Stop()
				}
				return
			}
		}
	}
}

func newFSWatcher(files ...string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		err = watcher.Add(f)
		if err != nil {
			watcher.Close()
			return nil, err
		}
	}

	return watcher, nil
}

func newOSWatcher(sigs ...os.Signal) chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, sigs...)

	return sigChan
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if err := api.LivenessProbe(api.SocketPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Server is healthy")
}
