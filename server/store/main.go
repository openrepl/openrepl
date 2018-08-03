package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// DirStore is a KVStore backed by a directory.
type DirStore struct {
	Dir string
}

func (ds DirStore) path(key []byte) string {
	return filepath.Join(ds.Dir, hex.EncodeToString(key))
}

// Set sets key-value pair.
func (ds DirStore) Set(key, value []byte) error {
	return ioutil.WriteFile(ds.path(key), value, 0600)
}

// Get gets a value with the given key.
// If the KV pair is not set, returns ErrNotExist.
func (ds DirStore) Get(key []byte) ([]byte, error) {
	dat, err := ioutil.ReadFile(ds.path(key))
	if err != nil {
		if os.IsNotExist(err) {
			err = ErrNotExist
		}
		return nil, err
	}
	return dat, err
}

// MemStore is an in-memory KVStore.
type MemStore struct {
	m sync.Map
}

// Set sets key-value pair.
func (ms *MemStore) Set(key, value []byte) error {
	ms.m.Store(hex.EncodeToString(key), value)
	return nil
}

// Get gets a value with the given key.
// If the KV pair is not set, returns ErrNotExist.
func (ms *MemStore) Get(key []byte) ([]byte, error) {
	dat, ok := ms.m.Load(hex.EncodeToString(key))
	if !ok {
		return nil, ErrNotExist
	}
	return dat.([]byte), nil
}

// ErrNotExist is an error indicating that a KV pair is not set.
var ErrNotExist = errors.New("kv pair does not exist")

// KVStore is an interface for a key-value store.
type KVStore interface {
	// Set sets key-value pair.
	Set(key, value []byte) error

	// Get gets a value with the given key.
	// If the KV pair is not set, returns ErrNotExist.
	Get(key []byte) ([]byte, error)
}

// Code is a struct containing code with metadata.
type Code struct {
	Code     string `json:"code"`
	Language string `json:"language"`
}

// CodeStore is a code storage system using a KVStore.
type CodeStore struct {
	KV KVStore
}

// Get retrieves a Code struct from the store.
func (cs CodeStore) Get(key string) (Code, error) {
	// decode key
	k, err := hex.DecodeString(key)
	if err != nil {
		return Code{}, err
	}

	// get code from store
	dat, err := cs.KV.Get(k)
	if err != nil {
		return Code{}, err
	}

	// unmarshal code
	var c Code
	err = json.Unmarshal(dat, &c)
	if err != nil {
		return Code{}, err
	}

	return c, nil
}

// Store stores Code into tha KVStore.
func (cs CodeStore) Store(c Code) (string, error) {
	// encode Code
	dat, err := json.Marshal(&c)
	if err != nil {
		return "", err
	}

	// hash Code to generate key
	hash := sha256.Sum256(dat)

	// save in KVStore
	err = cs.KV.Set(hash[:], dat)
	if err != nil {
		return "", err
	}

	// encode hash key into text format
	return hex.EncodeToString(hash[:]), nil
}

func main() {
	var driver string
	var dir string
	flag.StringVar(&driver, "driver", "mem", "driver for key-value store")
	flag.StringVar(&dir, "dir", "", "directory to use for dir driver")
	flag.Parse()

	// initialize KVStore
	var kv KVStore
	switch driver {
	case "mem":
		kv = new(MemStore)
	case "dir":
		kv = DirStore{dir}
	default:
		panic(fmt.Errorf("unrecognized driver %s", driver))
	}

	cs := CodeStore{kv}

	// store Code
	http.HandleFunc("/store", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not supported", http.StatusMethodNotAllowed)
			return
		}

		var c Code
		err := json.NewDecoder(r.Body).Decode(&c)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to decode request: %s", err.Error()), http.StatusBadRequest)
			return
		}

		key, err := cs.Store(c)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to store: %s", err.Error()), http.StatusInternalServerError)
			return
		}

		// write back key
		w.Write([]byte(key))
	})

	// load Code
	http.HandleFunc("/load", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not supported", http.StatusMethodNotAllowed)
			return
		}

		key := r.URL.Query().Get("key")
		c, err := cs.Get(key)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to load: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(c)
	})

	panic(http.ListenAndServe(":80", nil))
}
