package main

import (
	"io/ioutil"
	"path/filepath"
	"strings"
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
			Group:  "group/v1",
		},
		outputPath: "one.yaml",
	},
	{
		input: CR{
			Name:  "spec-test-2",
			LName: "metadata-name-2",
			Schema: `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "properties": {
        "id": {
            "$ref": "gorm.types.UUIDValue",
            "additionalProperties": true,
            "type": "object"
        }
    }
}`,
			Group: "group/v1",
		},
		outputPath: "two.yaml",
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

		if strings.TrimSpace(e) != strings.TrimSpace(s) {
			t.Errorf("got:\n%q\nwanted:\n%q", s, e)
		}
	}
}
