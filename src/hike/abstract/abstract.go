package abstract

type BuildFrame interface {
	PrintErrorFrame(level uint) error
}

type BuildError interface {
	PrintBuildError(level uint) error
	AddErrorFrame(frame *BuildFrame)
}

type ArtifactKey struct {
	Project string
	Artifact string
}

func (key *ArtifactKey) Unified() string {
	return key.Project + "::" + key.Artifact
}

type Artifact interface {
	ArtifactKey() ArtifactKey
	DisplayName() string
	PathNames(sink []string) []string
	Flatten() BuildError
	Require(plan *Plan) BuildError
}

type Step interface {
	Perform() BuildError
	SimpleDescr() string
}

type Transform interface {
	Plan(destinations []*Artifact, plan *Plan) BuildError
}

type Goal interface {
	Attain(plan *Plan) BuildError
}

type Config struct {
	ProjectName string
	InducedProjectName string
	TopDir string
}

func (config *Config) EffectiveProjectName() string {
	switch {
		case len(config.InducedProjectName) > 0:
			return config.InducedProjectName
		case len(config.ProjectName) > 0:
			return config.ProjectName
		default:
			return "this"
	}
}

type SpecState struct {
	Config *Config
	PipelineTip *Artifact
}

type Plan struct {
	steps []Step
	knownUpToDate map[*Artifact]bool
}

func (plan *Plan) AddStep(step Step) {
	plan.steps = append(plan.steps, step)
}

func (plan *Plan) StepCount() int {
	return len(plan.steps)
}

func (plan *Plan) BroughtUpToDate(artifact *Artifact) {
	plan.knownUpToDate[artifact] = true
}

func (plan *Plan) AlreadyUpToDate(artifact *Artifact) bool {
	return plan.knownUpToDate[artifact]
}
