
test:
	go test -v ./...

.PHONY examples:
examples:
	go run schema_to_cr.go --inputschema=example/schema --outputpath=example/cr --group=schemaregistry.infoblox.com

