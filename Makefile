.PHONY: build test

build:
	cd cmd/shortener && go build -o shortener *.go

test: build
	TEMP_FILE=$$(mktemp) && \
	~/Desktop/shortenertest -test.v -test.run=^TestIteration9$$ \
	-binary-path=cmd/shortener/shortener \
	-source-path=. \
	-file-storage-path=$$TEMP_FILE

clean:
	rm -f cmd/shortener/shortener