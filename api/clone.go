package api

import (
	"fmt"

	"github.com/google/go-github/github"
	"github.com/jkobyp/clonereap/clone"
	"github.com/pkg/errors"
)

type File struct {
	Path    string `json:"path"`
	Content []byte `json:"content"`
}

type Pr struct {
	Id     int               `json:"id"`
	Clones []clone.ClonePair `json:"clones"`
	Files  []File            `json:"files"`
}

func SavePrEvent(pr *github.PullRequest, clones []clone.ClonePair, files []File) error {
	fullname := pr.Base.Repo.GetFullName()
	err := saveRepo(fullname)
	if err != nil {
		return errors.Wrap(err, "saving pr")
	}
	err = savePr(fullname, pr.GetID())
	if err != nil {
		return errors.Wrap(err, "saving pr")
	}
	err = saveFiles(pr.GetID(), files)
	if err != nil {
		return errors.Wrap(err, "saving pr")
	}
	err = saveClones(pr.GetID(), clones)
	return err
}

func RetrievePrs(user, project string) ([]Pr, error) {
	prs := []Pr{}

	fullname := fmt.Sprintf("%s/%s", user, project)
	ids, err := getPrs(fullname)
	if err != nil {
		return nil, err
	}
	for _, id := range ids {
		clones, err := getClones(id)
		if err != nil {
			return nil, err
		}
		files, err := getFiles(id)
		if err != nil {
			return nil, err
		}
		pr := Pr{Id: id, Clones: clones, Files: files}
		prs = append(prs, pr)
	}
	return prs, err
}
