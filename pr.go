package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
)

type Author struct {
	Login string `json:"login"`
}

type PullRequest struct {
	Number       int    `json:"number"`
	Title        string `json:"title"`
	HeadRefName  string `json:"headRefName"`
	Author       Author `json:"author"`
	CreatedAt    string `json:"createdAt"`
	IsDraft      bool   `json:"isDraft"`
	Additions    int    `json:"additions"`
	Deletions    int    `json:"deletions"`
	ChangedFiles int    `json:"changedFiles"`
	AuthorName   string `json:"-"`
}

func fetchPRs(searchOptions string) ([]PullRequest, error) {
	args := []string{"pr", "list"}
	if searchOptions != "" {
		args = append(args, strings.Fields(searchOptions)...)
	}
	args = append(args, "--json", "number,title,headRefName,author,createdAt,isDraft,additions,deletions,changedFiles")

	cmd := exec.Command("gh", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("gh pr list: %w: %s", err, stderr.String())
	}

	var prs []PullRequest
	if err := json.Unmarshal(stdout.Bytes(), &prs); err != nil {
		return nil, fmt.Errorf("decode JSON: %w", err)
	}
	return prs, nil
}

func defaultBranches() ([]PullRequest, error) {
	cmd := exec.Command("git", "branch", "-r")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git branch -r: %w", err)
	}

	wanted := map[string]bool{
		"main": true, "master": true, "develop": true, "staging": true,
	}

	var branches []string
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "origin/") {
			continue
		}
		name := strings.TrimPrefix(line, "origin/")
		if wanted[name] {
			branches = append(branches, name)
		}
	}
	sort.Strings(branches)

	var prs []PullRequest
	for _, branch := range branches {
		epoch, err := branchFirstCommitTime(branch)
		if err != nil {
			continue
		}
		prs = append(prs, PullRequest{
			Number:      0,
			Title:       branch,
			HeadRefName: branch,
			Author:      Author{Login: "system"},
			CreatedAt:   epoch.UTC().Format("2006-01-02T15:04:05Z"),
			IsDraft:     false,
			Additions:   0,
			Deletions:   0,
		})
	}
	return prs, nil
}

func branchFirstCommitTime(branch string) (time.Time, error) {
	cmd := exec.Command("git", "log", "origin/"+branch, "--reverse", "--pretty=format:%ct")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return time.Time{}, err
	}
	lines := strings.SplitN(stdout.String(), "\n", 2)
	if len(lines) == 0 || lines[0] == "" {
		return time.Time{}, fmt.Errorf("no commits for %s", branch)
	}
	var epoch int64
	if _, err := fmt.Sscanf(lines[0], "%d", &epoch); err != nil {
		return time.Time{}, err
	}
	return time.Unix(epoch, 0), nil
}
