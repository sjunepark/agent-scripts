//go:build !windows

package sjskills

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type processTree struct{ cmd *exec.Cmd }

func prepareProcessTree(cmd *exec.Cmd) (*processTree, error) {
	tree := &processTree{cmd: cmd}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = tree.terminate
	return tree, nil
}
func (p *processTree) started() error { return nil }
func (p *processTree) close()         {}
func (p *processTree) terminate() error {
	if p.cmd.Process == nil {
		return os.ErrProcessDone
	}
	err := syscall.Kill(-p.cmd.Process.Pid, syscall.SIGKILL)
	if errors.Is(err, syscall.ESRCH) {
		return os.ErrProcessDone
	}
	return err
}
func (p *processTree) finish() error {
	err := p.terminate()
	if errors.Is(err, os.ErrProcessDone) {
		return nil
	}
	// Darwin may return EPERM while a previously killed group is disappearing.
	// Signal success alone is insufficient, and a signal error alone does not
	// prove live descendants: in both cases verify the remaining process set.
	signalErr := err
	deadline := time.Now().Add(processTreeCleanupBudget)
	for {
		active, err := processGroupActive(p.cmd.Process.Pid)
		if err == nil && !active {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("group active=%t: %w", active, errors.Join(errProcessTreeActive, signalErr, err))
		}
		time.Sleep(5 * time.Millisecond)
	}
}
