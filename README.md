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
- -group
  - This is the group that is used in the CRD file (example: notifications.infoblox.com)
- -makecrd
  - Boolean - use this if you want to generate a new CRD file in your repo. Default is FALSE, as the jsonschema CRD for CUD eventing is declared in the CR controller in the atlas.eventing.cr.controller repo.
- -omit
  - Comma-separated list of strings. Any types that start with the given strings will not have CRs created for them. Example: "read,list" will not create any CRs for message types that start with "Read" or "List"
- -crnamespace
  - Option to use a different namespace for the CRs, if {{ .Release.Namespace }} is not desired
    
## Integrating command line tool into a Makefile
The command line tool can be integrated into a Makefile by adding lines such as the last line in the following example. This will automatically translate existing protobuf schemas to json and then create custom resource files from those json schemas. Example variable definitions are below.

NOTE: You will need to use at least `v21.3` of the `infoblox/atlas-gentool` to generage .jsonschema files from your protobuf schema

```PROTOBUF_ARGS += --jsonschema_out=prefix_schema_files_with_package:$(PROJECT_ROOT)/$(SCHEMA_DIRECTORY)
```

```.PHONY protobuf: protobuf-atlas
protobuf-atlas:
	@$(GENERATOR) \
	$(PROTOBUF_ARGS) \
	$(PROJECT_ROOT)/pkg/pb/service.proto
	@$(SCHEMA_TO_CR) -inputschema=$(SCHEMA_DIRECTORY) -outputpath=$(CR_DIRECTORY)\ # add these lines
	-group=$(GROUP) -omit=$(OMIT) -crnamespace=$(CRNAMESPACE)                      # add these lines 
```

```# configuration for schema registry creator
CR_DIRECTORY     := charts/tagging-v2/templates
SCHEMA_DIRECTORY := charts/tagging-v2/schema
GROUP            := schemaregistry.infoblox.com
SCHEMA_TO_CR     := go run vendor/github.com/infobloxopen/schema-registry-helper/schema_to_cr.go
OMIT             := read,list
CRNAMESPACE      := atlas.tagging
```

To have access to the `schema-registry-helper` package in your Makefile, you'll need to manually import it using a method such as this: https://github.com/Infoblox-CTO/atlas.tagging/blob/master/hack/tools.go

The end result of this will create custom resource .yaml files in the directory provided. These files will need to be applied as part of the deployment to fully interface with the schema registry toolkit.
