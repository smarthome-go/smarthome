# Makefile for the smarthome-go/smarthome project
appname := smarthome
workingdir := smarthome
sources := $(wildcard *.go)
# Do not edit manually, use the `version` target to change the
# version programmatically in all places
version := 0.9.0-beta

build = CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build -ldflags "-s -w" -v -o $(appname) $(4)
# TODO: eliminate usage of workingdir
tar = mkdir -p build && cd ../ && tar -cvzf ./$(appname)_$(1)_$(2).tar.gz $(workingdir)/$(appname) $(workingdir)/web/dist $(workingdir)/web/html $(workingdir)/resources && mv $(appname)_$(1)_$(2).tar.gz $(workingdir)/build

.PHONY: all linux

all:	linux

# Setup dependencies for Go and NPM
setup:
	cd web && npm i && npm run prepare

deps:
	go mod tidy
	go get -u -v
	cd web && npm outdated; npm update && npm run prepare


# Lints most of the source code
# Used before a release
lint:
	golangci-lint run
	go vet
	typos
	cd web && npm run lint


# Run a normal integration and unit test procedure
test:
	mkdir -p web/dist/html
	touch web/dist/html/testing.html
	# Prevents server panic

	go test -v -p 1 ./... --timeout=10000s
	# Tests should be run one after another due to deletion of the database at every test start
	rm -rf web/dist/html/testing.html

# Runs the integration and unit tests and outputs the test coverage as `coverage.html`
vtest:
	mkdir -p web/dist/html
	touch web/dist/html/testing.html
	# Prevents server panic

	go test -v -p 1 ./... -coverprofile=coverage.out
	go tool cover --html=coverage.out -o coverage.html
	rm -rf web/dist/html/testing.html

# Update the current version in all locations
version:
	python3 update_version.py
	cd web && npm i

# Prepares everything for a version-release
# In order to publish the release to official registries
# run `make gh-release` and `make docker-push`
release-slim: cleanall lint build docker
release: cleanall lint test build docker

# Publishes the local release to Github releases
gh-release:
	gh release create v$(version) ./build/*.tar.gz -F ./docs/CHANGELOG.md -t 'Smarthome v$(version)'  --prerelease

# Starts the vite development server
vite-dev:
	cd web && npm run dev

# Start the project to mock production
# Generates the frontend web output first
# then runs the go backend
run: web
	go run -v -race .

# Similar to `run` but starts a development database first
# which is required for the backend to run
# docker and docker-compose are required for this target
run-full: web mysql
	go run -v -race .

# Removes most of the intermediate cache and build files
clean: cleanweb
	rm -rf bin
	rm -rf log
	rm -rf docker/container/cache
	rm -rf coverage.out
	rm -rf coverage.html

# Removes the output folder of the web interface
cleanweb:
	rm -rf web/dist

# Removes all intermediate cache and build files
# Also removes the `build` directory, which contains release-ready tarballs
cleanall: clean
	rm -rf build
	rm -f smarthome

# Builds the Go backend and the frontend web interface
# Produces the `build` directory, which contains release-ready tarballs
build: setup web all linux clean

# Prepares the local filesystem for a Docker build
# Mostly copies precompiled dependencies to the docker cache directory
# compiles the Go backend to an AMD64 binary and copies the
# pre generated web output to a docker cache directory
docker-prepare: web build

	mkdir -p docker/container/cache/web
	cp -r resources docker/container/cache/
	cp -r web/dist docker/container/cache/web/

	$(call build,linux,amd64, -ldflags '-s -w -extldflags "-fno-PIC -static"' -buildmode pie -tags 'osusergo netgo static_build')
	cp smarthome docker/container/smarthome_amd64

	$(call build,linux,arm,)
	cp smarthome docker/container/smarthome_arm

	$(call build,linux,arm64,)
	cp smarthome docker/container/smarthome_arm64

	$(info "docker-prepare: build context has been written to ./docker/cache")

# Is used after `release` in order to publish the built
# Docker image to Docker-Hub
docker-push:
	docker push mikmuellerdev/smarthome:$(version)-arm
	docker push mikmuellerdev/smarthome:latest-arm
	docker push mikmuellerdev/smarthome:$(version)-arm64
	docker push mikmuellerdev/smarthome:latest-arm64
	docker push mikmuellerdev/smarthome:$(version)-amd64
	docker push mikmuellerdev/smarthome:latest-amd64

	$(info "docker-push: successfully pushed to remote repository")

# Builds the Docker image using the pre compiled
# and setup build cache
docker: cleanall web docker-prepare
	docker buildx create --use

	sudo docker buildx build \
		-t mikmuellerdev/smarthome:$(version)-arm \
		-t mikmuellerdev/smarthome:latest-arm \
		--platform=linux/arm \
		--load \
		-f ./docker/container/Dockerfile \
		./docker/container/

	sudo docker buildx build \
		-t mikmuellerdev/smarthome:$(version)-arm64 \
		-t mikmuellerdev/smarthome:latest-arm64 \
		--platform=linux/arm64 \
		--load \
		-f ./docker/container/Dockerfile \
		./docker/container/

	docker buildx build \
		-t mikmuellerdev/smarthome:$(version)-amd64 \
		-t mikmuellerdev/smarthome:latest-amd64 \
		--platform=linux/amd64 \
		--load \
		-f ./docker/container/Dockerfile \
		./docker/container/

# Generates the output files for the frontend web interface
web: cleanweb
	cd web && npm run build

# Build architectures
linux: build/linux_arm.tar.gz build/linux_arm64.tar.gz build/linux_386.tar.gz build/linux_amd64.tar.gz

build/linux_386.tar.gz: $(sources)
	$(call build,linux,386,)
	$(call tar,linux,386)

build/linux_amd64.tar.gz: $(sources)
	$(call build,linux,amd64, -ldflags '-s -w -extldflags "-fno-PIC -static"' -buildmode pie -tags 'osusergo netgo static_build')
	$(call tar,linux,amd64)

build/linux_arm.tar.gz: $(sources)
	$(call build,linux,arm,)
	$(call tar,linux,arm)

build/linux_arm64.tar.gz: $(sources)
	$(call build,linux,arm64,)
	$(call tar,linux,arm64)

