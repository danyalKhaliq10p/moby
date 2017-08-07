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
	for k, v := range sigmap {
		tmpfile, err := ioutil.TempFile("", "main")
		defer os.Remove(tmpfile.Name())
		require.NoError(t, err)
		wd, _ := os.Getwd()
		testHelperCode := wd + "/testfiles/main.go"
		cmd := exec.Command("go", "build", "-o", tmpfile.Name(), testHelperCode)
		err = cmd.Run()
		require.NoError(t, err)
		cmd = exec.Command(tmpfile.Name())
		cmd.Env = append(os.Environ(), fmt.Sprintf("SIGNAL_TYPE=%s", k))
		err = cmd.Start()
		require.NoError(t, err)
		err = cmd.Wait()
		if e, ok := err.(*exec.ExitError); ok {
			code := e.Sys().(syscall.WaitStatus).ExitStatus()
			switch k {
			case "TERM", "QUIT":
				assert.Equal(t, (128 + int(v.(syscall.Signal))), code)
			case "INT":
				assert.Equal(t, 99, code)
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
