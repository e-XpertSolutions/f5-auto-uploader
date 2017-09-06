package main

import (
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

func watchDir(f5Client *f5.Client, logger logger, cfg watchConfig) (*watchRoutine, error) {
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
				switch {
				case e.isCreate():
					logger.Notice("CREATE: ", e.Name)
					// TODO(gilliek): create new file
				case e.isWrite():
					logger.Notice("WRITE: ", e.Name)
					// TODO(gilliek): update existing file
				case e.isRename():
					logger.Notice("RENAME: ", e.Name)
					// TODO(gilliek): scan the directory to see what to remove?
				case e.isRemove():
					logger.Notice("REMOVE: ", e.Name)
					// TODO(gilliek): delete?
				}
			case err := <-watcher.Errors:
				logger.Error("watcher error: ", err)
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
