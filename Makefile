.PHONY: build run go-test py-test

build:
    go mod tidy
    go build -o gui ./cmd/gui

run: build
    ./gui -addr=127.0.0.1:8080 -static=./webui

go-test:
    go test ./... -v

py-test:
    python -m pip install --upgrade pip
    pip install -r python/generator/requirements.txt
    pytest -q python/generator
