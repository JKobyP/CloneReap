package main

import (
    "path"
    "os"
    "net/http"
    "io"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/github"
	"gopkg.in/rjz/githubhook.v0"
	"io"
	"log"
	"net/http"
    "context"
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
		PREvent(payload)
	case "push":
		PushEvent(payload)
	}
}

func PREvent(payload []byte) {
	evt := github.PullRequestEvent{}
	client := github.NewClient(nil)

    // Unmarshal the PR Event
	if err := json.Unmarshal(payload, &evt); err != nil {
		fmt.Println("Invalid JSON?", err)
        return
	}
    // Log basic info
	log.Printf("Pull Request:\n\tAction:\n\t\t%v\n",*evt.Action)
    log.Printf("\tRepo:\n\t\t%v\n", *evt.Repo.Name)
    // Derive changed files
	files, resp, err := client.PullRequests.ListFiles(context.TODO(),
		*evt.Repo.Owner.Login, *evt.Repo.Name, evt.GetNumber(), nil)
	if err != nil {
        log.Printf("Error: %v\nResp: %v", err,resp)
        return
	}
    // Print changed files
	for _, file := range files {
		log.Printf("%v\n", *file.Filename)
	}
    url, resp, err := client.Repositories.GetArchiveLink(context.TODO(),
        *evt.Repo.Owner.Login, *evt.Repo.Name, github.Tarball,
        &github.RepositoryContentGetOptions{Ref:*evt.PullRequest.Head.Ref})

    filename := "head.tar.gz"
    tmpdir := "tmp"
    log.Printf("\nDownloading archive\n")
    err := downloadFile(filename, url)
    os.Mkdir(tmpdir, 0644)
    untar := exec.Command("tar", "-xvf", filename, "-C", tmpdir)
    err = untar.Start()
    untar.Wait()
    filename = path.Join(tmpdir, filename)

    log.Printf("\nUnwrapped the archive\n")
    log.Printf("\nExecuting code clone detector\n")
    ccfx := exec.Command("ccfx", "D", "cpp", "-d", filename)
    out, err = exec.Command("ccfx", "P", "a.ccfxd")
}

func downloadFile(filepath string, url string) (err error) {
  // Create the file
  out, err := os.Create(filepath)
  if err != nil  {
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
  if err != nil  {
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
