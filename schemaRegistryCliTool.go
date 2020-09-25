package schema_registry_helper

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
}

func main() {
	schemaRegistryUrlPtr := flag.String("url", "", "URL of the Schema Registry (Required)")
	schemaTypePtr := flag.String("type", "", "Schema Type {json|protobuf|avro} (Required)")
	schemaDirectoryPtr := flag.String("directory", "", "Directory containing the schema files (Required)")
	singleSchemaPtr := flag.String("single", "", "Export only the chosen schema file. " +
									"If left blank, all schemas in the directory will be exported (Optional)")
	flag.Parse()

	schemaRegistry, schemaType, schemaDirectory := validateFlags(schemaRegistryUrlPtr, schemaTypePtr, schemaDirectoryPtr)
	schemaRegistryClient := confluent.CreateSchemaRegistryClient(schemaRegistry)
	namespaces := parseNamespaces(schemaDirectory)

	if *singleSchemaPtr == "" {
		for _, n := range namespaces {
			namespaceDirectory := schemaDirectory + "/" + n
			fmt.Printf("Exporting schemas in directory %v", namespaceDirectory)
			files, err := ioutil.ReadDir(namespaceDirectory)
			if err != nil {
				fmt.Printf("Error reading input directory %v, skipping...", namespaceDirectory)
				break
			}
			for _, f := range files {
				exportSchema(f.Name(), namespaceDirectory, schemaType, *schemaRegistryClient)
			}
		}
	} else {
		fmt.Printf("Feature under construction")
	}
}

func exportSchema(fileName, namespaceDirectory string, schemaType confluent.SchemaType, src confluent.SchemaRegistryClient) {
	filePath := namespaceDirectory + "/" + fileName
	topic := n + "-" + strings.TrimSuffix(fileName, filepath.Ext(fileName))
	schemaBytes, _ := ioutil.ReadFile(filePath)
	_, err := src.CheckSchema(topic, string(schemaBytes), schemaType, false)
	if err.Error() != confluent.ErrNotFound {
		panic(fmt.Sprintf("Error checking the schema %v: %s", filePath, err))
	} else {
		_, err := src.CreateSchema(topic, string(schemaBytes), schemaType, false)
		if err != nil {
			panic(fmt.Sprintf("Error creating the schema: %v: %s", filePath, err))
		}
	}
}
//TODO (not in this tool):
// for each .proto file in input argument:
// grab namespace from go_package using bash script; e.g. "pb"
// create schema/pb directory
// protoc --jsonschema_out=schema/pb blahblah

