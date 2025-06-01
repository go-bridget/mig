package model

type Options struct {
	Language string
	Schema   string
	Output   string

	Go struct {
		FillJSON bool
	}

	PHP struct {
		Namespace string
	}
}
