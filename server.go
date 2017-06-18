package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"

	"github.com/google/go-github/github"
	"github.com/jkobyp/CloneReap/clone"
	"github.com/pkg/errors"
	"gopkg.in/rjz/githubhook.v0"
)

const (
	Secret = "hello"
)

// hello world, the web server
func HookServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello, world!\n")
	hook, err := githubhook.Parse([]byte(Secret), req)
	if err != nil {
		log.Printf("%v", errors.Wrap(err, "parsing githubhook"))
		return
	}
	log.Printf("Request:\n %v\n", req)
	log.Printf("Hook Event: \n %v\n", hook.Event)
	err = handleEvent(hook.Event, hook.Payload)
	if err != nil {
		log.Printf("%v", errors.Wrap(err, "handling event"))
	}
}

func handleEvent(event string, payload []byte) error {
	switch event {
	case "pull_request":
		return PREvent(payload)
	case "push":
		PushEvent(payload)
	}
	return nil
}

func processPRHook(payload []byte) (github.PullRequestEvent, error) {
	evt := github.PullRequestEvent{}
	// Unmarshal the PR Event
	if err := json.Unmarshal(payload, &evt); err != nil {
		return evt, errors.Wrap(err, "unmarshaling json")
	}
	// Log basic info
	log.Printf("Pull Request:\n")
	log.Printf("\tAction: %v\n", *evt.Action)
	log.Printf("\tRepo: %v\n", *evt.Repo.Name)

	return evt, nil
}

func getFilesAndRepo(client *github.Client,
	evt *github.PullRequestEvent) ([]*github.CommitFile, string, error) {
	// Derive changed files
	files, _, err := client.PullRequests.ListFiles(context.TODO(),
		*evt.Repo.Owner.Login, *evt.Repo.Name, evt.GetNumber(), nil)
	if err != nil {
		//log.Printf("Error: %v\nResp: %v", err, resp)
		return nil, "", errors.Wrap(err, "listing changed files")
	}
	// Print changed files
	for _, file := range files {
		log.Printf("%Changed: v\n", *file.Filename)
	}
	url, _, err := client.Repositories.GetArchiveLink(context.TODO(),
		*evt.Repo.Owner.Login, *evt.Repo.Name, github.Tarball,
		&github.RepositoryContentGetOptions{Ref: *evt.PullRequest.Head.Ref})

	// Download the repository proposing to be merged
	filename := "head.tar.gz"
	log.Printf("\nDownloading archive from %v\n", url.String())
	err = downloadFile(filename, url.String())
	if err != nil {
		return nil, "", fmt.Errorf("%v while downloading from url: %v", err, url.Path)
	}
	return files, filename, nil
}

// PREvent takes the byteslice of json encoded
// payload delivered with the event hook, triggers
// the clone detector, and stores the results
func PREvent(payload []byte) error {
	client := github.NewClient(nil)
	// unmarshal the PR Hook
	evt, err := processPRHook(payload)
	if err != nil {
		return err
	}

	// get a list of changed files and download the new repo
	files, filename, err := getFilesAndRepo(client, &evt)
	if err != nil {
		return err
	}

	// extract the repo
	wd, err := os.Getwd()
	tmpdir := path.Join(wd, "tmp")
	os.Mkdir(tmpdir, 0775)
	untar := exec.Command("tar", "-xvf", filename, "-C", tmpdir)
	err = untar.Start()
	if err != nil {
		return fmt.Errorf("%v while executing %v", err, untar)
	}
	untar.Wait()
	log.Printf("\nUnwrapped the archive\n")

	// get the root directory
	contents, err := ioutil.ReadDir(tmpdir)
	if len(contents) != 1 {
		return fmt.Errorf("wrong number of items in the tmp directory: expected 1")
	} else if err != nil {
		return err
	}
	root := path.Join(tmpdir, contents[0].Name())

	// Find all clones in the repository
	log.Printf("\nExecuting code clone detector\n")
	ccfx := exec.Command("ccfx", "D", "cpp", "-d", tmpdir)
	err = ccfx.Start()
	if err != nil {
		return fmt.Errorf("%v while executing %v", err, ccfx)
	}
	err = ccfx.Wait()
	if err != nil {
		return fmt.Errorf("%v while executing %v", err, ccfx)
	}
	out, err := exec.Command("ccfx", "P", "a.ccfxd").Output()
	if err != nil {
		return err
	}

	// Parse the output
	clonePairs, err := clone.CloneParse(string(out))
	fmt.Printf("\n***\tClone Pairs\t***\n")
	for _, pair := range clonePairs {
		fmt.Printf("%v\n", pair)
	}

	// Consider only the clones that are in the diff
	relPairs := make([]clone.ClonePair, 0)
	for _, pair := range clonePairs {
		contains := false
		for _, cfile := range files {
			// TODO match CommitFile paths to absolute path returned from
			// clonedetector
			if path.Join(root, cfile.GetFilename()) == pair.First.Filename {
				contains = true
				break
			}
		}
		if contains {
			relPairs = append(relPairs, pair)
		}
	}
	return err
}

func downloadFile(filepath string, url string) (err error) {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func PushEvent(payload []byte) {
	log.Printf("PushEvent: Not Implemented")
}
