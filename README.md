# image-gen-gui

Go frontend + Python generator のローカル GUI アプリ。

## Quickstart

1. Build
   make build

2. Run
   make run

3. Open
   http://127.0.0.1:8080

4. Upload images and watch generation progress.

## Tests
- Go: `make go-test`
- Python: `make py-test`

## Extensibility
- Replace `python/generator/model_stub.py` with Diffusers/Torch implementation.
- Add static handler in Go to serve generated images securely.
- Replace stdin/stdout IPC with gRPC for production.
