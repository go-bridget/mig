package gen

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
