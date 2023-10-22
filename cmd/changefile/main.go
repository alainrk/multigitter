package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	username := os.Getenv("GITHUB_USERNAME")
	token := os.Getenv("GITHUB_TOKEN")

	fmt.Println("Github username: %s", username)

	// Create a GitHub client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	name := "alainrk/multigitter-test-1"
	l, _, err := client.Search.Repositories(ctx, name, &github.SearchOptions{})
	if err != nil {
		log.Fatalf("Error listing repos: %s", err)
	}
	if len(l.Repositories) != 1 {
		log.Fatalf("Unexpected number of repositories for repository '%s': %d", name, len(l.Repositories))
	}
	repo := l.Repositories[0]

	branch, err := getBranch(client, username, repo.GetName(), "test-branch", "main", ctx)
	if err != nil {
		log.Fatalf("Error getting branch: %s", err)
	}

	fmt.Println("Branch:", branch)

	// 	client.Repositories.CreateFile(ctx, username, repo.GetName(), "README.md", &github.RepositoryContentFileOptions{
	// 		Message: github.String("creating readme"),
	// 		Content: []byte(`
	// # MultiGitter Test 1

	// TODO: Add content
	// 		`),
	// 	})

	f, d, r, err := client.Repositories.GetContents(ctx, username, repo.GetName(), "README.md", nil)
	if err != nil {
		log.Fatalf("Error getting contents: %s", err)
	}

	fmt.Println("File:", f)
	fmt.Println("Directory:", d)
	fmt.Println("Repo:", r)

	// repositories := map[string]string{
	// 	"multigitter-test-1": "alainrk/multigitter-test-1",
	// 	"multigitter-test-2": "alainrk/multigitter-test-2",
	// 	"multigitter-test-3": "alainrk/multigitter-test-3",
	// }

}

func getBranch(client *github.Client, owner string, repo string, branchName string, baseBranch string, ctx context.Context) (ref *github.Reference, err error) {
	if ref, _, err = client.Git.GetRef(ctx, owner, repo, "refs/heads/"+branchName); err == nil {
		return ref, nil
	}

	if branchName == baseBranch {
		return nil, errors.New("the commit branch does not exist but `-base-branch` is the same as `-commit-branch`")
	}

	if baseBranch == "" {
		return nil, errors.New("the `-base-branch` should not be set to an empty string when the branch specified by `-commit-branch` does not exists")
	}

	var baseRef *github.Reference
	if baseRef, _, err = client.Git.GetRef(ctx, owner, repo, "refs/heads/"+baseBranch); err != nil {
		return nil, err
	}
	newRef := &github.Reference{Ref: github.String("refs/heads/" + branchName), Object: &github.GitObject{SHA: baseRef.Object.SHA}}
	ref, _, err = client.Git.CreateRef(ctx, owner, repo, newRef)
	return ref, err
}
