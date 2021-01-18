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
func (e *NativeEnvironment) Execute(cmd []string, stdout func(string), stderr func(string)) (*ExecutionResult, error) {
	//fmt.Printf("exec: `/bin/sh -c \"%s\"`\n", strings.Join(cmd, " "))

	exc := exec.Command("/bin/sh", []string{"-c", strings.Join(cmd, " ")}...)
	exc.Dir = e.PWD

	// create stdout/stderr pipes
	outpr, err := exc.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer outpr.Close()
	outsig := make(chan struct{})
	var outBuf bytes.Buffer

	errpr, err := exc.StderrPipe()
	if err != nil {
		return nil, err
	}
	defer errpr.Close()
	errsig := make(chan struct{})
	var errBuf bytes.Buffer

	exc.Stdin = nil

	// start process
	err = exc.Start()
	if err != nil {
		return nil, err
	}

	// track stdout
	go func() {
		reader := bufio.NewScanner(outpr)
		for reader.Scan() {
			if stdout != nil {
				stdout(reader.Text())
			}
			outBuf.Write([]byte(reader.Text()))
			outBuf.Write([]byte("\n"))
		}
		outsig <- struct{}{}
	}()

	// track stderr
	go func() {
		reader := bufio.NewScanner(errpr)
		for reader.Scan() {
			if stderr != nil {
				stderr(reader.Text())
			}
			errBuf.Write([]byte(reader.Text()))
			errBuf.Write([]byte("\n"))
		}
		errsig <- struct{}{}
	}()

	// wait for exc to finish
	err = exc.Wait()

	// synchropnize with stdout/stderr
	<-outsig
	<-errsig

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
