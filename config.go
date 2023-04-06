package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/barkimedes/go-deepcopy"
	"github.com/cespare/xxhash/v2"
	"golang.org/x/xerrors"
)

// File defines a config file
type File[T any] struct {
	logger      *log.Logger
	path        string
	userConfig  T
	fileHash    uint64
	updated     func(T)
	watching    *atomic.Bool
	interval    time.Duration
	configMutex sync.Mutex
}

// New creates a new config file and starts watching for changes
// filepath is the JSON formatted file to monitor
// c is the interface to attempt to marshal the file into
// updated is called when there are updates to the file
func New[T any](filePath string, checkInterval time.Duration, l *log.Logger, updated func(T)) (*File[T], error) {
	ap, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	f := &File[T]{path: ap, interval: checkInterval, logger: l, updated: updated}
	f.watching = &atomic.Bool{}
	f.configMutex = sync.Mutex{}
	f.watching.Store(true)

	// start the watcher
	go f.watch(ap)

	// parse the config file first time and return
	f.configMutex.Lock()
	defer f.configMutex.Unlock()
	return f, f.loadData()
}

// Get returns the config serialized from the JSON, the returned config is always a copy
// and is goroutine safe
func (f *File[T]) Get() T {
	f.configMutex.Lock()
	defer f.configMutex.Unlock()

	copy, err := deepcopy.Anything(f.userConfig)
	if err != nil {
		f.logger.Printf("[ERROR] unable to copy struct: %s", err)
	}

	return copy.(T)
}

// Close the FileConfig and remove all watchers
func (f *File[T]) Close() {
	f.watching.Store(false)
}

// load the data from the config into the defined structure
func (f *File[T]) loadData() error {
	cf, err := os.Open(f.path)
	if err != nil {
		return xerrors.Errorf("Unable to open config file %s: %w", f.path, err)
	}
	defer cf.Close()

	err = json.NewDecoder(cf).Decode(&f.userConfig)
	if err != nil {
		return xerrors.Errorf("Unable to decode config file %s: %w", f.path, err)
	}

	hash, err := generateHash(f.path)
	if err != nil {
		log.Println("[ERROR] unable to generate hash:", err)
	}

	f.fileHash = hash

	return nil
}

// watch a file for changes
func (f *File[T]) watch(filepath string) {
	for f.watching.Load() {
		f.configMutex.Lock()

		// generate a hash of the file
		hash, err := generateHash(filepath)
		if err != nil {
			log.Println("[ERROR] unable to generate hash:", err)
		}

		//log.Printf("[DEBUG] hash: %d, new: %d\n", f.fileHash, hash)

		// file has changed
		if f.fileHash != 0 && f.fileHash != hash {
			log.Printf("[DEBUG] config file changed")

			err := f.loadData()
			if err != nil {
				log.Printf("[ERROR] Unable to load data %s\n", err)

				f.configMutex.Unlock()
				continue
			}

			if f.updated != nil {
				copy, err := deepcopy.Anything(f.userConfig)
				if err != nil {
					f.logger.Printf("[ERROR] unable to copy struct: %s\n", err)

					f.configMutex.Unlock()
					continue
				}

				f.updated(copy.(T))
			}
		}

		f.configMutex.Unlock()
		time.Sleep(f.interval)
	}
}

func generateHash(path string) (uint64, error) {
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("unable to read file: %s", err)
	}

	xh := xxhash.New()
	xh.Write(d)
	return xh.Sum64(), nil
}
