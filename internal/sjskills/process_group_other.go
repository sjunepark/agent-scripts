//go:build !windows && !linux && !darwin

package sjskills

import (
	"errors"
	"syscall"
)

func processGroupActive(group int) (bool, error) {
	err := syscall.Kill(-group, 0)
	if errors.Is(err, syscall.ESRCH) {
		return false, nil
	}
	return true, err
}
