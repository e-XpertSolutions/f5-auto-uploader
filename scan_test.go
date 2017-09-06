package main

import (
	"testing"

	"github.com/e-XpertSolutions/f5-rest-client/f5"
)

func TestScanDir(t *testing.T) {
	f5Client, _ := f5.NewBasicClient("https://192.168.10.40", "admin", "admin")
	f5Client.DisableCertCheck()
	if err := scanDir("/tmp/test", f5Client); err != nil {
		t.Error(err)
	}
}
