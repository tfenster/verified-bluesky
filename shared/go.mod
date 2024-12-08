module github.com/shared

go 1.20

replace github.com/s => ../shared

require (
	github.com/fermyon/spin/sdk/go/v2 v2.2.0
	github.com/julienschmidt/httprouter v1.3.0 // indirect
)
