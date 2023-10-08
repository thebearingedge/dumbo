module github.com/thebearingedge/dumbo/examples/conduit

go 1.21.1

require (
	github.com/go-faker/faker/v4 v4.2.0
	github.com/lib/pq v1.10.9
	github.com/stretchr/testify v1.8.4
	github.com/thebearingedge/dumbo v0.0.0-00010101000000-000000000000
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/thebearingedge/dumbo => ../../
