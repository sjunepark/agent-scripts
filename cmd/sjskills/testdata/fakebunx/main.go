// This executable keeps CLI integration tests offline on every supported OS.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	args := os.Args[1:]
	names := []string{}
	for index, arg := range args {
		if arg != "--skill" {
			continue
		}
		for _, name := range args[index+1:] {
			if strings.HasPrefix(name, "-") {
				break
			}
			names = append(names, name)
		}
		break
	}
	if path := os.Getenv("SJSKILLS_FAKE_LOG"); path != "" {
		labels := names
		if len(labels) == 0 {
			labels = []string{"version"}
		}
		file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			panic(err)
		}
		_, _ = fmt.Fprintln(file, strings.Join(labels, " "))
		_ = file.Close()
	}
	if len(args) == 1 && args[0] == "--version" {
		fmt.Println("bunx 1")
		return
	}
	if len(args) == 2 && args[0] == "skills@1.5.23" && args[1] == "--version" {
		fmt.Println("1.5.23")
		return
	}
	if len(args) < 2 || args[0] != "skills@1.5.23" || args[1] != "add" {
		os.Exit(4)
	}
	if os.Getenv("SJSKILLS_FAKE_BLOCK") == "1" {
		for {
			time.Sleep(time.Second)
		}
	}
	for _, skill := range names {
		if skill == os.Getenv("SJSKILLS_FAKE_FAIL_SKILL") {
			fmt.Fprintln(os.Stderr, "injected unavailable source")
			os.Exit(7)
		}
		target := filepath.Join(os.Getenv("CODEX_HOME"), "skills", skill)
		if err := os.MkdirAll(target, 0o755); err != nil {
			panic(err)
		}
		content := fmt.Sprintf("# %s%s\n", skill, os.Getenv("SJSKILLS_FAKE_CONTENT"))
		local := args[2]
		if local == "fixture/private/skills" && os.Getenv("SJSKILLS_GIT_FIXTURE") != "" {
			local = filepath.Join(os.Getenv("SJSKILLS_GIT_FIXTURE"), "skills")
		}
		if filepath.IsAbs(local) {
			data, err := os.ReadFile(filepath.Join(local, skill, "SKILL.md"))
			if err != nil {
				os.Exit(7)
			}
			content = string(data)
		}
		if err := os.WriteFile(filepath.Join(target, "SKILL.md"), []byte(content), 0o644); err != nil {
			panic(err)
		}
	}
	if len(names) == 0 {
		os.Exit(3)
	}
}
