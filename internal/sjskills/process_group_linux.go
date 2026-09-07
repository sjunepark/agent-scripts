package sjskills

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func processGroupActive(group int) (bool, error) {
	if err := syscall.Kill(-group, 0); errors.Is(err, syscall.ESRCH) {
		return false, nil
	}
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return true, err
	}
	for _, entry := range entries {
		if _, err := strconv.Atoi(entry.Name()); err != nil {
			continue
		}
		data, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "stat"))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return true, err
		}
		// comm is parenthesized and may itself contain spaces or parentheses.
		end := strings.LastIndexByte(string(data), ')')
		if end < 0 {
			continue
		}
		fields := strings.Fields(string(data[end+1:]))
		if len(fields) < 3 {
			continue
		}
		pgid, err := strconv.Atoi(fields[2])
		if err != nil {
			return true, err
		}
		if pgid == group && fields[0] != "Z" && fields[0] != "X" {
			return true, nil
		}
	}
	return false, nil
}
