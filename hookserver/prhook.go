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
	"github.com/jkobyp/clonereap/api"
	"github.com/jkobyp/clonereap/clone"
	"github.com/pkg/errors"
	"gopkg.in/rjz/githubhook.v0"
)

const (
	Secret = "hello"
)

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

func preProcessPRHook(payload []byte) (github.PullRequestEvent, error) {
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
		log.Printf("Changed: %v\n", *file.Filename)
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

func extract(filename string) (string, error) {
	wd, err := os.Getwd()
	tmpdir := path.Join(wd, "tmp")
	os.Mkdir(tmpdir, 0775)
	untar := exec.Command("tar", "-xvf", filename, "-C", tmpdir)
	err = untar.Start()
	untar.Wait()
	log.Printf("\nUnwrapped the archive\n")
	return tmpdir, err
}

// PREvent takes the byteslice of json encoded
// payload delivered with the event hook, triggers
// the clone detector, and stores the results
func PREvent(payload []byte) error {
	client := github.NewClient(nil)
	// unmarshal the PR Hook
	evt, err := preProcessPRHook(payload)
	if err != nil {
		return err
	}

	// get a list of changed files and download the new repo
	files, filename, err := getFilesAndRepo(client, &evt)
	if err != nil {
		return err
	}

	// extract the repo to a tmp directory
	tmpdir, err := extract(filename)
	if err != nil {
		return errors.Wrap(err, "while extracting repo")
	}

	// get the root directory
	contents, err := ioutil.ReadDir(tmpdir)
	if len(contents) != 1 {
		return fmt.Errorf("wrong number of items in the tmp directory: expected 1")
	} else if err != nil {
		return err
	}
	root := path.Join(tmpdir, contents[0].Name())

	clonePairs, err := clone.NewDetector().Detect(root)
	if err != nil {
		return err
	}

	fmt.Printf("\n***\tClone Pairs\t***\n")
	for _, pair := range clonePairs {
		fmt.Printf("%v\n", pair)
	}
	fmt.Printf("\n***\tChanged files \t***\n")
	for _, cfile := range files {
		fmt.Printf("%v\n", cfile)
	}

	// Consider only the clones that are in the diff
	relPairs := make([]clone.ClonePair, 0)
	relFiles := make(map[string]bool)
	for _, pair := range clonePairs {
		contains := false
		for _, cfile := range files {
			if path.Join(root, cfile.GetFilename()) == pair.First.Filename {
				relFiles[cfile.GetFilename()] = true
				contains = true
				break
			}
		}
		if contains {
			relPairs = append(relPairs, pair)
		}
	}

	// Create api.File for reach relevant clonepair
	fs := processFileset(root, relFiles)
    prs := stripRoot(root, relPairs)

	err = api.SavePrEvent(evt.PullRequest, prs, fs)
    if err != nil {
        log.Printf("%s", err)
    }
    err = cleanTmp(tmpdir)

	return err
}

func cleanTmp(dir string) error {
    return os.RemoveAll(dir)
}

func stripRoot(root string, pairs []clone.ClonePair) []clone.ClonePair {
    ret := make([]clone.ClonePair, len(pairs))
    for i, pair := range pairs {
        pair.First.Filename = pair.First.Filename[len(root):]
        pair.Second.Filename = pair.Second.Filename[len(root):]
        ret[i] = pair
    }
    return ret
}

func processFileset(root string, fileset map[string]bool) []api.File {
	ret := make([]api.File, 0)
	for file := range fileset {
		content, err := ioutil.ReadFile(path.Join(root,file))
		if err != nil {
			log.Println("Error reading file")
		}
		ret = append(ret, api.File{Path: file, Content: content})
	}
	return ret
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
