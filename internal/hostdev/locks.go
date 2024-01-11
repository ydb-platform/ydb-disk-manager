package hostdev

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"k8s.io/klog/v2"
)

func findDiskLocks(procPath string, diskFilenames []string, procLocks map[uint64]int64) map[uint64]string {
	var diskLocks = make(map[uint64]string)

	for lockIno, lockPid := range procLocks {
		fdDir := filepath.Join(procPath, strconv.FormatInt(lockPid, 10), "fd")
		files, _ := ioutil.ReadDir(fdDir)
		for _, f := range files {
			fdPath := filepath.Join(fdDir, f.Name())
			lockPath, err := os.Readlink(fdPath)
			if err != nil {
				klog.V(4).Infof("err: %v", err)
				continue
			}

			// search locks for disks
			for _, diskPath := range diskFilenames {
				if diskPath == lockPath {
					klog.V(4).Infof("Found lock with diskPath: %s equal lockPath: %s", diskPath, lockPath)
					diskIno, err := readInode(fdPath)
					if err != nil {
						continue
					}
					klog.V(4).Infof("Check diskIno: %d, lockIno: %d", diskIno, lockIno)
					if diskIno == lockIno {
						klog.V(2).Infof("Found lock with diskIno: %d equal lockIno: %d", diskIno, lockIno)
						diskLocks[lockIno] = lockPath
						break
					}
				}
			}
		}
	}
	return diskLocks
}

func setLock(diskPath string) (*os.File, error) {
	file, err := os.Open(diskPath)
	if err != nil {
		klog.Errorf("Error opening file: %s", diskPath)
		return nil, err
	}

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		klog.Errorf("Failed to lock on disk %s", diskPath)
		return nil, err
	}

	klog.V(0).Infof("Lock successfully set on host disk path %s", diskPath)
	return file, nil
}

func releaseLock(file *os.File) error {
	if file == nil {
		return fmt.Errorf("*os.File argument is nil")
	}

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_UN); err != nil {
		klog.Errorf("Failed to release lock from %s", file.Name())
		return err
	}

	klog.V(0).Infof("Lock successfully released from file: %s", file.Name())
	return file.Close()
}
