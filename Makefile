appname := smarthome
workingdir := smarthome
sources := $(wildcard *.go)
version := 0.0.18-beta

build = GOOS=$(1) GOARCH=$(2) go build -v -o $(appname) $(4)
tar = mkdir -p build && cd ../ && tar -cvzf ./$(appname)_$(1)_$(2).tar.gz $(workingdir)/$(appname) $(workingdir)/web/out $(workingdir)/web/html $(workingdir)/web/assets && mv $(appname)_$(1)_$(2).tar.gz $(workingdir)/build

.PHONY: all linux

all:	linux

# Setup
setup:
	go mod tidy
	cd web && npm i

# Testing
test:
	cd docker/testing && docker-compose up -d
	go test -v -p 1 ./...
	# Tests should be run one after another due to deletion of the database at every test start

vtest:
	go test -p 1 ./... -coverprofile=coverage.out
	go tool cover --html=coverage.out -o coverage.html

# Updating the current version in all locations
version:
	python3 update_version.py

# Change version on build
release: cleanall version test build

vite-dev:
	cd web && npm run dev

# Run
run: web
	go run -v .

run-full: web mysql
	go run -v .

# Cleaning
clean: cleanweb
	rm -rf app
	rm -rf bin
	rm -rf log
	rm -rf docker/app
	rm -rf docker/bin
	rm -rf docker/homescript
	rm -rf docker/homescript_linux_amd64.tar.gz
	rm -rf docker/smarthome
	rm coverage.out
	rm coverage.html

cleanweb:
	rm -rf web/out

cleanall: clean
	rm -rf build
	rm -f smarthome

# Mysql Database
mysql:
	sudo systemctl start docker
	cd docker && docker-compose up -d

# Builds
build: web all linux clean

docker: cleanall web test
	GOOS=linux GOARCH=amd64 go build -v -o smarthome -ldflags '-extldflags "-fno-PIC -static"' -buildmode pie -tags 'osusergo netgo static_build' 
	mkdir docker/app
	# rsync -rv --exclude=data/avatars data docker/app/
	rsync -rv --exclude=web/src --exclude=web/node_modules --exclude=web/*.json web docker/app/
	cp smarthome docker/app/
	cd docker && docker build . -t mikmuellerdev/smarthome:$(version)

web: cleanweb
	cd web && npm run build

# Build architectures
linux: build/linux_arm.tar.gz build/linux_arm64.tar.gz build/linux_386.tar.gz build/linux_amd64.tar.gz

build/linux_386.tar.gz: $(sources)
	$(call build,linux,386,)
	$(call tar,linux,386)

build/linux_amd64.tar.gz: $(sources)
	$(call build,linux,amd64, -ldflags '-extldflags "-fno-PIC -static"' -buildmode pie -tags 'osusergo netgo static_build')
	$(call tar,linux,amd64)

build/linux_arm.tar.gz: $(sources)
	$(call build,linux,arm,)
	$(call tar,linux,arm)

build/linux_arm64.tar.gz: $(sources)
	$(call build,linux,arm64,)
	$(call tar,linux,arm64)

