package main

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

var testCRs = []struct {
	input      CR
	outputPath string
}{
	{
		input: CR{
			Name:   "spec-test",
			LName:  "metadata-name",
			Schema: "schema",
			Group:  "group",
		},
		outputPath: "one.yaml",
	},
}

func TestSkeletion(t *testing.T) {
	for _, tc := range testCRs {
		path := filepath.Join("testdata", tc.outputPath)
		bs, err := ioutil.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}

		e := string(bs)
		s, err := createCR(tc.input)
		if err != nil {
			t.Fatal(err)
		}

		if e != s {
			t.Errorf("got:\n%q\nwanted:\n%q", s, e)
		}
	}
}
