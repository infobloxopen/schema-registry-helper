# schema-registry-helper
Simple library for interacting with the confluent schema registry API. Heavily borrowed from https://github.com/riferrei/srclient.

## Input Flags
- -url
  - This is the URL of the schema registry that will be storing the user's schema.
- -type
  - The type of schema that will be stored. Options are json, protobuf, or avro
- -inputschema
  - This is the path of the actual schema(s) that will be stored in the schema registry.
  - If a directory is given, the directory must contain subdirectories which store the schema files. Each subdirectory represents a different namespace.
    - Example: `--inputschema=schemas`. schemas contains two directories, `pb/` and `service/` which each contain schema files to be exported.
  - If a single schema file is given, the path must include a directory which will represent the namespace of that schema.
  - The topic for a given schema is created from the directory that it is contained in and the name of the schema file itself.
    - Example: `schemas/service/ChannelMessage.jsonschema` will register a schema with the topic `service-ChannelMessage`.
    - If `--inputschema` is just a file name with no directory structure, no schema will be exported.
    - If `--inputschema` is a directory which contains no subdirectories (or the subdirectories contain no schema files), no schema will be exported.
    - If `--inputschema` is a directory which contains subdirectories, **all** files in those subdirectories will attempt to be exported to the schema registry.
    
## Program Output
When the library exports a schema, a few things happen.
- First, an API call will be made to see if that schema already exists in the schema registry with the given topic.
  - If it does, the program will output the version of the schema that exists in the schema registry.
  - If it doesn't, the program will attempt to register the schema.
    - If the schema is registered successfully, the program will output the version of the schema.
    - If there is an error, the program will exit.
  - If an error occurs while checking the schema, no schema will be created and the program will exit.
