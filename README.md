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
    
## GRPC functions - Exporting to Schema Registry (package schema_registry_helper)
ExportSchema() is a function which takes an input schema and adds it to a schema registry. 
First, the function will check to see if that exact schema is already registered. 
If it is, the function will return with that schema's version.
If it is not, the schema will be added to the schema registry, and then that schema version will be returned.
