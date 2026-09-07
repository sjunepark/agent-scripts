//go:build !windows

package sjskills

import (
	"os"
	"syscall"
)

func openStatusFile(path string, writable bool) (*os.File, error) {
	flags := syscall.O_RDONLY
	if writable {
		flags = syscall.O_RDWR
	}
	// Nonblocking also prevents a raced FIFO from hanging an advisory check.
	fd, err := syscall.Open(path, flags|syscall.O_NOFOLLOW|syscall.O_NONBLOCK|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), path), nil
}
