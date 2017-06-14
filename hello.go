package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"gopkg.in/rjz/githubhook.v0"
)

const (
	Secret = "hello"
)

// hello world, the web server
func HelloServer(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello, world!\n")
	hook, err := githubhook.Parse([]byte(Secret), req)
	if err != nil {
		log.Fatalf("Err:\n %v\n\n", err)
	}
	log.Printf("Request:\n %v\n", req)
	log.Printf("Hook Event: \n %v\n", hook.Event)
	handleEvent(hook.Event, hook.Payload)
}

func handleEvent(event string, payload []byte) {
	switch event {
	case "pull_request":
		err := PREvent(payload)
		if err != nil {
			log.Println(err)
		}
	case "push":
		PushEvent(payload)
	}
}

func PREvent(payload []byte) error {
	evt := github.PullRequestEvent{}
	client := github.NewClient(nil)

	// Unmarshal the PR Event
	if err := json.Unmarshal(payload, &evt); err != nil {
		fmt.Println("Invalid JSON?", err)
		return err
	}
	// Log basic info
	log.Printf("Pull Request:\n\tAction:\n\t\t%v\n", *evt.Action)
	log.Printf("\tRepo:\n\t\t%v\n", *evt.Repo.Name)
	// Derive changed files
	files, resp, err := client.PullRequests.ListFiles(context.TODO(),
		*evt.Repo.Owner.Login, *evt.Repo.Name, evt.GetNumber(), nil)
	if err != nil {
		log.Printf("Error: %v\nResp: %v", err, resp)
		return err
	}
	// Print changed files
	for _, file := range files {
		log.Printf("%v\n", *file.Filename)
	}
	url, resp, err := client.Repositories.GetArchiveLink(context.TODO(),
		*evt.Repo.Owner.Login, *evt.Repo.Name, github.Tarball,
		&github.RepositoryContentGetOptions{Ref: *evt.PullRequest.Head.Ref})

	// Download the repository proposing to be merged
	filename := "head.tar.gz"
	wd, err := os.Getwd()
	tmpdir := path.Join(wd, "tmp")
	log.Printf("\nDownloading archive from %v\n", url.String())
	err = downloadFile(filename, url.String())
	if err != nil {
		return fmt.Errorf("%v while downloading from url: %v", err, url.Path)
	}
	os.Mkdir(tmpdir, 0775)
	untar := exec.Command("tar", "-xvf", filename, "-C", tmpdir)
	err = untar.Start()
	if err != nil {
		return fmt.Errorf("%v while executing %v", err, untar)
	}
	untar.Wait()
	log.Printf("\nUnwrapped the archive\n")

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
	clonePairs, err := cloneParse(string(out))
	_ = clonePairs
	// Consider only the clones that are in the diff

	return err
}

type Loc struct {
	Filename string
	Byte     uint
	End      uint
	Line     uint
}

type ClonePair struct {
	First  Loc
	Second Loc
}

func cloneParse(data string) ([]ClonePair, error) {
	files := make(map[int]string)
	pairs := make([]ClonePair, 0)
	readingSourceFiles := false
	readingClones := false
	for _, line := range strings.Split(data, "\n") {
		if strings.ContainsAny(line, "{}") {
			readingSourceFiles = false
			readingClones = false
			switch {
			case strings.Contains(line, "source_files"):
				readingSourceFiles = true
			case strings.Contains(line, "clone_pairs"):
				readingClones = true
			}
		} else if readingSourceFiles {
			words := strings.Split(line, "\t")
			fileid, err := strconv.Atoi(words[0])
			if err != nil {
				log.Println(err)
			}
			files[fileid] = strings.TrimSpace(words[1])
		} else if readingClones {
			words := strings.Split(line, "\t")
			if len(words) < 3 {
				return pairs, fmt.Errorf("Error parsing clones")
			}
			_, first, second := words[0], words[1], words[2]
			l1, err := parseLoc(first, files)
			if err != nil {
				return pairs, err
			}
			l2, err := parseLoc(second, files)
			if err != nil {
				return pairs, err
			}
			pairs = append(pairs, ClonePair{First: l1, Second: l2})
		}
	}
	return pairs, nil
}

func parseLoc(desc string, files map[int]string) (Loc, error) {
	str := strings.Split(desc, ".")
	fileId, err := strconv.Atoi(str[0])
	if err != nil {
		return Loc{}, err
	}
	interval := strings.Split(str[1], "-")
	bytenum, err := strconv.Atoi(interval[0])
	end, err2 := strconv.Atoi(interval[1])
	if err != nil {
		return Loc{}, err
	} else if err2 != nil {
		return Loc{}, err2
	}

	filename := files[fileId]
	return Loc{Filename: filename, Byte: uint(bytenum), End: uint(end)}, nil
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

func main() {
	http.HandleFunc("/", HelloServer)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
