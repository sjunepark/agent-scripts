//go:build !windows

package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

type signalPromptWriter struct {
	ready chan struct{}
	once  sync.Once
}

func (w *signalPromptWriter) Write(data []byte) (int, error) {
	if strings.Contains(string(data), "[y/N]") {
		w.once.Do(func() { close(w.ready) })
	}
	return len(data), nil
}

func TestCLISignalsDuringConfirmation(t *testing.T) {
	for _, mode := range []string{"interrupt-twice", "terminate-cleanup"} {
		t.Run(mode, func(t *testing.T) {
			f := newStatusCLIFixture(t, "version = 1\nprofiles = [\"go\"]\n")
			stage := t.TempDir()
			cmd, _, _ := f.command(t, map[string]string{"TMPDIR": stage}, "--no-status-check", "apply")
			input, writer, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer input.Close()
			defer writer.Close()
			cmd.Stdin = input
			cmd.Stdout = io.Discard
			prompt := &signalPromptWriter{ready: make(chan struct{})}
			cmd.Stderr = prompt
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			defer cmd.Process.Kill()
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			select {
			case <-prompt.ready:
			case err := <-done:
				t.Fatalf("exit before prompt: %v", err)
			case <-time.After(10 * time.Second):
				t.Fatal("no confirmation prompt")
			}
			sig := os.Interrupt
			if mode == "terminate-cleanup" {
				sig = syscall.SIGTERM
			}
			if err := cmd.Process.Signal(sig); err != nil {
				t.Fatal(err)
			}
			// The input pipe deliberately remains open. Cancellation must keep cleanup
			// possible, and a repeated interrupt must still be able to terminate.
			select {
			case err := <-done:
				if cmd.ProcessState.Sys().(syscall.WaitStatus).Signaled() {
					t.Fatalf("first signal bypassed cleanup: %v", err)
				}
				return
			case <-time.After(200 * time.Millisecond):
			}
			if mode == "interrupt-twice" {
				if err := cmd.Process.Signal(os.Interrupt); err != nil {
					t.Fatal(err)
				}
			} else if _, err := io.WriteString(writer, "n\n"); err != nil {
				t.Fatal(err)
			}
			select {
			case <-done:
			case <-time.After(3 * time.Second):
				t.Fatal("command ignored cancellation/repeated signal")
			}
			if mode == "terminate-cleanup" {
				if cmd.ProcessState.Sys().(syscall.WaitStatus).Signaled() {
					t.Fatal("SIGTERM bypassed cleanup")
				}
				entries, err := os.ReadDir(stage)
				if err != nil || len(entries) != 0 {
					t.Fatalf("staging retained: %v %v", entries, err)
				}
			}
			if _, err := os.Stat(filepath.Join(f.project, ".agents", "skills")); !os.IsNotExist(err) {
				t.Fatalf("cancelled confirmation mutated skills: %v", err)
			}
		})
	}
}

func TestDefaultStatusSignalCleanup(t *testing.T) {
	for _, args := range [][]string{nil, {"status"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			f := newStatusCLIFixture(t, "version = 1\nprofiles = [\"go\"]\n")
			cmd, out, errout := f.command(t, map[string]string{"SJSKILLS_FAKE_BLOCK": "1"}, args...)
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			defer cmd.Process.Kill()
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			deadline := time.NewTimer(10 * time.Second)
			defer deadline.Stop()
			tick := time.NewTicker(10 * time.Millisecond)
			defer tick.Stop()
			waiting := true
			for waiting {
				select {
				case <-tick.C:
					if len(f.calls(t)) >= 2 {
						waiting = false
					}
				case err := <-done:
					t.Fatalf("exit before refresh: %v", err)
				case <-deadline.C:
					t.Fatal("refresh did not start")
				}
			}
			if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-done:
				if err == nil || cmd.ProcessState.ExitCode() != 2 || !strings.Contains(errout.String(), "cancelled") {
					t.Fatalf("cancel %v %q %q", err, out.String(), errout.String())
				}
			case <-time.After(5 * time.Second):
				t.Fatal("status cancellation exceeded cleanup budget")
			}
			entries, err := os.ReadDir(f.stage)
			if err != nil || len(entries) != 0 {
				t.Fatalf("staging retained: %v %v", entries, err)
			}
			for _, path := range []string{filepath.Join(f.project, ".agents"), filepath.Join(f.project, ".claude"), filepath.Join(f.home, ".agents"), filepath.Join(f.home, ".claude")} {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatalf("status mutated %s: %v", path, err)
				}
			}
		})
	}
}
