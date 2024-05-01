package config

import (
	"github.com/fsnotify/fsnotify"
	"sync"
)

type Reloader struct {
	configInstance interface{}
	mutex          *sync.Mutex

	watcher *fsnotify.Watcher
	opList  []fsnotify.Op

	outlet chan interface{}
}

func NewReloader(config interface{}) *Reloader {
	return &Reloader{
		configInstance: config,
		mutex:          &sync.Mutex{},

		opList: []fsnotify.Op{fsnotify.Write, fsnotify.Create},
	}
}

func (r *Reloader) WithOps(ops ...fsnotify.Op) {
	r.opList = ops
}

func (r *Reloader) WatchPath(path string) error {
	var err error
	r.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	err = r.watcher.Add(path)
	if err != nil {
		return err
	}
	_ = LoadConfig(r.configInstance, path)
	go func() {
		for event := range r.watcher.Events {
			for _, op := range r.opList {
				if event.Op == op {
					r.mutex.Lock()
					_ = LoadConfig(r.configInstance, path)
					if r.outlet != nil {
						r.outlet <- r.configInstance
					}
					r.mutex.Unlock()
				}
			}
		}
	}()
	return err
}

func (r *Reloader) FetchConfig() interface{} {
	if r.mutex != nil {
		r.mutex.Lock()
		defer r.mutex.Unlock()
	}
	return r.configInstance
}

func (r *Reloader) SubscribeConfig() chan interface{} {
	r.outlet = make(chan interface{}, 1)
	return r.outlet
}

func (r *Reloader) Stop() {
	_ = r.watcher.Close()
}
