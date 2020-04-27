package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/xerrors"
)

// File defines a config file
type File struct {
	logger     *log.Logger
	path       string
	userConfig interface{}
	watcher    *fsnotify.Watcher
	updated    func()
}

// New creates a new config file and starts watching for changes
// filepath is the JSON formatted file to monitor
// c is the interface to attempt to marshal the file into
// updated is called when there are updates to the file
func New(fp string, c interface{}, l *log.Logger, updated func()) (*File, error) {
	ap, err := filepath.Abs(fp)
	if err != nil {
		return nil, err
	}

	f := &File{path: ap, userConfig: c, logger: l, updated: updated}
	go f.watch(ap)

	// sleep to allow the watch to setup
	time.Sleep(10 * time.Millisecond)

	return f, f.loadData()
}

// Close the FileConfig and remove all watchers
func (f *File) Close() {
	// watcher.Close can cause panic when running on CI in tests
	defer func() {
		if r := recover(); r != nil {
			f.logger.Printf("[ERROR] Recovered from panic %s\n", r)
		}
	}()

	f.watcher.Close()
}

// load the data from the config into the defined structure
func (f *File) loadData() error {
	cf, err := os.Open(f.path)
	if err != nil {
		return xerrors.Errorf("Unable to open config file %s: %w", f.path, err)
	}
	defer cf.Close()

	err = json.NewDecoder(cf).Decode(f.userConfig)
	if err != nil {
		return xerrors.Errorf("Unable to decode config file %s: %w", f.path, err)
	}

	return nil
}

// watch a file for changes
func (f *File) watch(filepath string) {
	// creates a new file watcher
	var err error
	f.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Printf("[ERROR] Unablet to create file watcher: %s\n", err)
	}

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-f.watcher.Events:
				if !ok {
					continue
				}
				// running on Docker we are not going to reliably get the Write or create event
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Printf("[DEBUG] Recieved event %#v\n", event)

					err := f.loadData()
					if err != nil {
						log.Printf("[ERROR] Unable to load data %s\n", err)
						continue
					}

					if f.updated != nil {
						f.updated()
					}
				}
			case err, ok := <-f.watcher.Errors:
				if !ok {
					continue
				}

				log.Printf("[ERROR] Received error from watcher %s\n", err)
			}
		}
	}()

	err = f.watcher.Add(filepath)
	if err != nil {
		log.Printf("[ERROR] Unable to watch file %s: %s\n", filepath, err)
	}

	<-done
}
