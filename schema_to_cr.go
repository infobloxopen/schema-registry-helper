package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type CR struct {
	Name   string
	Schema string
	Group  string
}

type CRD struct {
	Group string
}

const cr_skeleton = "apiVersion: \"{{ .Group}}/v1\"\nkind: Jsonschema\nmetadata:\n  name: {{ .Name}}\nspec:\n  name: {{ .Name}}\n  schema: >\n    {{ .Schema}}\n"
const crd_skeleton = "apiVersion: apiextensions.k8s.io/v1\nkind: CustomResourceDefinition\nmetadata:\n  name: jsonschemas.{{ .Group}}\nspec:\n  " +
	"group: {{ .Group}}\n  versions:\n    - name: v1\n      served: true\n      storage: true\n      schema:\n        openAPIV3Schema:\n          " +
	"type: object\n          properties:\n            spec:\n              type: object\n              properties:\n                " +
	"schema:\n                  type: string\n                name:\n                  type: string\n  scope: Namespaced\n  names:\n    plural: jsonschemas\n    singular: jsonschema\n    " +
	"kind: Jsonschema\n    shortNames:\n      - js\n"

func main() {
	inputSchemaPtr := flag.String("inputschema", "", "The directory containing the schema files. tool will automatically import all schema files within subdirectories (required).")
	outputPathPtr := flag.String("outputpath", "", "The path to the directory where the result CRs will go (required).")
	groupPtr := flag.String("group", "", "The string of the group for the created CR and CRD files (example: schemaregistry.infoblox.com) (required).")
	makeCrdPtr := flag.Bool("makecrd", false, "Boolean option to choose whether to generate a new CRD file (optional; default false)")
	omitPtr := flag.String("omit", "", "Option to omit creating CR entries for types starting with the given string(s). Multiple strings should be comma-separated - e.g. \"read,list\" (optional).")
	crNamespacePtr := flag.String("crnamespace", "", "Option to use a different namespace for the CRs if {{ .Release.Namespace }} is not desired")

	flag.Parse()
	if *inputSchemaPtr == "" || *outputPathPtr == "" || *groupPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	inputSchema := *inputSchemaPtr
	outputPath := *outputPathPtr
	group := *groupPtr
	omit := strings.Split(*omitPtr, ",")

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

	crOutput := createCrOutput(inputSchema, group, *crNamespacePtr, omit)
	writeFiles(crOutput, outputPath, group, *makeCrdPtr)
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

func createCrOutput(inputSchema, group, crNamespace string, omit []string) map[string]string {
	crOutput := make(map[string]string)
	namespaces := parseNamespaces(inputSchema)
	for _, n := range namespaces {
		namespaceDirectory := inputSchema + "/" + n
		fmt.Printf("Creating CRs for schemas in directory %v...\r\n", namespaceDirectory)
		files, err := ioutil.ReadDir(namespaceDirectory)
		if err != nil {
			fmt.Printf("Error reading input directory %v, skipping...\r\n", namespaceDirectory)
			break
		}
		namespaceOutput := ""
		for _, f := range files {
			skip := false
			filePath := namespaceDirectory + "/" + f.Name()
			schemaType := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
			for _, o := range omit {
				if o == "" {
					continue
				}
				if strings.HasPrefix(strings.ToLower(schemaType), strings.ToLower(o)) {
					skip = true
				}
			}
			if skip {
				continue
			}
			if crNamespace == "" {
				crNamespace = "{{ .Release.Namespace }}"
			}
			schemaName := crNamespace + "." + n + "." + schemaType
			if namespaceOutput != "" {
				namespaceOutput = namespaceOutput + "---\n"
			}
			fmt.Printf("Creating custom resource for topic %v...\r\n", schemaName)
			namespaceOutput = namespaceOutput + createCR(filePath, schemaName, group)
		}
		crOutput[n] = namespaceOutput
	}
	return crOutput
}

func createCR(inputFilePath, schemaName, group string) string {
	inputString, err := ioutil.ReadFile(inputFilePath)
	if err != nil {
		fmt.Printf("Error reading input file %v\r\n", inputFilePath)
		os.Exit(1)
	}
	t, err := template.New("cr").Parse(cr_skeleton)
	if err != nil {
		fmt.Printf("Error processing template for schema %v", schemaName)
	}
	var cr CR
	cr.Name = schemaName
	cr.Schema = string(strings.ReplaceAll(string(inputString), "\n", ""))
	cr.Group = group
	var tpl bytes.Buffer
	if err := t.Execute(&tpl, cr); err != nil {
		fmt.Printf("Error creating cr %v\r\n", schemaName)
		os.Exit(1)
	}
	return tpl.String()
}

func writeFiles(crOutput map[string]string, outputPath, group string, makeCrd bool) {
	for namespace, output := range crOutput {
		fo1, err := os.Create(outputPath + "/jsonschema-" + namespace + "-cr.yaml")
		if err != nil {
			fmt.Printf("Error creating jsonschema-%v-cr.yaml file\r\n", namespace)
			os.Exit(1)
		}
		_, err = fo1.WriteString(output)
		if err != nil {
			fmt.Printf("Error writing to jsonschema-%v-cr.yaml file\r\n", namespace)
		}
	}
	if !makeCrd {
		return
	}
	fo2, err := os.Create(outputPath + "/jsonschema-crd.yaml")
	if err != nil {
		fmt.Printf("Error creating crd.yaml file\r\n")
		os.Exit(1)
	}
	t, err := template.New("crd").Parse(crd_skeleton)
	if err != nil {
		fmt.Printf("Error processing template for crd\r\n")
	}
	var crd CRD
	crd.Group = group
	var tpl bytes.Buffer
	if err := t.Execute(&tpl, crd); err != nil {
		fmt.Printf("Error creating crd \r\n")
		os.Exit(1)
	}
	_, err = fo2.WriteString(tpl.String())
	if err != nil {
		fmt.Printf("Error writing to crd.yaml file\r\n")
	}
}
