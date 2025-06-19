package githubconnector

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

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
	FixContent    string
	CommitMessage string
	PRTitle       string
	PRBody        string
}

func CreatePR(cfg PRConfig) error {
	ctx := context.Background()

	if err := runGit(cfg.LocalRepoPath, "checkout", "-b", cfg.NewBranch); err != nil {
		return fmt.Errorf("failed to create branch:%v", err)
	}
	fixes := map[string]string{
		"console.log(person.age);":         `console.log("Age:", person?.age || "N/A");`,
		"inputRef.current.value = 'Test';": `if (inputRef.current) inputRef.current.value = 'Test';`,
		"setCount(count++);":               `setCount(prev => prev + 1);`,
	}
	
	err := writeIntoFile(cfg.FixFilePath,fixes)
	if err != nil {
		return fmt.Errorf("failed to fix the file %v", err)
	}

	if err := runGit(cfg.LocalRepoPath, "add", "."); err != nil {
		return err
	}

	if err := runGit(cfg.LocalRepoPath, "commit", "-m", cfg.CommitMessage); err != nil {
		return err
	}

	if err := runGit(cfg.LocalRepoPath, "push", "--set-upstream", "origin", cfg.NewBranch); err != nil {
		return err
	}

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
	pr,_,err:= client.PullRequests.Create(ctx,cfg.RepoOwner,cfg.RepoName,newPR)

	if err!= nil{
		return fmt.Errorf("failed to create PR %v",err)
	}
	fmt.Printf("Pull request created: %s \n", pr.GetHTMLURL())

	return nil
}

func writeIntoFile(filePath string, fixes map[string]string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	
	contentBytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}
	content := string(contentBytes)

	for oldLine, newLine := range fixes {
		content = strings.ReplaceAll(content, oldLine, newLine)
	}

	
	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to write updated file: %v", err)
	}
	return nil
	
}

func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
