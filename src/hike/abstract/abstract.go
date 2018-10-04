package abstract

type BuildFrame interface {
	PrintErrorFrame(level uint) error
}

type BuildError interface {
	PrintBuildError(level uint) error
	AddErrorFrame(frame *BuildFrame)
}

type Artifact interface {
	DisplayName() string
	PathNames() []string
}

type Step interface {
	Perform()
	SimpleDescr() string
}

type Transform interface {
	//TODO
}
