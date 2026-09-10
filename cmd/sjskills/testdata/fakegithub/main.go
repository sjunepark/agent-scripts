// Controlled native Git/gh boundary for authentication acceptance tests.
// Unknown commands fail; this fixture never forwards network operations.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const token = "sjskills-sentinel-credential-never-log"

func fail() { os.Exit(7) }
func main() {
	name := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
	if os.Getenv("SJSKILLS_FAKE_GITHUB_LOG") != "" {
		f, e := os.OpenFile(os.Getenv("SJSKILLS_FAKE_GITHUB_LOG"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if e != nil {
			fail()
		}
		fmt.Fprintln(f, name)
		f.Close()
	}
	switch name {
	case "gh":
		if strings.Join(os.Args[1:], " ") != "auth git-credential get" {
			fail()
		}
		if os.Getenv("SJSKILLS_EXPECT_TOKEN") == "1" {
			if os.Getenv("GH_TOKEN") != token || os.Getenv("GH_CONFIG_DIR") != filepath.Join(os.Getenv("HOME"), ".gh-auth") {
				fail()
			}
		} else if os.Getenv("GH_CONFIG_DIR") != os.Getenv("SJSKILLS_EXPECT_GH_CONFIG") {
			fail()
		}
		if os.Getenv("HOME") == os.Getenv("SJSKILLS_ORIGINAL_HOME") {
			fail()
		}
		if os.Getenv("GH_DEBUG") != "" || os.Getenv("GIT_TRACE") != "" {
			fail()
		}
		if os.Getenv("SJSKILLS_FAKE_AUTH_BLOCK") == "1" {
			deadline := time.Now().Add(8 * time.Second)
			for time.Now().Before(deadline) {
				_ = os.WriteFile(os.Getenv("SJSKILLS_FAKE_AUTH_HEARTBEAT"), []byte(fmt.Sprint(time.Now().UnixNano())), 0600)
				time.Sleep(20 * time.Millisecond)
			}
			fail()
		}
		fmt.Fprintln(os.Stderr, token) // Must never reach any diagnostic channel.
		if os.Getenv("SJSKILLS_FAKE_AUTH_FAIL") == "1" {
			fail()
		}
		if os.Getenv("SJSKILLS_FAKE_AUTH_OVERSIZE") == "1" {
			fmt.Print(strings.Repeat("x", 128*1024))
			return
		}
		fmt.Printf("protocol=https\nhost=github.com\nusername=fixture\npassword=%s\n\n", token)
	case "git":
		args := os.Args[1:]
		index := -1
		for i, a := range args {
			if a == "clone" {
				index = i
				break
			}
		}
		if index < 0 || len(args) < 2 || os.Getenv("GIT_ALLOW_PROTOCOL") != "https" || os.Getenv("GIT_CONFIG_COUNT") != "" || os.Getenv("GIT_TRACE") != "" {
			fail()
		}
		joined := strings.Join(args[:index], " ")
		for _, want := range []string{"credential.helper=", "credential.useHttpPath=true", "http.followRedirects=false", "submodule.recurse=false", "core.hooksPath="} {
			if !strings.Contains(joined, want) {
				fail()
			}
		}
		remote := args[len(args)-2]
		if remote != "https://github.com/fixture/private.git" {
			fail()
		}
		realGit := os.Getenv("SJSKILLS_REAL_GIT")
		if !filepath.IsAbs(realGit) {
			fail()
		}
		runCredential := func(host, path string) ([]byte, error) {
			ca := append(append([]string{}, args[:index]...), "credential", "fill")
			cmd := exec.Command(realGit, ca...)
			cmd.Env = os.Environ()
			cmd.Stdin = strings.NewReader("protocol=https\nhost=" + host + "\npath=" + path + "\n\n")
			return cmd.Output()
		}
		for _, request := range [][2]string{{"evil.invalid", "fixture/private.git"}, {"github.com", "fixture/other.git"}} {
			out, err := runCredential(request[0], request[1])
			if err == nil || strings.Contains(string(out), token) {
				fail()
			}
		}
		out, err := runCredential("github.com", "fixture/private.git")
		if err != nil || !strings.Contains(string(out), "password="+token) {
			fmt.Fprintln(os.Stderr, token)
			fail()
		}
		cloneArgs := append([]string{}, args[:index]...)
		cloneArgs = append(cloneArgs, "clone", "--no-recurse-submodules", "--", os.Getenv("SJSKILLS_GIT_FIXTURE"), args[len(args)-1])
		cmd := exec.Command(realGit, cloneArgs...)
		env := []string{}
		for _, e := range os.Environ() {
			if !strings.HasPrefix(e, "GIT_ALLOW_PROTOCOL=") {
				env = append(env, e)
			}
		}
		cmd.Env = append(env, "GIT_ALLOW_PROTOCOL=file")
		if err := cmd.Run(); err != nil {
			fail()
		}
	default:
		fail()
	}
}
