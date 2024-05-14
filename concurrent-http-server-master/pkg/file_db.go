package pkg

import (
	"errors"
)

type FileDB map[string]string

func NewFileTable() FileDB {
	return make(FileDB)
}

func (fdb *FileDB) AddToDB(id, path string) error {
	if _, exist := (*fdb)[id]; exist {
		return errors.New("file with the given id exists")
	}
	(*fdb)[id] = path
	return nil
}

func (fdb *FileDB) GetFilePath(id string) (string, error) {
	retrievedPath, exist := (*fdb)[id]
	if !exist {
		return "", errors.New("no file with the given id")
	}
	return retrievedPath, nil
}
