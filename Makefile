.PHONY: build test

build:
	cd cmd/shortener && go build -o shortener *.go

test: build
	TEMP_FILE=$$(mktemp) && \
	~/Desktop/shortenertest -test.v -test.run=^TestIteration10$$ \
	-binary-path=cmd/shortener/shortener \
	-source-path=. \
	-file-storage-path=$$TEMP_FILE \
    -database-dsn="postgres://illustrv:example@localhost:5432/praktikum?sslmode=disable"

clean:
	rm -f cmd/shortener/shortener