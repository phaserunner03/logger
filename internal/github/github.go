package githubconnector

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"
)

type PRConfig struct {
	RepoOwner     string
	RepoName      string
	BaseBranch    string
	NewBranch     string
	GithubToken   string
	LocalRepoPath string
	FixFilePath   string
	CommitMessage string
	PRTitle       string
	PRBody        string
}

func CreatePR(cfg PRConfig) error {
	ctx := context.Background()

	// Step 1: Checkout new branch
	if err := runGit(cfg.LocalRepoPath, "checkout", "-b", cfg.NewBranch); err != nil {
		return fmt.Errorf("failed to create branch: %v", err)
	}

	// Step 2: Add changes
	if err := runGit(cfg.LocalRepoPath, "add", cfg.FixFilePath); err != nil {
		return fmt.Errorf("git add failed: %v", err)
	}

	// Step 3: Commit changes
	if err := runGit(cfg.LocalRepoPath, "commit", "-m", cfg.CommitMessage,"--no-verify"); err != nil {
		return fmt.Errorf("git commit failed: %v", err)
	}

	// Step 4: Push to remote
	if err := runGit(cfg.LocalRepoPath, "push", "--set-upstream", "origin", cfg.NewBranch,"--no-verify"); err != nil {
		return fmt.Errorf("git push failed: %v", err)
	}

	// Step 5: Create GitHub PR
	return createPullRequest(ctx, cfg)
}

func createPullRequest(ctx context.Context, cfg PRConfig) error {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: cfg.GithubToken})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	newPR := &github.NewPullRequest{
		Title:               github.String(cfg.PRTitle),
		Head:                github.String(cfg.NewBranch),
		Base:                github.String(cfg.BaseBranch),
		Body:                github.String(cfg.PRBody),
		MaintainerCanModify: github.Bool(true),
	}

	pr, _, err := client.PullRequests.Create(ctx, cfg.RepoOwner, cfg.RepoName, newPR)
	if err != nil {
		return fmt.Errorf("failed to create GitHub PR: %v", err)
	}

	fmt.Printf("🚀 PR created: %s\n", pr.GetHTMLURL())
	return nil
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
