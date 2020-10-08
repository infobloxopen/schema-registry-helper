# schema-registry-helper
Simple library for interacting with the confluent schema registry API. Heavily borrowed from https://github.com/riferrei/srclient.

## Command Line Tool - Input Flags (schema_to_cr.go)
- -inputschema
  - This is the path of the actual schema(s) that will be converted into custom resource files.
  - The directory must contain subdirectories which store the schema files. Each subdirectory represents a different namespace.
    - Example: `--inputschema=schemas`. schemas contains two directories, `pb/` and `service/` which each contain schema files to be converted.
  - The topic for a given schema is created from the directory that it is contained in and the name of the schema file itself.
    - Example: `schemas/service/ChannelMessage.jsonschema` will register a schema with the topic `service-ChannelMessage`.
    - If `--inputschema` is just a file name with no directory structure, no schema will be converted.
    - If `--inputschema` is a directory which contains no subdirectories (or the subdirectories contain no schema files), no schema will be converted.
    - If `--inputschema` is a directory which contains subdirectories, **all** files in those subdirectories will attempt to be converted.
- -outputpath
  - This is the path to the directory which will contain the output custom resource files, as well as the custom resource definition file.
  
## Integrating command line tool into a Makefile
The command line tool can be integrated into a Makefile by adding lines such as this. This will automatically translate existing protobuf schemas to json and then create custom resource files (and a custom resource definition file) from those json schemas (Make sure that `CR_DIRECTORY` and `SCHEMA_DIRECTORY` are defined somewhere in your Makefile as well)

```.PHONY schema-clean: schema
schema:
  @GOSUMDB=off go get github.com/chrusty/protoc-gen-jsonschema/cmd/protoc-gen-jsonschema
  @GOSUMDB=off go get github.com/infobloxopen/schema-registry-helper
  @mkdir -p $(CR_DIRECTORY) $(SCHEMA_DIRECTORY)
  @protoc --jsonschema_out=prefix_schema_files_with_package:$(SCHEMA_DIRECTORY) -I=vendor -I=pkg/pb pkg/pb/*.proto
  @schema-registry-helper -inputschema=schema -outputpath=$(CR_DIRECTORY)
```

The end result of this will create jsonschema-cr.yaml and jsonschema-crd.yaml files in the directory provided. These files will need to be applied as part of the deployment to fully interface with the schema registry toolkit.

## GRPC functions - Exporting to Schema Registry (package schema_registry_helper)
ExportSchema() is a function which takes an input schema and adds it to a schema registry. 
First, the function will check to see if that exact schema is already registered. 
If it is, the function will return with that schema's version.
If it is not, the schema will be added to the schema registry, and then that schema version will be returned.
