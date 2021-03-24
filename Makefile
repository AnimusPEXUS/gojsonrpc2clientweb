all: get

get:
		GOOS=js GOARCH=wasm go get -u -v "./..."
		GOOS=js GOARCH=wasm go mod tidy
