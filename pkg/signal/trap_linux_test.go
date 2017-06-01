// +build linux
package signal

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrap(t *testing.T) {
	sigmap := map[string]os.Signal{
		"TERM": syscall.SIGTERM,
		"QUIT": syscall.SIGQUIT,
		"INT":  os.Interrupt,
	}
	if os.Getenv("TEST_TRAP") == "1" {
		defer time.Sleep(10 * time.Millisecond)
		Trap(func() {
			time.Sleep(10 * time.Millisecond)
			os.Exit(99)
		})
		go func() {
			p, err := os.FindProcess(os.Getpid())
			require.NoError(t, err)
			switch s := os.Getenv("SIGNAL_TYPE"); s {
			case "TERM":
				for {
					p.Signal(sigmap[s])
				}
			case "QUIT":
				p.Signal(sigmap[s])
			case "INT":
				p.Signal(sigmap[s])
			}
		}()
		select {}
	}
	for k, v := range sigmap {
		cmd := exec.Command(os.Args[0], "-test.run=TestTrap")
		cmd.Env = append(os.Environ(), "TEST_TRAP=1", fmt.Sprintf("SIGNAL_TYPE=%s", k))
		err := cmd.Start()
		require.NoError(t, err)
		err = cmd.Wait()
		if e, ok := err.(*exec.ExitError); ok {
			code := e.Sys().(syscall.WaitStatus).ExitStatus()
			switch k {
			case "TERM", "QUIT":
				assert.Equal(t, code, (128 + int(v.(syscall.Signal))))
			case "INT":
				assert.Equal(t, code, 99)
			}
			continue
		}
		t.Fatal("process didn't end with any error")
	}
}
func TestDumpStacks(t *testing.T) {
	directory, errorDir := ioutil.TempDir("", "test")
	assert.NoError(t, errorDir)
	defer os.RemoveAll(directory)
	_, error := DumpStacks(directory)
	assert.NoError(t, error)
	path := filepath.Join(directory, fmt.Sprintf(stacksLogNameTemplate, strings.Replace(time.Now().Format(time.RFC3339), ":", "", -1)))
	readFile, _ := ioutil.ReadFile(path)
	fileData := string(readFile)
	assert.Contains(t, fileData, "goroutine")
}
func TestDumpStacksWithEmptyInput(t *testing.T) {
	path, errorPath := DumpStacks("")
	assert.NoError(t, errorPath)
	assert.Equal(t, os.Stderr.Name(), path)
}
