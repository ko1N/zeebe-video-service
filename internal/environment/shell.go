package environment

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// implements a native environment ("bare metal")

// NativeEnvironment -
type NativeEnvironment struct {
	PWD string
}

// CreateNativeEnvironment -
func CreateNativeEnvironment() (*NativeEnvironment, error) {
	tempFolder := path.Join(".", "temp", uuid.New().String())
	err := os.MkdirAll(tempFolder, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &NativeEnvironment{
		PWD: tempFolder,
	}, nil
}

// Name - returns the name of the native environment
func (e *NativeEnvironment) Name() string {
	return "native"
}

// Execute -
func (e *NativeEnvironment) Execute(cmd string, args []string, stdout func(string), stderr func(string)) (*ExecutionResult, error) {
	//fmt.Printf("exec: `%s %s`\n", cmd, strings.Join(args, " "))

	exc := exec.Command(cmd, args...)
	exc.Dir = e.PWD

	// create stdout/stderr pipes
	stdoutpipe, err := exc.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderrpipe, err := exc.StderrPipe()
	if err != nil {
		return nil, err
	}

	exc.Stdin = nil

	// start process
	err = exc.Start()
	if err != nil {
		return nil, err
	}

	// track stderr
	stderrsig := make(chan struct{})
	var errBuf bytes.Buffer
	go func() {
		reader := bufio.NewReader(stderrpipe)
		for {
			text, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			if stderr != nil {
				stderr(strings.TrimSuffix(text, "\n"))
			}
			errBuf.Write([]byte(text))
		}
		stderrsig <- struct{}{}
	}()

	// track stdout
	stdoutsig := make(chan struct{})
	var outBuf bytes.Buffer
	go func() {
		reader := bufio.NewReader(stdoutpipe)
		for {
			text, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			if stdout != nil {
				stdout(strings.TrimSuffix(text, "\n"))
			}
			outBuf.Write([]byte(text))
		}
		stdoutsig <- struct{}{}
	}()

	// wait for both pipes to be closed before calling wait
	<-stderrsig
	<-stdoutsig

	// wait for exc to finish
	err = exc.Wait()
	if err != nil {
		return nil, err
	}

	return &ExecutionResult{
		ExitCode: exc.ProcessState.ExitCode(),
		StdOut:   outBuf.String(),
		StdErr:   errBuf.String(),
	}, nil
}

func (self *NativeEnvironment) FullPath(relpath string) string {
	return path.Join(self.PWD, relpath)
}

// FileReader - returns a reader for a file in the environment
func (self *NativeEnvironment) FileReader(filename string) (EnvironmentReader, error) {
	realpath := path.Join(self.PWD, filename)
	subpath, _ := isSubPath(self.PWD, realpath)
	if !subpath {
		return nil, fmt.Errorf("can't open file reader outside of sandbox")
	}

	file, err := os.Open(realpath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// FileWriter - returns a writer for a file in the environment
func (self *NativeEnvironment) FileWriter(filename string) (EnvironmentWriter, error) {
	realpath := path.Join(self.PWD, filename)
	subpath, _ := isSubPath(self.PWD, realpath)
	if !subpath {
		return nil, fmt.Errorf("can't open file reader outside of sandbox")
	}

	file, err := os.Create(realpath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Close - shuts down the environment and removes the temp folder
func (e *NativeEnvironment) Close() {
	if e.PWD != "" {
		os.RemoveAll(e.PWD)
	}
}

func isSubPath(parent, sub string) (bool, error) {
	up := ".." + string(os.PathSeparator)

	// path-comparisons using filepath.Abs don't work reliably according to docs (no unique representation).
	rel, err := filepath.Rel(parent, sub)
	if err != nil {
		return false, err
	}
	if !strings.HasPrefix(rel, up) && rel != ".." {
		return true, nil
	}
	return false, nil
}
