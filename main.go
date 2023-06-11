package main

import (
	"flag"
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/xanzy/go-gitlab"
)

// GitlabClient - struct holding pointer to client
type GitLabClient struct {
	Client *gitlab.Client
}

// GLC - Global Gitlab Client struct
var GLC GitLabClient

func main() {
	gitLabToken := flag.String("token", "", "GitLab API Token")
	dest := flag.String("dest", "", "Destination to clone repositories")
	groupId := flag.Int("groupid", 0, "GitLab Group ID")
	privKeyFile := flag.String("key", "", "Path to SSH key")
	flag.Parse()

	if *gitLabToken == "" {
		log.Fatal("Please set gitlab token with -token=glpat-xxxxx")
	}

	if *dest == "" {
		log.Fatal("Please set destination with -dest=/home/Test/git")
	}

	if *groupId == 0 {
		log.Fatal("Please set groupID  with -groupid=1020304")
	}

	if *privKeyFile == "" {
		log.Fatal("Please set SSH private key path with -key=file.pem")
	}

	// Get Client
	initGitlabClient(*gitLabToken)

	// Load SSH Key
	keys, err := loadKeys(privKeyFile)
	if err != nil {
		log.Fatal(err)
	}

	//Fetch projects first
	projects := getAllProjects()

	// Loop through and clone
	for _, project := range projects {
		if project.Namespace.ID != *groupId {
			continue
		}

		// Clone!
		cloneProject(project, dest, keys)
	}
}

// initGitlabClient - boot gitlab client with request token
func initGitlabClient(token string) {
	git, err := gitlab.NewClient(token)
	if err != nil {
		log.Fatal(err)
	}

	GLC = GitLabClient{
		Client: git,
	}
}

// getAllProjects - Get all projects owned by account
func getAllProjects() []*gitlab.Project {
	var projects []*gitlab.Project
	opts := gitlab.ListProjectsOptions{
		OrderBy:    gitlab.String("created_at"),
		Sort:       gitlab.String("asc"),
		Archived:   gitlab.Bool(false),
		Membership: gitlab.Bool(true),
	}

	opts.PerPage = 100
	opts.Page = 1

	resultCount := 1
	for resultCount > 0 {
		log.Printf("Fetching all projects. Page: %d", opts.Page)
		projectResp, _, err := GLC.Client.Projects.ListProjects(&opts)
		opts.Page = opts.Page + 1
		if err != nil {
			log.Fatal(err)
		}

		projects = append(projects, projectResp...)
		resultCount = len(projectResp)
	}

	return projects
}

// loadKeys - Loads private key
func loadKeys(privateKeyFile *string) (*ssh.PublicKeys, error) {
	_, err := os.Stat(*privateKeyFile)
	if err != nil {
		log.Fatal(err)
	}

	return ssh.NewPublicKeysFromFile("git", *privateKeyFile, "")
}

// cloneProject - Clones to a local directory
func cloneProject(project *gitlab.Project, dest *string, publicKeys *ssh.PublicKeys) {
	log.Printf("Cloning: %s", project.NameWithNamespace)

	_, err := git.PlainClone(*dest+"/"+project.Name, false, &git.CloneOptions{
		URL:      project.SSHURLToRepo,
		Progress: os.Stdout,
		Auth:     publicKeys,
	})

	if err != nil {
		log.Fatal(err)
	}
}
