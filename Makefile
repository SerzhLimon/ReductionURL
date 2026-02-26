.PHONY: build test

build:
	cd cmd/shortener && go build -o shortener *.go

test: build
	~/Desktop/shortenertest -test.v -test.run=^TestIteration13$$ \
		-binary-path=cmd/shortener/shortener \
		-source-path=. \
		-database-dsn="postgres://illustrv:example@localhost:5432/reductionurl?sslmode=disable"

clean:
	rm -f cmd/shortener/shortener