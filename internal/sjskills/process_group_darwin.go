package sjskills

import "golang.org/x/sys/unix"

func processGroupActive(group int) (bool, error) {
	processes, err := unix.SysctlKinfoProcSlice("kern.proc.pgrp", group)
	if err != nil {
		return true, err
	}
	for _, process := range processes {
		// SZOMB has released its address space and cannot retain staging access;
		// its eventual reaping is the parent's responsibility.
		if process.Proc.P_stat != 5 {
			return true, nil
		}
	}
	return false, nil
}
