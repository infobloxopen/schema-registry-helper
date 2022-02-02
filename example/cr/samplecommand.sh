# output of this file can be found at example/cr/jsonschema-pb-cr.yaml
go run schema_to_cr.go --inputschema=/Users/broc/go/src/github.com/Infoblox-CTO/atlas.tagging/charts/tagging-v2/schema/ --outputpath=example/cr --group=schemaregistry.infoblox.com --omit=read,list --crnamespace=atlas.tagging --commonname=TestCommon --fullname=test.full.name
