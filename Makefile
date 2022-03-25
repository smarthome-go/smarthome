appname := smarthome
workingdir := smarthome
sources := $(wildcard *.go)
homescript_cli_version := v0.3.0-beta

build = GOOS=$(1) GOARCH=$(2) go build -o $(appname) $(4)
tar = mkdir -p build && cd ../ && tar -cvzf ./$(appname)_$(1)_$(2).tar.gz $(workingdir)/$(appname) $(workingdir)/web/out $(workingdir)/web/html $(workingdir)/web/assets && mv $(appname)_$(1)_$(2).tar.gz $(workingdir)/build

.PHONY: all linux

all:	linux

# Setup
setup:
	go mod tidy
	cd web && npm i

# Updating the current version in all locations
version:
	python3 update_version.py

# Run
run: web
	go run . &
	cd web && npm run watch

run-full: web mysql
	go run .

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

docker: cleanall web
	GOOS=linux GOARCH=amd64 go build -o smarthome -ldflags '-extldflags "-fno-PIC -static"' -buildmode pie -tags 'osusergo netgo static_build' 
	mkdir docker/app
	# rsync -rv --exclude=data/avatars data docker/app/
	rsync -rv --exclude=web/src web docker/app/
	cp smarthome docker/app/
	cd docker && wget "https://github.com/MikMuellerDev/homescript-cli/releases/download/$(homescript_cli_version)/homescript_linux_amd64.tar.gz"
	cd docker && tar -xvf homescript_linux_amd64.tar.gz
	cd docker && mv bin/homescript .
	cd docker && docker build . -t mikmuellerdev/smarthome

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

