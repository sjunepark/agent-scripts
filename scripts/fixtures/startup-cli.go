// Native process fixture for the startup adapter; never used by the plugin.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

func main() {
	if len(os.Args) == 4 && os.Args[1] == "--orphan" {
		child := exec.Command(os.Args[2], "-e", "setInterval(()=>{},1000)")
		child.Stdout, child.Stderr = os.Stdout, os.Stderr
		if err := child.Start(); err != nil {
			panic(err)
		}
		if err := os.WriteFile(os.Args[3], []byte(strconv.Itoa(child.Process.Pid)), 0600); err != nil {
			panic(err)
		}
		child.Process.Release()
		return
	}
	if file := os.Getenv("HOOK_FIXTURE_LOG"); file != "" {
		cwd, _ := os.Getwd()
		entry, _ := json.Marshal(map[string]any{"args": os.Args[1:], "cwd": cwd})
		f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			panic(err)
		}
		fmt.Fprintln(f, string(entry))
		f.Close()
	}
	if len(os.Args) == 3 && os.Args[1] == "--json" && os.Args[2] == "status" {
		fmt.Println(os.Getenv("HOOK_FIXTURE_STATUS"))
		return
	}
	if len(os.Args) == 7 && os.Args[1] == "plugin" && os.Args[2] == "list" {
		fmt.Println(os.Getenv("HOOK_FIXTURE_PLUGINS"))
		return
	}
	os.Exit(17)
}
