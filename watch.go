package main

import (
	"fmt"
	"path/filepath"

	"github.com/e-XpertSolutions/f5-rest-client/f5"
	fsnotify "gopkg.in/fsnotify.v1"
)

type watchEvent fsnotify.Event

func (e watchEvent) isCreate() bool {
	return e.Op&fsnotify.Create == fsnotify.Create
}

func (e watchEvent) isWrite() bool {
	return e.Op&fsnotify.Write == fsnotify.Write
}

func (e watchEvent) isRemove() bool {
	return e.Op&fsnotify.Remove == fsnotify.Remove
}

func (e watchEvent) isRename() bool {
	return e.Op&fsnotify.Rename == fsnotify.Rename
}

func (e watchEvent) isChmod() bool {
	return e.Op&fsnotify.Chmod == fsnotify.Chmod
}

type watchRoutine struct {
	watcher *fsnotify.Watcher
	stopCh  chan struct{}
}

func watchDir(f5Client *f5.Client, l logger, cfg watchConfig) (*watchRoutine, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	stopCh := make(chan struct{})
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				e := watchEvent(event)
				if e.isChmod() {
					continue
				}

				fmt.Printf("testing %q against %v", e.Name, cfg.Exclude)
				if isExcluded(filepath.Base(e.Name), cfg.Exclude) {
					l.Noticef("skipping %q due to an exclusion pattern defined in the configuration file", e.Name)
					continue
				}

				tx, err := f5Client.Begin()
				if err != nil {
					l.Errorf("cannot start f5 transaction for file %q", e.Name)
					continue
				}

				switch {
				case e.isCreate():
					l.Noticef("event received %q for file %q", "CREATE", e.Name)
					err = uploadNewFile(tx, filepath.Base(e.Name), e.Name)
				case e.isWrite():
					l.Noticef("event received %q for file %q", "WRITE", e.Name)
					err = uploadExistingFile(tx, filepath.Base(e.Name), e.Name)
				case e.isRename():
					l.Noticef("event received %q for file %q", "RENAME", e.Name)
					err = deleteFile(tx, filepath.Base(e.Name))
				case e.isRemove():
					l.Noticef("event received %q for file %q", "REMOVE", e.Name)
					err = deleteFile(tx, filepath.Base(e.Name))
				}
				if err != nil {
					l.Errorf("cannot upload file %q: %v", e.Name, err)
					continue
				}

				if err := tx.Commit(); err != nil {
					l.Errorf("cannot commit f5 transaction for file %q: %v", e.Name, err)
				}
			case err := <-watcher.Errors:
				l.Error("watcher error: ", err)
			case <-stopCh:
				return
			}
		}
	}()
	if err := watcher.Add(cfg.Dir); err != nil {
		stopCh <- struct{}{}
		return nil, err
	}
	return &watchRoutine{
		watcher: watcher,
		stopCh:  stopCh,
	}, nil
}

func (wr *watchRoutine) stop() error {
	return wr.watcher.Close()
}
