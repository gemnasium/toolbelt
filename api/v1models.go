package api

import "time"

type Advisory struct {
	ID               int      `json:"id"`
	Title            string   `json:"title"`
	Identifier       string   `json:"identifier"`
	Description      string   `json:"description"`
	Solution         string   `json:"solution"`
	AffectedVersions string   `json:"affected_versions"`
	Package          Package  `json:"package"`
	CuredVersions    string   `json:"cured_versions"`
	Credits          string   `json:"credits"`
	Links            []string `json:"links"`
}

type Alert struct {
	Advisory Advisory  `json:"advisory"`
	OpenAt   time.Time `json:"open_at"`
	Status   string    `json:"status"`
}

type Dependency struct {
	Requirement   string     `json:"requirement"`
	LockedVersion string     `json:"locked"`
	Package       Package    `json:"package"`
	FirstLevel    bool       `json:"first_level"`
	Color         string     `json:"color"`
	Advisories    []Advisory `json:"advisories,omitempty"`
}

type DependencyFile struct {
	Path    string `json:"path"`
	SHA     string `json:"sha,omitempty"`
	Content []byte `json:"content"`
}

type LiveEvalResponse struct {
	Status string `json:"status"`
	Result struct {
		RuntimeStatus     string              `json:"runtime_status"`
		DevelopmentStatus string              `json:"development_status"`
		Dependencies      []Dependency `json:"dependencies"`
	} `json:"result"`
}

type Package struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	Type string `json:"type"`
}

type Project struct {
	Name              string `json:"name,omitempty"`
	Slug              string `json:"slug,omitempty"`
	Description       string `json:"description,omitempty"`
	Origin            string `json:"origin,omitempty"`
	Private           bool   `json:"private,omitempty"`
	Color             string `json:"color,omitempty"`
	Monitored         bool   `json:"monitored,omitempty"`
	UnmonitoredReason string `json:"unmonitored_reason,omitempty"`
	CommitSHA         string `json:"commit_sha"`
}

type RequirementUpdate struct {
	File  DependencyFile `json:"file"`
	Patch string                `json:"patch"`
}

type UpdateSet struct {
	ID                 int                            `json:"id"`
	RequirementUpdates map[string][]RequirementUpdate `json:"requirement_updates"`
	VersionUpdates     map[string][]VersionUpdate     `json:"version_updates"`
}

type UpdateSetResult struct {
	UpdateSetID     int                     `json:"-"`
	ProjectSlug     string                  `json:"-"`
	State           string                  `json:"state"`
	DependencyFiles []DependencyFile `json:"dependency_files"`
}

type VersionUpdate struct {
	Package       Package
	OldVersion    string `json:"old_version"`
	TargetVersion string `json:"target_version"`
}