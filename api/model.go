package api

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"log"

	"github.com/boltdb/bolt"
	"github.com/jkobyp/clonereap/clone"
	"github.com/pkg/errors"
)

var model Model = Model{}

type Model struct {
	*Bolt
}

/*
type Model interface {
	SavePR(prid int, clones []clone.ClonePairs)
	RetrievePR(table, key string) []clone.ClonePairs
	Close() error
}
*/

func GetModel() (Model, error) {
	if model == (Model{}) {
		model, err := NewModel()
		return model, err
	} else {
		return model, nil
	}
}

func NewModel() (Model, error) {
	m, e := NewBolt()
	return Model{m}, e
}

func NewBolt() (*Bolt, error) {
	db, err := bolt.Open("clonereap.db", 0600, nil)
	return &Bolt{db}, err
}

type Bolt struct {
	DB *bolt.DB
}

func (b *Bolt) SaveCFiles(prid int) error {
	log.Printf("Did NOT store cfiles")
	return nil
}

func (b *Bolt) SaveRepo(fullname string, prid int) error {
	//key := []byte(fullname)
	var val []byte
	binary.LittleEndian.PutUint32(val, uint32(prid))

	// TODO: Retrieve current val, append new prid to it, update
	log.Printf("Did NOT store PR id by fullname")

	return nil
}

func toByteBuffer(v interface{}) []byte {
	buf := bytes.NewBuffer([]byte(""))
	enc := gob.NewEncoder(buf)
	enc.Encode(v)
	return buf.Bytes()
}

func decodeCloneSlice(b []byte) ([]clone.ClonePair, error) {
	var cps []clone.ClonePair
	err := gob.NewDecoder(bytes.NewBuffer(b)).Decode(&cps)
	return cps, err
}

func (b *Bolt) SavePR(prid int, clones []clone.ClonePair) error {
	var key []byte
	binary.LittleEndian.PutUint32(key, uint32(prid))
	val := toByteBuffer(clones)

	err := b.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("pr-pairs"))
		if err != nil {
			return err
		}
		return bucket.Put(key, val)
	})
	return err
}

func (b *Bolt) RetrievePR(prid int) ([]clone.ClonePair, error) {
	var key []byte
	binary.LittleEndian.PutUint32(key, uint32(prid))
	var val []byte

	b.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("pr-pairs"))
		v := b.Get(key)
		copy(v, val)
		return nil
	})

	cps, err := decodeCloneSlice(val)
	if err != nil {
		return nil, errors.Wrap(err, "decoding clone slice")
	}

	return cps, nil
}
