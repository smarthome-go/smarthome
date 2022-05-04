appname := smarthome
workingdir := smarthome
sources := $(wildcard *.go)
version := 0.0.26-beta-rc.3

build = CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build -ldflags "-s -w" -v -o $(appname) $(4)
tar = mkdir -p build && cd ../ && tar -cvzf ./$(appname)_$(1)_$(2).tar.gz $(workingdir)/$(appname) $(workingdir)/web/dist $(workingdir)/web/html $(workingdir)/resources && mv $(appname)_$(1)_$(2).tar.gz $(workingdir)/build

.PHONY: all linux

all:	linux

# Setup
setup:
	go mod tidy
	cd web && npm i
	cd web && npm run prepare

# Testing
test:
	mkdir -p web/dist/html
	touch web/dist/html/testing.html
	# Prevents server panic

	cd docker/testing && docker-compose up -d
	go test -v -p 1 ./... --timeout=10000s
	# Tests should be run one after another due to deletion of the database at every test start
	rm -rf web/dist/html/testing.html

vtest:
	mkdir -p web/dist/html
	touch web/dist/html/testing.html
	# Prevents server panic

	go test -v -p 1 ./... -coverprofile=coverage.out
	go tool cover --html=coverage.out -o coverage.html
	rm -rf web/dist/html/testing.html

# Updating the current version in all locations
version:
	python3 update_version.py

# Change version on build
release: cleanall version
	make test
	make build
	make docker

vite-dev:
	cd web && npm run dev

# Run
run: web
	go run -v -race .

run-full: web mysql
	go run -v -race .

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
	rm -rf coverage.out
	rm -rf coverage.html

cleanweb:
	rm -rf web/dist

cleanall: clean
	rm -rf build
	rm -f smarthome

# Mysql Database
mysql:
	sudo systemctl start docker
	cd docker && docker-compose up -d

# Builds
build: setup web all linux clean

docker-prepare:
	CGO_ENABLED=0 GOOS=linux go build -v -installsuffix cgo -ldflags '-s -w' -o smarthome
	mkdir -p docker/app/web
	rsync -rv resources docker/app/
	rsync -rv web/dist docker/app/web/
	cp smarthome docker/app/

docker-push:
	docker push mikmuellerdev/smarthome:$(version)
	docker push mikmuellerdev/smarthome:latest

docker: cleanall web docker-prepare
	cd docker && docker build . -t mikmuellerdev/smarthome:$(version) -t mikmuellerdev/smarthome:latest --network=host

sudo-docker: cleanall web docker-prepare
	cd docker && sudo docker build . -t mikmuellerdev/smarthome:$(version) -t mikmuellerdev/smarthome:latest --network=host

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

