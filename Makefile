version_file=VERSION
working_dir=$(shell pwd)
arch="armhf"
# Remove `v` from the tag: v0.0.7 -> 0.0.7
version:=`git describe --tags | cut -c 2-`
#version:="1.0.1"
remote_host = "fh@cube.local"

clean:
	-rm -f ./tibber

init:
	git config core.hooksPath .githooks

build-go:
	go build -o tibber service.go

build-go-arm: init
	GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="-s -w" -o tibber service.go

build-go-amd: init
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o tibber service.go


configure-arm:
	python ./scripts/config_env.py prod $(version) armhf

configure-amd64:
	python ./scripts/config_env.py prod $(version) amd64


package-tar:
	tar cvzf tibber_$(version).tar.gz tibber $(version_file)

clean-deb:
	find package -name ".DS_Store" -delete
	find package -name "delete_me" -delete

package-deb-doc:clean-deb
	@echo "Packaging application as debian package"
	chmod a+x package/debian/DEBIAN/*
	mkdir -p package/debian/var/log/futurehome/tibber package/debian/var/lib/futurehome/tibber/data package/debian/usr/bin
	mkdir -p package/build
	cp ./tibber package/debian/usr/bin/tibber
	cp $(version_file) package/debian/var/lib/futurehome/tibber
#	dpkg-deb --build package/debian
	docker run --rm -v ${working_dir}:/build -w /build --name debuild debian dpkg-deb --build package/debian

deb-arm: clean configure-arm build-go-arm package-deb-doc
	@echo "Building Futurehome ARM package"
	@mv package/debian.deb package/build/tibber_$(version)_armhf.deb
	@echo "Created package/build/tibber_$(version)_armhf.deb"

deb-amd : configure-amd64 build-go-amd package-deb-doc
	@echo "Building Thingsplex AMD package"
	mv package/debian.deb package/build/tibber_$(version)_amd64.deb

upload :
	scp package/build/tibber_$(version)_armhf.deb $(remote_host):~/

remote-install : upload
	ssh -t $(remote_host) "sudo dpkg -i tibber_$(version)_armhf.deb"

deb-remote-install : deb-arm remote-install
	@echo "Installed on remote host"

run :
	 go run service.go -c testdata


.phony : clean
