package ddns

import (
	"github.com/boltdb/bolt"
)

var (
	defaultStorePath   = "./domain.db"
	defaultStoreBucket = "domain"
)

type StoreConfig struct {
	Path   string `json:"path" toml:"path"`
	Bucket string `json:"bucket" toml:"bucket"`
}

type Store struct {
	path   string
	bucket string
	db     *bolt.DB
}

func NewStore(cfg *StoreConfig) (*Store, error) {
	path := cfg.Path
	if path == "" {
		path = defaultStorePath
	}

	bucketName := cfg.Bucket
	if bucketName == "" {
		bucketName = defaultStoreBucket
	}

	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &Store{
		path:   path,
		bucket: bucketName,
		db:     db,
	}, nil
}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) Get(key string) (interface{}, error) {
	var v []byte

	s.db.View(func(ctx *bolt.Tx) error {
		b := ctx.Bucket([]byte(s.bucket))
		v = b.Get([]byte(key))
		return nil
	})

	return string(v), nil
}

func (s *Store) Set(key string, value string) error {
	var err error

	s.db.Update(func(ctx *bolt.Tx) error {
		b := ctx.Bucket([]byte(s.bucket))
		err = b.Put([]byte(key), []byte(value))
		return err
	})

	return err
}
