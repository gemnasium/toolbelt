package api

import (
	"errors"
	"regexp"
	"log"
	"encoding/base64"
)

type V2ToV1 struct {
	*APIv2
}

func (a *V2ToV1) Login(email, password string) error {
	return a.APIv2.Login(email, password)
}
func (a *V2ToV1) Endpoint() string {
	return a.APIv2.Endpoint()
}
func (a *V2ToV1) Host() string {
	return a.APIv2.Host()
}
func (a *V2ToV1) Key() string {
	return a.APIv2.Key()
}
func (a *V2ToV1) SetKey(key string) {
	a.APIv2.SetKey(key)
}
func (a *V2ToV1) AutoUpdateStepsBest(projectSlug string, revision string) (dfiles []DependencyFile, err error) {
	return a.APIv2.AutoUpdateStepsBest(projectSlug, revision)
}
func (a *V2ToV1) AutoUpdateStepsNext(projectSlug string, revision string) (updateSet *UpdateSet, err error) {
	return a.APIv2.AutoUpdateStepsNext(projectSlug, revision)
}
func (a *V2ToV1) AutoUpdateStepsPush(revision string, rs *UpdateSetResult) (err error) {
	return a.APIv2.AutoUpdateStepsPush(revision, rs)
}
func (a *V2ToV1) DependencyAlertsGet(p *Project) (alerts []Alert, err error) {
	//Not implemented
	return alerts, err
}
func (a *V2ToV1) DependencyFilesPush(projectSlug string, dfiles []*DependencyFile) (jsonResp map[string][]DependencyFile, err error) {
	// Convert files to v2
	v2dfiles := []*V2DependencyFile{}
	for _, dfile := range dfiles {
		v2dfile := V2DependencyFile{}
		V1DependencyFileToV2(dfile, &v2dfile)
		// Base64 encode content
		v2dfile.Content = base64.StdEncoding.EncodeToString([]byte(v2dfile.Content))
		v2dfiles = append(v2dfiles, &v2dfile)
	}
	_, err = a.APIv2.DependencyFilesPush(projectSlug, v2dfiles)
	return jsonResp, err
}
func (a *V2ToV1) LiveEvalStart(requestDeps map[string][]*DependencyFile) (jsonResp map[string]interface{}, err error) {
	return a.APIv2.LiveEvalStart(requestDeps)
}
func (a *V2ToV1) LiveEvalGetResponse(jobId interface{}) (response LiveEvalResponse, body []byte, err error) {
	return a.APIv2.LiveEvalGetResponse(jobId)
}
func (a *V2ToV1) ProjectList(privateOnly bool) (owner2Project map[string][]Project, err error) {
	projects, err := a.APIv2.ProjectList(privateOnly)
	if err != nil {
		return owner2Project, err
	}
	// Get current user, its projects get a special treatment bellow
	currentUser, err := a.User()
	if err != nil {
		return owner2Project, err
	}
	// Transforms into expected return type
	owner2Project = map[string][]Project{}
	for _, v2p := range projects {
		owner := v2p.Team.Owner.Name
		if owner == currentUser.Name {
			// Replace user's own name by "owned"
			owner = "owned"
		}
		p := Project{}
		V2ProjectToV1(&v2p, &p)
		ownersProjects, ok := owner2Project[owner]
		if ok {
			owner2Project[owner] = append(ownersProjects, p)
		} else {
			owner2Project[owner] = []Project{p}
		}
	}
	return owner2Project, err
}
func (a *V2ToV1) ProjectUpdate(p *Project, update map[string]interface{}) (err error) {
	// Convert project into a v2 project
	v2p := V2Project{}
	V1ProjectToV2(p, &v2p)
	// Adapts update map
	if update["desc"] != "" {
		update["description"] = update["desc"]
		delete(update, "desc")
	}
	// Call update on v2 project
	err = a.APIv2.ProjectUpdate(&v2p, update)
	return err
}
func (a *V2ToV1) ProjectCreate(p *Project) (jsonResp map[string]interface{}, err error) {
	// Convert project into a v2 project
	v2p := V2Project{}
	V1ProjectToV2(p, &v2p)
	//Fill basename from Name
	v2p.Basename = makeBasename(v2p.Name)
	// v2 needs to specify a team to create the project into. Get the current user's team
	teams, err := a.APIv2.UserTeams()
	if len(teams) == 0 {
		return jsonResp, errors.New("Current user has no team !")
	}
	//Pick the first team arbitrarily
	//TODO allow choosing team with a command line parameter
	team := teams[0]
	return a.APIv2.ProjectCreate(team.Slug, &v2p)
}
func (a *V2ToV1) ProjectSync(p *Project) (err error) {
	// Convert project into a v2 project
	v2p := V2Project{}
	V1ProjectToV2(p, &v2p)
	return a.APIv2.ProjectSync(&v2p)
}
func (a *V2ToV1) ProjectFetch(p *Project) (err error) {
	// Fetch a v2 project with the same slug
	v2p := V2Project{Slug: p.Slug}
	if err = a.APIv2.ProjectFetch(&v2p); err != nil {
		return err
	}
	// Convert the v2 project to v1
	V2ProjectToV1(&v2p, p)
	return nil
}
func (a *V2ToV1) ProjectGetDependencies(p *Project) (deps []Dependency, err error) {
	return a.APIv2.ProjectGetDependencies(p)
}
func (a *V2ToV1) ProjectGetDependencyFiles(p *Project) (dfiles []DependencyFile, err error) {
	// Convert project into a v2 project
	v2p := V2Project{}
	V1ProjectToV2(p, &v2p)
	v2dfiles, err := a.APIv2.ProjectGetDependencyFiles(&v2p)
	if err != nil {
		return dfiles, err
	}
	for _, v2dfile := range v2dfiles {
		dfile := DependencyFile{}
		V2DependencyFileToV1(&v2dfile, &dfile)
		dfiles = append(dfiles, dfile)
	}
	return dfiles, nil

}

// Utility functions

func V2ProjectToV1(v2p *V2Project, v1p *Project) {
	v1p.Name = v2p.Name
	v1p.Slug = v2p.Slug
	v1p.Description = v2p.Description
	v1p.Origin = v2p.Origin
	v1p.Private = v2p.Private
	v1p.Color = v2p.Color
	v1p.Monitored = true
	v1p.CommitSHA = v2p.LatestCommit.CommitSHA
}
func V1ProjectToV2(v1p *Project, v2p *V2Project) {
	v2p.Name = v1p.Name
	v2p.Slug = v1p.Slug
	v2p.Description = v1p.Description
	v2p.Origin = v1p.Origin
	v2p.Private = v1p.Private
	v2p.Color = v1p.Color
	v2p.LatestCommit = V2Commit{CommitSHA: v1p.CommitSHA}
}

func V2DependencyFileToV1(v2df *V2DependencyFile, v1df *DependencyFile) {
	v1df.Path = v2df.Path
	v1df.SHA = v2df.SHA
	v1df.Content = []byte(v2df.Content)
}

func V1DependencyFileToV2(v1df *DependencyFile, v2df *V2DependencyFile) {
	v2df.Path = v1df.Path
	v2df.SHA = v1df.SHA
	v2df.Content = string(v1df.Content)
}

func makeBasename(name string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9_-]+")
	if err != nil {
		log.Fatal(err)
	}
	return reg.ReplaceAllString(name, "")
}
