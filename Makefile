appname := smarthome
workingdir := smarthome
sources := $(wildcard *.go)

build = GOOS=$(1) GOARCH=$(2) go build -o ./bin/$(appname)$(3) $(4)
tar = mkdir -p build && cd ../ && tar -cvzf ./$(appname)_$(1)_$(2).tar.gz $(applicationDir)/bin $(applicationDir)/web/out $(applicationDir)/web/html $(applicationDir)/web/assets && mv $(appname)_$(1)_$(2).tar.gz $(applicationDir)/build

# Run
run: web
	go run .

run-full: web mysql
	go run .

# Cleaning
clean: cleanweb
	rm -rf app
	rm -rf log
	rm -rf docker/app

cleanweb:
	rm -rf web/out

cleanall: clean
	rm -rf build bin

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
	sudo systemctl start docker
	cd docker && sudo docker build . -t mikmuellerdev/smarthome

web: cleanweb
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

