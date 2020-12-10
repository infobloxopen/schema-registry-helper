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

	"github.com/Masterminds/sprig"
)

type CR struct {
	Name   string
	LName  string
	Schema string
	Group  string
}

type CRD struct {
	Group string
}

const cr_skeleton = `apiVersion: {{ .Group }}
kind: Jsonschema
metadata:
  name: {{ .LName }}
spec:
  name: {{ .Name }}
  schema: |
  {{- .Schema | nindent 4 }}
`

const crd_skeleton = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: jsonschemas.{{ .Group}}
spec:
  group: {{ .Group}}
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                schema:
                  type: string
                name:
                  type: string
  scope: Namespaced
  names:
    plural: jsonschemas
    singular: jsonschema
    kind: Jsonschema
    shortNames:
      - js
`

func main() {
	inputSchemaPtr := flag.String("inputschema", "", "The directory containing the schema files. tool will automatically import all schema files within subdirectories (required).")
	outputPathPtr := flag.String("outputpath", "", "The path to the directory where the result CRs will go (required).")
	groupPtr := flag.String("group", "", "The string of the group for the created CR and CRD files (example: schemaregistry.infoblox.com) (required).")
	makeCrdPtr := flag.Bool("makecrd", false, "Boolean option to choose whether to generate a new CRD file (optional; default false)")
	omitPtr := flag.String("omit", "", "Option to omit creating CR entries for types starting with the given string(s). Multiple strings should be comma-separated - e.g. \"read,list\" (optional).")
	crNamespacePtr := flag.String("crnamespace", "", "Option to use a different namespace for the CRs if {{ .Release.Namespace }} is not desired")
	skipGuardPtr := flag.Bool("skipguard", false, "Boolean option to choose whether to skip the guard condition in the CR and CRD files (optional; default false)")

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
	writeFiles(crOutput, outputPath, group, *makeCrdPtr, *skipGuardPtr)
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
			schemaName := crNamespace + "-" + n + "-" + schemaType
			if namespaceOutput != "" {
				namespaceOutput = namespaceOutput + "---\n"
			}
			fmt.Printf("Creating custom resource for topic %v...\r\n", schemaName)
			text, err := strCreateCR(filePath, schemaName, group)
			if err != nil {
				panic(err.Error())
			}
			namespaceOutput = namespaceOutput + text
		}
		crOutput[n] = namespaceOutput
	}
	return crOutput
}

func strCreateCR(inputFilePath, schemaName, group string) (string, error) {
	inputString, err := ioutil.ReadFile(inputFilePath)
	if err != nil {
		fmt.Printf("Error reading input file %v\r\n", inputFilePath)
		os.Exit(1)
	}
	var cr CR
	cr.LName = strings.ToLower(schemaName)
	cr.Name = schemaName
	cr.Schema = strings.TrimRight(string(strings.ReplaceAll(string(inputString), "\n", "\n    ")), " ")
	cr.Group = group
	return createCR(cr)
}

func createCR(cr CR) (string, error) {
	t, err := template.New("cr").Funcs(sprig.TxtFuncMap()).Parse(cr_skeleton)
	if err != nil {
		fmt.Printf("Error processing template for schema %v", cr.Name)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, cr); err != nil {
		panic(fmt.Sprintf("Error creating cr %s", cr.Name))
	}
	return buf.String(), nil
}

func writeFiles(crOutput map[string]string, outputPath, group string, makeCrd, skipGuard bool) {
	for namespace, output := range crOutput {
		fo1, err := os.Create(outputPath + "/jsonschema-" + namespace + "-cr.yaml")
		if err != nil {
			fmt.Printf("Error creating jsonschema-%v-cr.yaml file\r\n", namespace)
			os.Exit(1)
		}
		if !skipGuard {
			output = "{{- if .Values.schemaregistry.enabled }}\r\n" + output + "{{- end }}\r\n"
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

	var crd CRD
	crd.Group = group

	s, err := createCRD(crd)
	if err != nil {
		panic(err.Error())
	}
	if !skipGuard {
		s = "{{- if .Values.schemaregistry.enabled }}\r\n" + s + "{{- end }}\r\n"
	}
	_, err = fo2.WriteString(s)
	if err != nil {
		fmt.Printf("Error writing to crd.yaml file\r\n")
	}
}

func createCRD(input CRD) (string, error) {
	t, err := template.New("crd").Parse(crd_skeleton)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, input); err != nil {
		return "", err
	}
	return buf.String(), nil
}
