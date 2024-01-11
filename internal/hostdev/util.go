package hostdev

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

func sanitizeName(path string) string {
	sanitizeChar := func(r rune) rune {
		switch {
		case r >= 'A' && r <= 'Z':
			return r
		case r >= 'a' && r <= 'z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '_':
			return r
		case r == '-':
			return r
		}
		return '_'
	}
	return strings.Map(sanitizeChar, path)
}

func readProcLocks(procLocksPath string) (map[uint64]int64, error) {
	f, err := os.Open(procLocksPath)
	if err != nil {
		return nil, err
	}

	var foundLocks = make(map[uint64]int64)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\n")
		for i := range parts {
			fields := strings.Fields(parts[i])
			// Another process try to setting lock to already locked file
			if fields[1] == "->" {
				continue
			}
			pid, err := strconv.ParseInt(fields[4], 10, 64)
			if err != nil {
				continue
			}
			subfields := strings.Split(fields[5], ":")
			inode, err := strconv.ParseInt(subfields[2], 10, 64)
			if err != nil {
				continue
			}
			foundLocks[uint64(inode)] = pid
		}
	}
	return foundLocks, nil
}

func readDevDirectory(dirToList string, allowedRecursions uint8) (files []string, err error) {
	fType, err := os.Stat(dirToList)
	if err != nil {
		return nil, err
	}

	if !fType.IsDir() {
		return nil, nil
	}

	f, err := os.Open(dirToList)
	if err != nil {
		return nil, err
	}
	files, err = f.Readdirnames(-1)

	if err != nil {
		f.Close()
		return nil, err
	}

	_ = f.Close()

	var foundFiles []string

	for _, subDir := range files {
		foundFiles = append(foundFiles, subDir)
		if allowedRecursions > 0 {
			filesDir, err := readDevDirectory(filepath.Join(dirToList, subDir), allowedRecursions-1)
			if err == nil {
				for _, fileName := range filesDir {
					foundFiles = append(foundFiles, filepath.Join(subDir, fileName))
				}
			}
		}
	}

	return foundFiles, nil
}

func matchDiskPattern(listDisks []string, pattern string) ([]string, error) {
	var found []string

	for _, file := range listDisks {
		res, err := regexp.MatchString(pattern, file)
		if err != nil {
			return nil, err
		}
		if res {
			found = append(found, file)
		}
	}
	return found, nil
}

func readInode(diskPath string) (uint64, error) {
	stat := syscall.Stat_t{}
	err := syscall.Stat(diskPath, &stat)
	if err != nil {
		return 0, err
	}
	return stat.Ino, nil
}
