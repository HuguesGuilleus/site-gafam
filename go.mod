module frontend-gafam

go 1.24.1

require (
	github.com/stretchr/testify v1.10.0
	sniffle v0.0.0
)

require (
	github.com/BurntSushi/toml v1.4.0
	github.com/tdewolff/minify/v2 v2.22.4 // indirect
	github.com/tdewolff/parse v2.3.4+incompatible // indirect
	github.com/tdewolff/parse/v2 v2.7.21 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/tdewolff/minify v2.3.6+incompatible
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace sniffle => ../sniffle/
