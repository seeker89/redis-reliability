version ?= v0.0.1
bin ?= rr
name ?= rr
namespace ?= seeker89
tag = $(name):$(version)
files = $(shell find . -iname "*.go")


run:
	go run main.go version

bin/$(bin): $(files)
	CGO_ENABLED=0 \
	go build \
		-ldflags "-extldflags=-static" \
		-ldflags "-X 'main.Version=${version}' -X 'main.Build=`date`'" \
		-o ./bin/${bin} \
		main.go

clean:
	rm -f ./bin/$(bin)

image:
	docker build -t $(namespace)$(tag) --target simple -f ./Dockerfile .


.PHONY: run clean image
