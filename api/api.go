package api


// The API instance
var APIImpl API

// API abstract type
type API interface {
	Login(email, password string) error
	Endpoint() string
	Host() string
	Key() string
	SetKey(token string)
	AutoUpdateStepsBest(projectSlug string, revision string) (dfiles []DependencyFile, err error)
	AutoUpdateStepsNext(projectSlug string, revision string) (updateSet *UpdateSet, err error)
	AutoUpdateStepsPush(revision string, rs *UpdateSetResult) (err error)
	DependencyAlertsGet(p *Project) (alerts []Alert, err error)
	DependencyFilesPush(projectSlug string, dfiles []*DependencyFile) (jsonResp map[string][]DependencyFile, err error)
	LiveEvalStart(requestDeps map[string][]*DependencyFile) (jsonResp map[string]interface{}, err error)
	LiveEvalGetResponse(jobId interface{}) (response LiveEvalResponse, body []byte, err error)
	ProjectList(privateOnly bool) (owner2Project map[string][]Project, err error)
	ProjectUpdate(p *Project, update map[string]interface{}) (err error)
	ProjectCreate(p *Project) (jsonResp map[string]interface{}, err error)
	ProjectSync(p *Project) (err error)
	ProjectFetch(p *Project) (err error)
	ProjectGetDependencies(p *Project) (deps []Dependency, err error)
	ProjectGetDependencyFiles(p *Project) (dfiles []DependencyFile, err error)
}
