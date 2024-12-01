module github.com/validate_javachamps

go 1.20

require github.com/fermyon/spin/sdk/go/v2 v2.2.0

require (
	github.com/julienschmidt/httprouter v1.3.0 // indirect
	github.com/shared v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/shared => ../shared
