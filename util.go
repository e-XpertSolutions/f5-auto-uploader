package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/e-XpertSolutions/f5-rest-client/f5"
	"github.com/e-XpertSolutions/f5-rest-client/f5/ltm"
	"github.com/e-XpertSolutions/f5-rest-client/f5/sys"
)

func isExcluded(name string, excl []string) bool {
	for _, pattern := range excl {
		matched, err := filepath.Match(pattern, name)
		if err == nil && matched {
			fmt.Printf("%q is excluded by pattern %q\n", name, pattern)
			return true
		} else if err != nil {
			fmt.Println("err = ", err)
		}
		fmt.Printf("%q is not excluded by pattern %q\n", name, pattern)
	}
	return false
}

func uploadNewFile(tx *f5.Client, name, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot read file %q: %v", path, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("cannot stat file %q: %v", path, err)
	}

	sysClient := sys.New(tx)
	if err := sysClient.FileIFile().CreateFromFile(name, f, info.Size()); err != nil {
		return fmt.Errorf("an error occured while uploading %q: %v", path, err)
	}

	ltmClient := ltm.New(tx)
	if err := ltmClient.IFile().Create(name, name); err != nil {
		return fmt.Errorf("cannot create file %q in ltm ifiles: %v", path, err)
	}

	return nil
}

func uploadExistingFile(tx *f5.Client, name, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot read file %q: %v", path, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("cannot stat file %q: %v", path, err)
	}

	sysClient := sys.New(tx)
	if err := sysClient.FileIFile().EditFromFile(name, f, info.Size()); err != nil {
		return fmt.Errorf("an error occured while re-uploading %q: %v", path, err)
	}

	ltmClient := ltm.New(tx)
	if err := ltmClient.IFile().Edit(name, name); err != nil {
		return fmt.Errorf("cannot update file %q in ltm ifiles: %v", path, err)
	}

	return nil
}

func deleteFile(tx *f5.Client, name string) error {
	ltmClient := ltm.New(tx)
	if err := ltmClient.IFile().Delete(name); err != nil {
		return fmt.Errorf("cannot delete ltm ifile %q: %v", name, err)
	}

	sysClient := sys.New(tx)

	if err := sysClient.FileIFile().Delete(name); err != nil {
		return fmt.Errorf("cannot delete ifile %q: %v", name, err)
	}

	if err := sysClient.FileIFile().Delete(name); err != nil {
		return fmt.Errorf("cannot delete ifile %q from disk: %v", name, err)
	}

	return nil
}

func isSameRevision(tx *f5.Client, name, path string) (bool, error) {
	sysClient := sys.New(tx)
	ifile, err := sysClient.FileIFile().Get(name)
	if err != nil {
		return false, fmt.Errorf("cannot get ifile meta for %q: %v", name, err)
	}

	f, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("cannot open file %q: %v", path, err)
	}
	defer f.Close()

	algo, _, checksum := splitChecksum(ifile.Checksum)

	var h hash.Hash
	switch strings.ToLower(algo) {
	case "sha1":
		h = sha1.New()
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	case "md5":
		h = md5.New()
	default:
		return false, fmt.Errorf("unsupported algo %q for file %q", algo, path)
	}
	if _, err := io.Copy(h, f); err != nil {
		return false, fmt.Errorf("cannot write file %q into hash function of type %q: %v", path, algo, err)
	}
	expectedChecksum := hex.EncodeToString(h.Sum(nil)[:])

	return checksum == expectedChecksum, nil
}

func splitChecksum(ifileChecksum string) (algo, opts, checksum string) {
	elmts := strings.Split(ifileChecksum, ":")
	switch l := len(elmts); l {
	case 0:
		return
	case 1:
		checksum = elmts[0]
	case 2:
		algo = elmts[0]
		checksum = elmts[1]
	default: // 3 or more
		algo = elmts[0]
		opts = strings.Join(elmts[1:l-1], ":")
		checksum = elmts[l-1]
	}
	return
}
