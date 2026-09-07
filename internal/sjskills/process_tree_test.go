package sjskills

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The descendant retains an open staging handle and writes continuously. Its
// parent deliberately does not wait for it and may exit before cancellation.
func TestProcessTreeFixture(t *testing.T) {
	mode := os.Getenv("SJSKILLS_PROCESS_TREE_FIXTURE")
	if mode == "" {
		return
	}
	ready := os.Getenv("SJSKILLS_PROCESS_TREE_READY")
	path := os.Getenv("SJSKILLS_PROCESS_TREE_WRITES")
	if mode == "writer" {
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
		if err != nil {
			os.Exit(3)
		}
		if err := os.WriteFile(ready, []byte("ready"), 0o600); err != nil {
			os.Exit(4)
		}
		for {
			if _, err := file.WriteString("x"); err != nil {
				os.Exit(5)
			}
			if err := file.Sync(); err != nil {
				os.Exit(6)
			}
			time.Sleep(time.Millisecond)
		}
	}
	child := exec.Command(os.Args[0], "-test.run=^TestProcessTreeFixture$")
	child.Env = append(os.Environ(), "SJSKILLS_PROCESS_TREE_FIXTURE=writer")
	child.Stdout = os.Stdout
	child.Stderr = os.Stderr
	if err := child.Start(); err != nil {
		os.Exit(7)
	}
	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		if time.Now().After(deadline) {
			os.Exit(8)
		}
		time.Sleep(time.Millisecond)
	}
	if mode == "exit" {
		os.Exit(0)
	}
	for {
		time.Sleep(time.Hour)
	}
}
func TestProcessTreeStopsDescendantsBeforeReturning(t *testing.T) {
	for _, mode := range []string{"cancel", "timeout", "exit"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			ready := filepath.Join(root, "ready")
			writes := filepath.Join(root, "writes")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if mode == "timeout" {
				ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
			}
			runner := boundedExecRunner{limits: MaterializerLimits{MaxStdoutBytes: 1024, MaxStderrBytes: 1024}}
			done := make(chan error, 1)
			go func() {
				_, err := runner.Run(ctx, os.Args[0], []string{"-test.run=^TestProcessTreeFixture$"}, append(os.Environ(), "SJSKILLS_PROCESS_TREE_FIXTURE="+mode, "SJSKILLS_PROCESS_TREE_READY="+ready, "SJSKILLS_PROCESS_TREE_WRITES="+writes))
				done <- err
			}()
			deadline := time.Now().Add(8 * time.Second)
			for {
				if _, err := os.Stat(ready); err == nil {
					break
				}
				select {
				case err := <-done:
					t.Fatalf("runner exited before descendant was ready: %v", err)
				default:
				}
				if time.Now().After(deadline) {
					t.Fatal("descendant never became ready")
				}
				time.Sleep(5 * time.Millisecond)
			}
			stopped := time.Now()
			if mode == "cancel" {
				cancel()
			}
			select {
			case err := <-done:
				if errors.Is(err, errProcessTreeActive) {
					t.Fatal(err)
				}
				if mode != "exit" && err == nil {
					t.Fatal("cancelled process succeeded")
				}
			case <-time.After(6 * time.Second):
				t.Fatal("process tree did not stop within budget")
			}
			if mode == "cancel" && time.Since(stopped) > processTreeCleanupBudget+time.Second {
				t.Fatal("cancellation exceeded cleanup budget")
			}
			before, err := os.ReadFile(writes)
			if err != nil {
				t.Fatal(err)
			}
			time.Sleep(50 * time.Millisecond)
			after, err := os.ReadFile(writes)
			if err != nil {
				t.Fatal(err)
			}
			if string(before) != string(after) {
				t.Fatal("descendant retained staging access after runner returned")
			}
			if err := os.RemoveAll(root); err != nil {
				t.Fatalf("staging still held open: %v", err)
			}
		})
	}
}
func TestMaterializerStopsDescendantBeforeRemovingStaging(t *testing.T) {
	for _, mode := range []string{"cancel", "timeout"} {
		t.Run(mode, func(t *testing.T) {
			parent := t.TempDir()
			ready := filepath.Join(t.TempDir(), "ready")
			writes := ""
			runner := defaultMaterializeRunner()
			runner.invoke = func(ctx context.Context, _ string, args, env []string) (ProcessResult, error) {
				if len(args) == 1 {
					return ProcessResult{Stdout: []byte("bunx 1\n")}, nil
				}
				if len(args) == 2 {
					return ProcessResult{Stdout: []byte(SkillsCLIVersion + "\n")}, nil
				}
				writes = filepath.Join(envValue(env, "HOME"), "held-staging-file")
				bounded := boundedExecRunner{limits: MaterializerLimits{MaxStdoutBytes: 1024, MaxStderrBytes: 1024}}
				return bounded.Run(ctx, os.Args[0], []string{"-test.run=^TestProcessTreeFixture$"}, append(os.Environ(), "SJSKILLS_PROCESS_TREE_FIXTURE="+mode, "SJSKILLS_PROCESS_TREE_READY="+ready, "SJSKILLS_PROCESS_TREE_WRITES="+writes))
			}
			materializer := NewMaterializer(MaterializerConfig{Runner: runner, TempRootFactory: func() (string, error) { return os.MkdirTemp(parent, "stage-") }})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if mode == "timeout" {
				ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
				defer cancel()
			}
			done := make(chan error, 1)
			go func() {
				_, err := materializer.Materialize(ctx, []DesiredSkill{desiredMaterializeSkill("fixture", "example/fixture")})
				done <- err
			}()
			deadline := time.Now().Add(8 * time.Second)
			for {
				if _, err := os.Stat(ready); err == nil {
					break
				}
				select {
				case err := <-done:
					t.Fatalf("premature materializer return %v", err)
				default:
				}
				if time.Now().After(deadline) {
					t.Fatal("no descendant")
				}
				time.Sleep(5 * time.Millisecond)
			}
			if mode == "cancel" {
				cancel()
			}
			select {
			case err := <-done:
				if err == nil || strings.Contains(err.Error(), "staging retained") {
					t.Fatalf("materializer error %v", err)
				}
			case <-time.After(6 * time.Second):
				t.Fatal("materializer hung")
			}
			entries, err := os.ReadDir(parent)
			if err != nil || len(entries) != 0 {
				t.Fatalf("staging cleanup: %v %v", entries, err)
			}
			if _, err := os.Stat(writes); !os.IsNotExist(err) {
				t.Fatalf("staging file still exists: %v", err)
			}
		})
	}
}
func TestMaterializerRetainsStagingWhenTerminationIsUnverifiable(t *testing.T) {
	runner := defaultMaterializeRunner()
	runner.invoke = func(context.Context, string, []string, []string) (ProcessResult, error) {
		return ProcessResult{}, errProcessTreeActive
	}
	materializer, stage := testMaterializer(t, runner, MaterializerLimits{})
	_, err := materializer.Materialize(context.Background(), []DesiredSkill{desiredMaterializeSkill("fixture", "example/fixture")})
	if err != errProcessTreeActive || stage() == "" {
		t.Fatalf("unsafe cleanup: stage=%s error=%v", stage(), err)
	}
	if strings.Contains(fmt.Sprint(err), stage()) {
		t.Fatal("staging path leaked")
	}
}
