// This executable keeps CLI integration tests offline on every supported OS.
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	args := os.Args[1:]
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
	for index := 2; index+1 < len(args); index++ {
		if args[index] != "--skill" {
			continue
		}
		skill := args[index+1]
		target := filepath.Join(os.Getenv("CODEX_HOME"), "skills", skill)
		if err := os.MkdirAll(target, 0o755); err != nil {
			panic(err)
		}
		content := fmt.Sprintf("# %s%s\n", skill, os.Getenv("SJSKILLS_FAKE_CONTENT"))
		if err := os.WriteFile(filepath.Join(target, "SKILL.md"), []byte(content), 0o644); err != nil {
			panic(err)
		}
		return
	}
	os.Exit(3)
}
