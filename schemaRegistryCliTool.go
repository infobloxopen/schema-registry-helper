package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/infobloxopen/schema-registry-helper/confluent"
)

func validateFlags(schemaRegistryUrlPtr, schemaTypePtr, schemaDirectoryPtr *string) (string, confluent.SchemaType, string){
	if *schemaRegistryUrlPtr == "" || *schemaDirectoryPtr == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	var schemaType confluent.SchemaType
	if strings.EqualFold(*schemaTypePtr, "json") {
		schemaType = confluent.Json
	} else if strings.EqualFold(*schemaTypePtr, "protobuf") {
		schemaType = confluent.Protobuf
	} else if strings.EqualFold(*schemaTypePtr, "avro") {
		schemaType = confluent.Avro
	} else {
		flag.PrintDefaults()
		os.Exit(1)
	}
	return *schemaRegistryUrlPtr, schemaType, *schemaDirectoryPtr
}

func parseNamespaces(schemaDirectory string) ([]string){
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

func main() {
	schemaRegistryUrlPtr := flag.String("url", "", "URL of the Schema Registry (Required)")
	schemaTypePtr := flag.String("type", "", "Schema Type {json|protobuf|avro} (Required)")
	inputSchemaPtr := flag.String("inputschema", "", "Either the directory containing the schema files, or a specific schema file. " +
		"If a directory is given, tool will automatically import all schema files within subdirectories.")
	flag.Parse()

	schemaRegistry, schemaType, inputSchema := validateFlags(schemaRegistryUrlPtr, schemaTypePtr, inputSchemaPtr)
	schemaRegistryClient := confluent.CreateSchemaRegistryClient(schemaRegistry)
	fi, err := os.Stat(inputSchema)
	if err != nil {
		fmt.Printf("Error reading the input schema: %v\r\n", err)
		os.Exit(1)
	}
	if fi.Mode().IsDir() {
		fmt.Printf("Reading input as a directory. Automatically exporting all schemas in this directory tree: <%v>\r\n", inputSchema)
		namespaces := parseNamespaces(inputSchema)
		for _, n := range namespaces {
			namespaceDirectory := inputSchema + "/" + n
			fmt.Printf("Exporting schemas in directory %v...\r\n", namespaceDirectory)
			files, err := ioutil.ReadDir(namespaceDirectory)
			if err != nil {
				fmt.Printf("Error reading input directory %v, skipping...\r\n", namespaceDirectory)
				break
			}
			for _, f := range files {
				filePath := namespaceDirectory + "/" + f.Name()
				topic := n + "-" + strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
				exportSchema(filePath, topic, schemaType, *schemaRegistryClient)
			}
		}
	} else {
		fmt.Printf("Reading input as a single file. Exporting just this file: <%v>\r\n", inputSchema)
		filePathParts := strings.Split(inputSchema, "/")
		if len(filePathParts) < 2 {
			fmt.Printf("Please include the file's parent directory as part of the input argument.\r\n")
			os.Exit(1)
		}
		namespace := filePathParts[len(filePathParts)-2]
		fileName := filePathParts[len(filePathParts)-1]
		filePrefix := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		exportSchema(inputSchema, namespace + "-" + filePrefix, schemaType, *schemaRegistryClient)
	}
}

func exportSchema(filePath, topic string, schemaType confluent.SchemaType, src confluent.SchemaRegistryClient) {
	schemaBytes, _ := ioutil.ReadFile(filePath)
	schema, err := src.CheckSchema(topic, string(schemaBytes), schemaType, false)
	if err != nil && !strings.Contains(err.Error(),confluent.ErrNotFound) {
		panic(fmt.Sprintf("  Error checking the schema %v: %s", filePath, err))
	} else if err != nil {
		schema, err := src.CreateSchema(topic, string(schemaBytes), schemaType, false)
		if err != nil {
			panic(fmt.Sprintf("  Error creating the schema: %v: %s", filePath, err))
		}
		fmt.Printf("  New schema successfully created for topic %v - Schema version %v.\r\n", topic, schema.Version())
	} else {
		fmt.Printf("  Schema already exists for topic %v - Schema version %v.\r\n", topic, schema.Version)
	}
}
//TODO (not in this tool):
// for each .proto file in input argument:
// grab namespace from go_package using bash script; e.g. "pb"
// create schema/pb directory
// protoc --jsonschema_out=schema/pb blahblah

