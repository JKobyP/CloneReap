package api

import (
	"github.com/boltdb/bolt"
	"io"
	"os"
)

type Model interface {
	Save(table string, key string, val interface{}) error
	Retrieve(table, key string) interface{}
	Close() error
}

func NewModel() (Model, error) {
	return NewBoltModel()
}

func NewBolt() (*Bolt, error) {
	db, err := bolt.Open("clonereap.db", 0600, nil)
	return &Bolt{db}, err
}

type Bolt struct {
	DB *bolt.DB
}

func (b *Bolt) Save(table, key string, value interface{}) error {
	b.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(table))
		if err != nil {
			return err
		}
	})
}
