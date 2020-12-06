.PHONY: build clean deploy gomodgen

build: gomodgen
	export GO111MODULE=on
	env GOOS=linux go build -ldflags="-s -w" -o bin/release release/main.go

clean:
	rm -rf ./bin ./vendor go.sum

deploy: clean build
	sls deploy --verbose  --aws-profile sean

gomodgen:
	chmod u+x gomod.sh
	./gomod.sh

destroy:
	serverless remove --verbose --aws-profile sean