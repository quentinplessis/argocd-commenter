package github

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/google/go-github/v39/github"
)

type CheckRunID struct {
	Repository Repository
	ID         int64
}

var patternCheckRunURL = regexp.MustCompile(`^https://api\.github\.com/repos/(.+?)/(.+?)/check-runs/(\d+)$`)

// ParseCheckRunURL parses the URL.
// For example, https://api.github.com/repos/int128/sandbox/check-runs/5348989392
func ParseCheckRunURL(s string) *CheckRunID {
	m := patternCheckRunURL.FindStringSubmatch(s)
	if len(m) != 4 {
		return nil
	}
	id, err := strconv.ParseInt(m[3], 10, 64)
	if err != nil {
		return nil
	}
	return &CheckRunID{
		Repository: Repository{Owner: m[1], Name: m[2]},
		ID:         id,
	}
}

type CheckRun struct {
	Status     string
	Conclusion string
	Title      string
	Summary    string
}

func (c *client) UpdateCheckRun(ctx context.Context, id CheckRunID, cr CheckRun) error {
	o := github.UpdateCheckRunOptions{
		Status: github.String(cr.Status),
		Output: &github.CheckRunOutput{
			Title:   github.String(cr.Title),
			Summary: github.String(cr.Summary),
		},
	}
	if cr.Conclusion != "" {
		o.Conclusion = github.String(cr.Conclusion)
	}
	_, _, err := c.rest.Checks.UpdateCheckRun(ctx, id.Repository.Owner, id.Repository.Name, id.ID, o)
	if err != nil {
		return fmt.Errorf("GitHub API error: %w", err)
	}
	return nil
}
