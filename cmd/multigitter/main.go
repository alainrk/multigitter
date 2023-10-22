package main

import (
	"context"
	"fmt"
	"log"
	"os"

	gitHttp "github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/go-git/go-git/v5"
	"github.com/google/go-github/github"
	"github.com/google/uuid"
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
	// password := os.Getenv("GITHUB_PASSWORD")

	// Create a GitHub client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	fmt.Println("GitHub client created:", client)

	uuid := uuid.New()
	folder := fmt.Sprintf("/tmp/%s", uuid.String())

	repositories := []string{
		"https://github.com/alainrk/multigitter-test-1.git",
		"https://github.com/alainrk/multigitter-test-2.git",
		"https://github.com/alainrk/multigitter-test-3.git",
	}

	// Clone options with authentication
	auth := &gitHttp.BasicAuth{
		Username: username,
		Password: token,
	}

	// Loop through the list of repositories
	for _, repo := range repositories {

		_, err := git.PlainClone(folder, false, &git.CloneOptions{
			URL:      repo,
			Progress: os.Stdout,
			Auth:     auth,
		})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Cloned repo:", repo)
	}
}
