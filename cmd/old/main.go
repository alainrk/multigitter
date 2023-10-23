package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

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

	token := os.Getenv("GITHUB_TOKEN")
	// password := os.Getenv("GITHUB_PASSWORD")

	// Create a GitHub client
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	l, _, err := client.Repositories.List(ctx, "", nil)
	if err != nil {
		log.Fatalf("Error listing repos: %s", err)
	}
	// TODO: Just to test
	fmt.Println("Repos:", len(l))

	// Create multigitter folder if it doesn't exist
	if _, err := os.Stat("/tmp/multigitter"); os.IsNotExist(err) {
		os.Mkdir("/tmp/multigitter", 0755)
	}

	// Set a parent folder for this session
	parentFolder := fmt.Sprintf("/tmp/multigitter/%s", uuid.New().String())

	repositories := map[string]string{
		"multigitter-test-1": "alainrk/multigitter-test-1",
		"multigitter-test-2": "alainrk/multigitter-test-2",
		"multigitter-test-3": "alainrk/multigitter-test-3",
	}

	// Loop through the list of repositories
	for name, repo := range repositories {
		fmt.Printf("\n------------------------------\nCloning repo: %s\n", repo)
		err := cloneRepo(name, repo, parentFolder)

		// All or nothing behavior
		if err != nil {
			log.Fatalf("Error cloning repo: %s", err)
		}
	}

	// Show content of the folder
	cmd := exec.Command("ls", "-la", parentFolder)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

// cloneRepo clones a repository using gh cli inside a given parent folder
func cloneRepo(name, repo, parentFolder string) error {
	folder := fmt.Sprintf("%s/%s", parentFolder, name)
	errors := []error{}

	// Sometimes gh gives "Connection reset by peer", so we have to retry a few times
	attempts := 3
	for i := 0; i < attempts; i++ {
		cmd := exec.Command("gh", "repo", "clone", repo, folder)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err == nil {
			return nil
		}

		errors = append(errors, err)

		// Remove the folder to avoid gh complaining
		err = os.RemoveAll(folder)
		if err != nil {
			log.Fatalf("Error removing folder %s: %s", folder, err)
		}

		if i <= attempts-1 {
			fmt.Printf("Retrying (%d/%d)...\n", i+1, attempts)
			continue
		}
	}

	errorMsg := ""
	for _, err := range errors {
		errorMsg += fmt.Sprintf("%s\n", err)
	}

	return fmt.Errorf("failed to clone repository %s: %v", repo, errorMsg)
}
