# The module needs Go 1.24 or later, which builds WebAssembly exports.
ID      := $(shell sed -n 's/^id: //p' plugin.yaml)
VERSION := $(shell sed -n 's/^version: //p' plugin.yaml)
ZIP     := dist/$(ID)-$(VERSION).zip

# Windows 本机没有 zip 时，用 tar 或手工打包即可；plugin.wasm 才是构建产物。
.PHONY: build test zip clean

build:
	GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -trimpath -ldflags '-s -w' -o plugin.wasm .

test:
	go test ./...

# What a site installs: the package without the module's source.
zip: build
	mkdir -p dist
	rm -f $(ZIP)
	zip -qr $(ZIP) plugin.yaml plugin.wasm assets i18n README.md LICENSE
	@echo $(ZIP)

clean:
	rm -rf plugin.wasm dist
