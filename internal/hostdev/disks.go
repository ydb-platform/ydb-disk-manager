package hostdev

import (
	"os"
	"path/filepath"
	"sync"

	"k8s.io/klog/v2"
)

type FileLock struct {
	mu    sync.RWMutex
	locks map[uint64]*os.File
}

type DiskManager struct {
	DiskFilenames []string
	DiskInodes    map[uint64]string
	fileLocks     FileLock
	tickerChan    chan<- bool
}

// NewSmarterDevicePlugin returns an initialized SmarterDevicePlugin
func NewDiskmanager(ch chan<- bool) *DiskManager {
	return &DiskManager{
		DiskFilenames: []string{},
		DiskInodes:    make(map[uint64]string),

		fileLocks:  FileLock{locks: make(map[uint64]*os.File)},
		tickerChan: ch,
	}
}

func (mgr *DiskManager) UpdateDisks(diskRegexp string, devDir string) error {
	existingDisks, err := readDevDirectory(devDir, 10)
	if err != nil {
		return err
	}

	matchedDisks, err := matchDiskPattern(existingDisks, diskRegexp)
	if err != nil {
		return err
	}

	// Check for new disks
	for _, diskToAppend := range matchedDisks {
		diskPath := filepath.Join("/dev", diskToAppend)
		diskIno, err := readInode(diskPath)
		if err != nil {
			return err
		}
		shouldAppend := true
		for _, existingDisk := range mgr.DiskFilenames {
			if diskPath == existingDisk {
				shouldAppend = false
				break
			}
		}
		if shouldAppend {
			mgr.DiskFilenames = append(mgr.DiskFilenames, diskPath)
			mgr.DiskInodes[diskIno] = diskPath
			klog.V(0).Infof("Detected new host disk %s with inode: %d", diskPath, diskIno)
		}
	}

	// Check for removed disks
	temp := mgr.DiskFilenames[:0]
	for _, existingDisk := range mgr.DiskFilenames {
		shouldRemove := true
		for _, diskToRemove := range matchedDisks {
			diskPath := filepath.Join("/dev", diskToRemove)
			if existingDisk == diskPath {
				temp = append(temp, existingDisk)
				shouldRemove = false
				break
			}
		}
		if shouldRemove {
			for diskIno, diskPath := range mgr.DiskInodes {
				if existingDisk == diskPath {
					delete(mgr.DiskInodes, diskIno)
					klog.V(0).Infof("Removed stale host disk %s with inode %d", diskPath, diskIno)
					break
				}
			}
		}
	}
	mgr.DiskFilenames = temp
	return nil
}

func (mgr *DiskManager) UpdateLocks(procPath string) error {
	mgr.fileLocks.mu.Lock()
	defer mgr.fileLocks.mu.Unlock()

	locksPath := filepath.Join(procPath, "locks")
	procLocks, err := readProcLocks(locksPath)
	if err != nil {
		return err
	}
	foundDiskLocks := findDiskLocks(procPath, mgr.DiskFilenames, procLocks)

	var containerLocks = make(map[string]uint64)
	var hostLocks = make(map[string]uint64)

	klog.V(5).Infof("mgr.DiskInodes: %v", mgr.DiskInodes)

	for diskIno, diskPath := range foundDiskLocks {
		_, exist := mgr.DiskInodes[diskIno]
		if !exist {
			// lock belongs to disk in container fs
			containerLocks[diskPath] = diskIno
		} else {
			// lock belongs to disk in host fs
			hostLocks[diskPath] = diskIno
		}
	}

	klog.V(5).Info("containerLocks map: %v", containerLocks)
	klog.V(5).Infof("hostLocks map: %v", hostLocks)

	for diskPath, _ := range containerLocks {
		if _, exist := hostLocks[diskPath]; !exist {
			klog.V(0).Infof("Lock exist in container, but not in host disk: %s", diskPath)

			inoFromHost := uint64(0)
			for ino, hostDiskPath := range mgr.DiskInodes {
				if hostDiskPath == diskPath {
					inoFromHost = ino
					break
				}
			}

			klog.V(0).Infof("Setting lock to host disk: %s, inode %d...", diskPath, inoFromHost)
			file, err := setLock(diskPath)
			if err != nil {
				return err
			}

			klog.V(3).Infof("setting host lock, inode:%v, lockfile:%v to fileLocks", inoFromHost, file)
			mgr.fileLocks.locks[inoFromHost] = file
		}
	}

	klog.V(5).Infof("mgr.fileLocks.locks map: %v", mgr.fileLocks.locks)

	for diskPath, diskIno := range hostLocks {
		if _, exist := containerLocks[diskPath]; !exist {
			if file, exist := mgr.fileLocks.locks[diskIno]; exist {
				klog.V(0).Infof("Lock exist in host, but not in container disk: %s", diskPath)
				klog.V(0).Infof("Releasing lock from host disk: %s, inode %d...", diskPath, diskIno)
				if err := releaseLock(file); err != nil {
					return err
				}
				delete(mgr.fileLocks.locks, diskIno)
			}
		}
	}

	return nil
}

func (mgr *DiskManager) SetLocks() error {
	mgr.fileLocks.mu.Lock()
	defer mgr.fileLocks.mu.Unlock()

	defer func() {
		go mgr.resetTicker()
	}()
	for _, diskPath := range mgr.DiskFilenames {
		diskIno, err := readInode(diskPath)
		if err != nil {
			klog.Error(err)
			return err
		}
		if _, exist := mgr.fileLocks.locks[diskIno]; exist {
			klog.V(0).Infof("Lock already exist on %s", diskPath)
			continue
		}
		file, err := setLock(diskPath)
		if err != nil {
			klog.Error(err)
			return err
		}
		mgr.fileLocks.locks[diskIno] = file
	}
	return nil
}

func (mgr *DiskManager) ReleaseLocks() error {
	mgr.fileLocks.mu.Lock()
	defer mgr.fileLocks.mu.Unlock()
	for diskIno, file := range mgr.fileLocks.locks {
		if err := releaseLock(file); err != nil && !os.IsNotExist(err) {
			return err
		}
		delete(mgr.fileLocks.locks, diskIno)
	}
	return nil
}

func (mgr *DiskManager) resetTicker() {
	mgr.tickerChan <- true
}
