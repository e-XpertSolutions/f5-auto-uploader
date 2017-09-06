package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/e-XpertSolutions/f5-rest-client/f5"
	"github.com/e-XpertSolutions/f5-rest-client/f5/ltm"
	"github.com/e-XpertSolutions/f5-rest-client/f5/sys"
)

func scanDir(dir string, f5Client *f5.Client) error {
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
		if fi.IsDir() {
			continue
		}
		if _, ok := existingFiles[fi.Name()]; !ok {
			path := filepath.Join(dir, fi.Name())
			filesize := fi.Size()
			if filesize == 0 {
				continue
			}
			if err := uploadNewFile(tx, fi.Name(), path, filesize); err != nil {
				return err
			}
			totalChanges++
		}
	}
	if totalChanges > 0 {
		if err := tx.Commit(); err != nil {
			return errors.New("cannot commit transaction: " + err.Error())
		}
	}
	return nil
}

func uploadNewFile(tx *f5.Client, name, path string, filesize int64) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot read file %q: %v", path, err)
	}
	defer f.Close()

	sysClient := sys.New(tx)
	if err := sysClient.FileIFile().CreateFromFile(name, f, filesize); err != nil {
		return fmt.Errorf("an error occured while uploading %q: %v", path, err)
	}

	ltmClient := ltm.New(tx)
	if err := ltmClient.IFile().Create(name, name); err != nil {
		return fmt.Errorf("cannot create file %q in ltm ifiles: %v", path, err)
	}

	return nil
}

func uploadExistingFile(tx *f5.Client, name, path string, filesize int64) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot read file %q: %v", path, err)
	}
	defer f.Close()

	sysClient := sys.New(tx)
	if err := sysClient.FileIFile().EditFromFile(name, f, filesize); err != nil {
		return fmt.Errorf("an error occured while re-uploading %q: %v", path, err)
	}

	ltmClient := ltm.New(tx)
	if err := ltmClient.IFile().Edit(name, name); err != nil {
		return fmt.Errorf("cannot update file %q in ltm ifiles: %v", path, err)
	}

	return nil
}
