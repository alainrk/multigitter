package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/go-github/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

// repositories := map[string]string{
// 	"multigitter-test-1": "alainrk/multigitter-test-1",
// 	"multigitter-test-2": "alainrk/multigitter-test-2",
// 	"multigitter-test-3": "alainrk/multigitter-test-3",
// }

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	username := os.Getenv("GITHUB_USERNAME")
	token := os.Getenv("GITHUB_TOKEN")

	fmt.Println("Github username used:", username)

	// Create a GitHub client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	name := "alainrk/multigitter-test-1"

	// Search for the repository we want to work with
	l, _, err := client.Search.Repositories(ctx, name, &github.SearchOptions{})
	if err != nil {
		log.Fatalf("Error listing repos: %s", err)
	}
	if len(l.Repositories) != 1 {
		log.Fatalf("Unexpected number of repositories for repository '%s': %d", name, len(l.Repositories))
	}
	currentRepo := l.Repositories[0]

	// Get or create the currentBranch we want to work with
	// TODO: Maybe I just want to first check if it already exists and create another one if so
	currentBranch, err := getOrCreateBranch(client, username, currentRepo.GetName(), "test-branch-3", "main", ctx)
	if err != nil {
		log.Fatalf("Error getting branch: %s", err)
	}

	fmt.Println("Working branch:", currentBranch.GetRef())

	// Create a file in the branch.
	// It doesn't complain if it exists, and just returns the same commit.
	client.Repositories.CreateFile(ctx, username, currentRepo.GetName(), "README.md", &github.RepositoryContentFileOptions{
		Message: github.String("Creating readme"),
		Content: []byte(""),
		Branch:  currentBranch.Ref,
	})

	// Get content of a file or directory
	f, _, _, err := client.Repositories.GetContents(ctx, username, currentRepo.GetName(), "README.md", nil)
	if err != nil {
		log.Fatalf("Error getting content of the file or directory: %v", err)
	}
	content, err := f.GetContent()
	if err != nil {
		log.Fatalf("Error getting content of the file: %v", err)
	}
	// Add a line with the current time to the file
	fmt.Println("Current content:", content)
	content = fmt.Sprintf("%s\n- Updated from API at %s\n", content, time.Now().String())

	// Update the file in the branch
	sha := github.String(f.GetSHA())

	fmt.Println("Current SHA:", *sha)
	resSha, _, _ := client.Repositories.GetCommitSHA1(ctx, username, currentRepo.GetName(), currentBranch.GetRef(), *sha)
	fmt.Println("ResSHA:", resSha)

	res, _, err := client.Repositories.UpdateFile(ctx, username, currentRepo.GetName(), "README.md", &github.RepositoryContentFileOptions{
		Message: github.String("Updating readme"),
		Content: []byte(content),
		Branch:  currentBranch.Ref,
		SHA:     f.SHA,
		// SHA: &resSha,
	})
	if err != nil {
		log.Fatalf("Error updating file: %v", err)
	}

	fmt.Printf("Updated file: %v\n", res.GetHTMLURL())

	// Open a pull request with that branch and commit
	// pr, _, err := client.PullRequests.Create(ctx, username, repo.GetName(), &github.NewPullRequest{
	// 	Title: github.String("Testing multigitter file creation and PR"),
	// 	Head:  branch.Ref,
	// 	Base:  github.String("main"),
	// 	Body:  github.String("And that's all"),
	// })
	// if err != nil {
	// 	log.Fatalf("Error creating pull request: %s", err)
	// }

	// fmt.Println("Pull request:", pr.GetHTMLURL())
}

func getOrCreateBranch(client *github.Client, owner string, repo string, branchName string, baseBranch string, ctx context.Context) (ref *github.Reference, err error) {
	if branchName == baseBranch {
		return nil, errors.New("the commit branch does not exist but `-base-branch` is the same as `-commit-branch`")
	}

	if baseBranch == "" {
		return nil, errors.New("the `-base-branch` should not be set to an empty string when the branch specified by `-commit-branch` does not exists")
	}

	if ref, _, err = client.Git.GetRef(ctx, owner, repo, "refs/heads/"+branchName); err == nil {
		fmt.Printf("Branch '%s' already exists, returning it\n", branchName)
		return ref, nil
	}

	var baseRef *github.Reference
	if baseRef, _, err = client.Git.GetRef(ctx, owner, repo, "refs/heads/"+baseBranch); err != nil {
		return nil, err
	}
	newRef := &github.Reference{Ref: github.String("refs/heads/" + branchName), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	ref, _, err = client.Git.CreateRef(ctx, owner, repo, newRef)
	return ref, err
}

func getBranch(client *github.Client, owner string, repo string, branchName string, ctx context.Context) (ref *github.Reference, err error) {
	if ref, _, err = client.Git.GetRef(ctx, owner, repo, "refs/heads/"+branchName); err != nil {
		return nil, err
	}
	return ref, nil
}
