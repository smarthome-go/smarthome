appname := smarthome
workingdir := smarthome
sources := $(wildcard *.go)

build = GOOS=$(1) GOARCH=$(2) go build -o ./bin/$(appname)$(3) $(4)
tar = mkdir -p build && cd ../ && tar -cvzf ./$(appname)_$(1)_$(2).tar.gz $(applicationDir)/bin $(applicationDir)/web/out $(applicationDir)/web/html $(applicationDir)/web/assets && mv $(appname)_$(1)_$(2).tar.gz $(applicationDir)/build

# Cleaning
clean:
	rm -rf web/out
	rm -rf app
	rm -rf log

cleanall: clean
	rm -rf build bin

# Mysql Database
mysql:
	sudo systemctl start docker
	cd docker && sudo docker-compose up -d

# Run
run-full: web mysql
	go run .

run: web
	go run .

# Builds
build: web all linux clean

web: clean
	cd web && npm run typescript-build && npm run postcss-build

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

