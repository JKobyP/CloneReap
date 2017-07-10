package api

import (
	"io"

	"github.com/google/go-github/github"
	"github.com/jkobyp/clonereap/clone"
	"github.com/pkg/errors"
)

type file struct {
	path    string
	content []byte
}

func SavePR(pr *github.PullRequest, clones []clone.ClonePair, files io.Reader) error {
	model, err := GetModel()
	if err != nil {
		return errors.Wrap(err, "saving pr")
	}
	err = model.SavePR(pr.GetID(), clones)
	if err != nil {
		return errors.Wrap(err, "saving pr")
	}
	err = model.SaveCFiles(pr.GetID())
	if err != nil {
		return errors.Wrap(err, "saving pr")
	}
	err = model.SaveRepo(pr.Head.Repo.GetFullName(), pr.GetID())
	return err
}

func RetrievePR(prid int) ([]clone.ClonePair, error) {
	model, err := GetModel()
	if err != nil {
		return nil, errors.Wrap(err, "getting pr")
	}

	clones, err := model.RetrievePR(prid)
	return clones, err
}
