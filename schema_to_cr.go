package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type CR struct {
	Name string
	Schema string
}

const cr_skeleton = "apiVersion: \"eventing.infoblox.com/v1\"\nkind: JsonSchema\nmetadata:\n  name: {{ .Name}}\nspec:\n  schema: {{ .Schema}}  registry: \n"
const crd = "apiVersion: apiextensions.k8s.io/v1\nkind: CustomResourceDefinition\nmetadata:\n  name: jsonschemas.eventing.infoblox.com\nspec:\n  " +
	"group: eventing.infoblox.com\n  versions:\n    - name: v1\n      served: true\n      storage: true\n      schema:\n        openAPIV3Schema:\n          " +
	"type: object\n          properties:\n            spec:\n              type: object\n              properties:\n                registry:\n                  " +
	"type: string\n                schema:\n                  type: object\n  scope: Namespaced\n  names:\n    plural: jsonschemas\n    singular: jsonschema\n    " +
	"kind: JsonSchema\n    shortNames:\n      - js"

func main() {
	inputSchemaPtr := flag.String("inputschema", "", "The directory containing the schema files. tool will automatically import all schema files within subdirectories.")
	outputPathPtr := flag.String("outputpath", "", "The path to the directory where the result CRs will go (required).")
	flag.Parse()
	if *inputSchemaPtr == "" || *outputPathPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	inputSchema := *inputSchemaPtr
	outputPath := *outputPathPtr
	fi1, err := os.Stat(inputSchema)
	if err != nil {
		fmt.Printf("Error reading the input schema: %v\r\n", err)
		os.Exit(1)
	}
	fi2, err := os.Stat(outputPath)
	if !fi1.Mode().IsDir() || !fi2.Mode().IsDir() {
		fmt.Printf("Input schema and output path must both be directories.\r\n")
		os.Exit(1)
	}
	namespaces := parseNamespaces(inputSchema)
	for _, n := range namespaces {
		namespaceDirectory := inputSchema + "/" + n
		fmt.Printf("Creating CRs for schemas in directory %v...\r\n", namespaceDirectory)
		files, err := ioutil.ReadDir(namespaceDirectory)
		if err != nil {
			fmt.Printf("Error reading input directory %v, skipping...\r\n", namespaceDirectory)
			break
		}
		for _, f := range files {
			filePath := namespaceDirectory + "/" + f.Name()
			topic := n + "-" + strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
			fmt.Printf("Creating custom resource file for topic %v...\r\n", topic)
			createCR(filePath, outputPath, topic)
		}
	}
	f, err := os.Create(outputPath + "/crd.yaml")
	if err != nil {
		fmt.Printf("Error creating crd.yaml file\r\n")
		os.Exit(1)
	}
	_, err = f.WriteString(crd)
	if err != nil {
		fmt.Printf("Error writing to crd.yaml file\r\n")
	}
}

func parseNamespaces(schemaDirectory string) []string {
	files, err := ioutil.ReadDir(schemaDirectory)
	if err != nil {
		fmt.Printf("Error reading input directory: %v", err)
		os.Exit(1)
	}

	namespaces := make([]string, 0)
	for _, f := range files {
		if f.IsDir() {
			namespaces = append(namespaces, f.Name())
		}
	}
	if len(namespaces) == 0 {
		fmt.Printf("Schema directory contains no subdirectories. No schemas to update.\r\n")
		os.Exit(1)
	}
	return namespaces
}

func createCR(inputFilePath, outputDirectory, topic string) {
	inputString, err := ioutil.ReadFile(inputFilePath)
	if err != nil {
		fmt.Printf("Error reading input file %v\r\n", inputFilePath)
		os.Exit(1)
	}
	f, err := os.Create(outputDirectory + "/" + topic + ".yaml")
	if err != nil {
		fmt.Printf("Error creating output file for schema %v\r\n", topic)
		os.Exit(1)
	}
	t, err := template.New("cr").Parse(cr_skeleton)
	if err != nil {
		fmt.Printf("Error processing template for schema %v", topic)
	}
	var cr CR
	cr.Name = topic
	cr.Schema = string(inputString)
	err = t.Execute(f, cr)
}
