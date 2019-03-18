version="0.0.1~~pre1"
version_file=VERSION
working_dir=$(shell pwd)
arch="armhf"

clean:
	-rm tpflow

build-go:
	go build -o tibber src/service.go

build-go-arm:
	cd ./src;GOOS=linux GOARCH=arm GOARM=6 go build -o tibber service.go;cd ../

build-go-amd:
	cd ./src;GOOS=linux GOARCH=amd64 go build -o tibber src/service.go;cd ../


configure-arm:
	python ./scripts/config_env.py prod $(version) armhf

configure-amd64:
	python ./scripts/config_env.py prod $(version) amd64


package-tar:
	tar cvzf tibber_$(version).tar.gz tibber VERSION

package-deb-doc:
	@echo "Packaging application as debian package"
	chmod a+x package/debian/DEBIAN/*
	cp ./src/tibber package/debian/usr/bin/tibber
	cp VERSION package/debian/var/lib/futurehome/tibber
	docker run --rm -v ${working_dir}:/build -w /build --name debuild debian dpkg-deb --build package/debian
	@echo "Done"

tar-arm: build-js build-go-arm package-deb-doc
	@echo "The application was packaged into tar archive "

deb-arm : clean configure-arm build-go-arm package-deb-doc
	mv package/debian.deb package/build/tibber_$(version)_armhf.deb

deb-amd : configure-amd64 build-go-amd package-deb-doc
	mv debian.deb tibber_$(version)_amd64.deb

run :
	cd ./src; go run service.go -c testdata/var/config.json;cd ../


.phony : clean
