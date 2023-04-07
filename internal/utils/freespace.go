package utils

import (
	"syscall"
)

func FreeSpace(directory string) (int, error) {
	// Get free space on download area
	var fs syscall.Statfs_t
	err := syscall.Statfs(directory, &fs)
	if err != nil {
		return 0, err
	}

	freeBytes := fs.Bavail * uint64(fs.Bsize)

	return int(freeBytes), nil
}
