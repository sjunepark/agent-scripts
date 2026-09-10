package sjskills

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// Access is a fetching policy, independent of installed source ownership.
// The zero value retains the public policy of older registries and manifests.
type Access string

const (
	AccessPublic              Access = "public"
	AccessGitHubAuthenticated Access = "github-authenticated"
)

func (a Access) Effective() Access {
	if a == "" {
		return AccessPublic
	}
	return a
}

func (a Access) valid() bool {
	return a.Effective() == AccessPublic || a == AccessGitHubAuthenticated
}

func (a *Access) UnmarshalText(data []byte) error {
	value := Access(data)
	if value == "" || !value.valid() {
		return fmt.Errorf("access must be public or github-authenticated")
	}
	*a = value
	return nil
}

func (a *Access) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("access must be public or github-authenticated")
	}
	return a.UnmarshalText([]byte(value))
}

// githubRepository accepts only source forms whose GitHub repository identity
// is unambiguous before invoking Git or a credential helper.
func githubRepository(source string) (string, error) {
	if SkillsCLIPathProblem(source) != "" {
		return "", fmt.Errorf("authenticated source must be a credential-free GitHub.com repository")
	}
	path := source
	if strings.HasPrefix(source, "https://") {
		u, err := url.Parse(source)
		if err != nil || u.Host != "github.com" || u.RawPath != "" || strings.Contains(u.Path, "%") {
			return "", fmt.Errorf("authenticated source must use https://github.com without a port or escaped path")
		}
		path = strings.TrimPrefix(u.Path, "/")
	}
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 || !shorthandPattern.MatchString(path) {
		return "", fmt.Errorf("authenticated source must identify a GitHub owner and repository")
	}
	for _, part := range parts {
		if part == "." || part == ".." {
			return "", fmt.Errorf("authenticated source must not contain traversal")
		}
	}
	repo := strings.TrimSuffix(parts[1], ".git")
	if repo == "" || repo == "." || repo == ".." {
		return "", fmt.Errorf("authenticated source must identify a repository")
	}
	return parts[0] + "/" + repo, nil
}

func validateAccess(path string, access Access, source string, issues *[]Issue) {
	if !access.valid() {
		addIssue(issues, IssueInvalidSource, path+".access", "must be public or github-authenticated")
	} else if access == AccessGitHubAuthenticated {
		if _, err := githubRepository(source); err != nil {
			addIssue(issues, IssueInvalidSource, path+".source", "%s", err)
		}
	}
}
