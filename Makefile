version ?= v0.0.1
bin ?= rrt
name ?= rrt
tag = $(name):$(version)
files = $(shell find . -iname "*.go")


run:
	go run main.go version

bin/$(bin): $(files)
	time \
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
