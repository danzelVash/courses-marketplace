package storage

import (
	"sync"
	"time"

	"github.com/danzelVash/courses-marketplace"
	"github.com/sirupsen/logrus"
)

type DataStorage struct {
	sync.Mutex
	cache map[string]int
}

func NewDataStorage() *DataStorage {
	return &DataStorage{
		cache: make(map[string]int),
	}
}

func (ds *DataStorage) cleaner(key string) {
	time.Sleep(time.Minute)
	if _, ok := ds.Get(key); ok {
		err := ds.Delete(key)
		if err != nil {
			logrus.Errorf("memory leak (can`t delete code from DataStorage): %s", err.Error())
		} else {
			logrus.Infof("code deleted from DataStorage, storage state: %v", ds.cache)
		}
	} else {
		logrus.Infof("%s`s code already had deleted from DataStorage", key)
	}
}

func (ds *DataStorage) add(key string, val int) {
	ds.cache[key] = val
}

func (ds *DataStorage) get(key string) (int, bool) {
	if val, ok := ds.cache[key]; ok {
		return val, true
	}
	return 0, false
}

func (ds *DataStorage) delete(key string) error {
	if _, ok := ds.get(key); !ok {
		return courses.NoActiveCodesForEmail
	}
	delete(ds.cache, key)
	return nil
}

func (ds *DataStorage) Add(key string, val int) {
	ds.Lock()
	ds.add(key, val)
	ds.Unlock()
	go ds.cleaner(key)
	logrus.Infof("storage state: %v", ds.cache)
}

func (ds *DataStorage) Get(key string) (int, bool) {
	ds.Lock()
	defer ds.Unlock()
	if val, ok := ds.get(key); ok {
		return val, true
	}
	return 0, false
}

func (ds *DataStorage) Delete(key string) error {
	ds.Lock()
	err := ds.delete(key)
	ds.Unlock()
	return err
}
