package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/e-XpertSolutions/f5-rest-client/f5"
	"github.com/e-XpertSolutions/f5-rest-client/f5/ltm"
)

func scanDir(dir string, excl []string, f5Client *f5.Client) error {
	ltmClient := ltm.New(f5Client)
	ifilesList, err := ltmClient.IFile().ListAll()
	if err != nil {
		return errors.New("cannot retrieve list of existing ifiles: " + err.Error())
	}
	existingFiles := make(map[string]struct{})
	for _, item := range ifilesList.Items {
		existingFiles[filepath.Base(item.FileName)] = struct{}{}
	}
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("cannot read content of directory %q: %v", dir, err)
	}
	tx, err := f5Client.Begin()
	if err != nil {
		return errors.New("cannot start f5 transaction: " + err.Error())
	}
	var totalChanges int
	for _, fi := range fis {
		if fi.IsDir() || !fi.Mode().IsRegular() || isExcluded(fi.Name(), excl) {
			continue
		}
		path := filepath.Join(dir, fi.Name())
		filesize := fi.Size()
		if filesize == 0 {
			continue
		}
		if _, ok := existingFiles[fi.Name()]; !ok {
			if err := uploadNewFile(tx, fi.Name(), path); err != nil {
				return err
			}
		} else {
			same, err := isSameRevision(tx, fi.Name(), path)
			if err != nil {
				return err
			}
			if same {
				continue
			}
			if err := uploadExistingFile(tx, fi.Name(), path); err != nil {
				return err
			}
		}
		totalChanges++
	}
	if totalChanges > 0 {
		if err := tx.Commit(); err != nil {
			return errors.New("cannot commit transaction: " + err.Error())
		}
	}
	return nil
}
