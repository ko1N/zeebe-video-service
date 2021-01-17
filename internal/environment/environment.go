package environment

// ExecutionResult - represents a execution
type ExecutionResult struct {
	ExitCode int
	StdOut   string
	StdErr   string
}

type EnvironmentReader interface {
	Read(p []byte) (n int, err error)
	Close() error
}

type EnvironmentWriter interface {
	Write(p []byte) (n int, err error)
	Close() error
}

// Environment -
type Environment interface {
	Name() string
	Execute(cmd []string, stdout func(string), stderr func(string)) (*ExecutionResult, error)

	FullPath(relpath string) string

	FileReader(path string) (EnvironmentReader, error)
	FileWriter(path string) (EnvironmentWriter, error)

	Close()
}
