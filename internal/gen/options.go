package gen

// Options contains code generation options.
type Options struct {
	Language string
	Schema   string
	Output   string

	Go struct {
		FillJSON bool
		SkipJSON bool
	}

	PHP struct {
		Namespace string
	}
}

// NewOptions creates a new Options instance with default values.
func NewOptions() *Options {
	return &Options{
		Language: "go",
		Output:   "types",
	}
}
