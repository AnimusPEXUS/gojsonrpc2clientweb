export GONOPROXY=github.com/AnimusPEXUS/*
export GOOS=js
export GOARCH=wasm

all: get

get:
		go get -u -v "./..."
		go mod tidy
