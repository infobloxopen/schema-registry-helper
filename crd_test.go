package main

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

var testCRDs = []struct {
	input      CRD
	outputPath string
}{
	{
		input: CRD{
			Group: "group/v1",
		},
		outputPath: "crd.one.yaml",
	},
}

func TestCRD(t *testing.T) {
	for _, tc := range testCRDs {
		path := filepath.Join("testdata", tc.outputPath)
		bs, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		e := string(bs)
		s, err := createCRD(tc.input)
		if err != nil {
			t.Fatal(err)
		}

		if strings.TrimSpace(e) != strings.TrimSpace(s) {
			t.Errorf("got:\n%q\nwanted:\n%q", s, e)
		}
	}

}
