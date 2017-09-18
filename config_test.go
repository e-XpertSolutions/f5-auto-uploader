package main

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"
)

const validConfigFileContent = `[f5]
auth_method = "basic"
url = "http://localhost/bigip"
user = "admin"
password = "admin"
ssl_check = false
`
const invalidConfigFileContent = `{invalid}`

func createTempConfigFile(data string) (*os.File, error) {
	f, err := ioutil.TempFile(os.TempDir(), "f5-auto-uploader-test")
	if err != nil {
		return nil, errors.New("cannot create temporary configuration file")
	}
	if _, err := f.Write([]byte(data)); err != nil {
		f.Close()
		return nil, errors.New("cannot write configuration in temporary configuration file")
	}
	return f, nil
}

func TestReadConfig(t *testing.T) {
	// Setup temporary configuration files and make sure they will be deleted
	// at the end of this test.
	validFile, err := createTempConfigFile(validConfigFileContent)
	if err != nil {
		t.Fatal("setup: ", err)
	}
	defer os.Remove(validFile.Name())
	defer validFile.Close()

	invalidFile, err := createTempConfigFile(invalidConfigFileContent)
	if err != nil {
		t.Fatal("setup: ", err)
	}
	defer os.Remove(invalidFile.Name())
	defer invalidFile.Close()

	// Run subtests
	t.Run("Happy Path", func(t *testing.T) { testReadConfigHappyPath(t, validFile) })
	t.Run("Fail Open", testReadConfigFailOpen)
	t.Run("Fail Decode", func(t *testing.T) { testReadConfigFailDecode(t, invalidFile) })
}

func testReadConfigHappyPath(t *testing.T, validFile *os.File) {
	path := validFile.Name()

	cfg, err := readConfig(path)
	if err != nil {
		t.Fatalf("readConfig(%q): unexpected error %q", path, err.Error())
	}
	want := config{}
	want.F5 = f5Config{
		AuthMethod: "basic",
		URL:        "http://localhost/bigip",
		User:       "admin",
		Password:   "admin",
	}
	if got := cfg.F5.AuthMethod; got != want.F5.AuthMethod {
		t.Errorf("readConfig(%q): got auth_method %q; want %q",
			path, got, want.F5.AuthMethod)
	}
	if got := cfg.F5.URL; got != want.F5.URL {
		t.Errorf("readConfig(%q): got url %q; want %q",
			path, got, want.F5.URL)
	}
	if got := cfg.F5.User; got != want.F5.User {
		t.Errorf("readConfig(%q): got user %q; want %q",
			path, got, want.F5.User)
	}
	if got := cfg.F5.Password; got != want.F5.Password {
		t.Errorf("readConfig(%q): got password %q; want %q",
			path, got, want.F5.Password)
	}
	if got := cfg.F5.SSLCheck; got != want.F5.SSLCheck {
		t.Errorf("readConfig(%q): got password %v; want %v",
			path, got, want.F5.SSLCheck)
	}
	if got := cfg.F5.LoginProviderName; got != want.F5.LoginProviderName {
		t.Errorf("readConfig(%q): got password %q; want %q",
			path, got, want.F5.LoginProviderName)
	}
}

func testReadConfigFailOpen(t *testing.T) {
	invalidPath := "some-path-that-does-not-exist"
	_, err := readConfig(invalidPath)
	if err == nil {
		t.Fatalf("readConfig(%q): expected error, got nil", invalidPath)
	}
	wantErr := "cannot open configuration file: open some-path-that-does-not-exist: no such file or directory"
	if err.Error() != wantErr {
		t.Errorf("readConfig(%q): got error %q; want %q", invalidPath, err.Error(), wantErr)
	}
}

func testReadConfigFailDecode(t *testing.T, invalidFile *os.File) {
	path := invalidFile.Name()
	_, err := readConfig(path)
	if err == nil {
		t.Fatalf("readConfig(%q): expected error, got nil", path)
	}
	wantErr := "cannot read configuration file: Near line 0 (last key parsed ''): bare keys cannot contain '{'"
	if err.Error() != wantErr {
		t.Errorf("readConfig(%q): got error %q; want %q", path, err.Error(), wantErr)
	}
}
