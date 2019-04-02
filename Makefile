version="0.0.5"
version_file=VERSION
working_dir=$(shell pwd)
arch="armhf"

clean:
	-rm tpflow

build-go:
	cd ./src;go build -o fh-tibber service.go;cd ../

build-go-arm:
	cd ./src;GOOS=linux GOARCH=arm GOARM=6 go build -o fh-tibber service.go;cd ../

build-go-amd:
	cd ./src;GOOS=linux GOARCH=amd64 go build -o fh-tibber src/service.go;cd ../


configure-arm:
	python ./scripts/config_env.py prod $(version) armhf

configure-amd64:
	python ./scripts/config_env.py prod $(version) amd64


package-tar:
	tar cvzf fh-tibber_$(version).tar.gz fh-tibber VERSION

package-deb-doc:
	@echo "Packaging application as debian package"
	chmod a+x package/debian/DEBIAN/*
	mkdir -p package/debian/var/log/futurehome/fh-tibber package/debian/var/lib/futurehome/fh-tibber/data
	cp ./src/fh-tibber package/debian/usr/bin/fh-tibber
	cp VERSION package/debian/var/lib/futurehome/fh-tibber
	docker run --rm -v ${working_dir}:/build -w /build --name debuild debian dpkg-deb --build package/debian
	@echo "Done"

tar-arm: build-js build-go-arm package-deb-doc
	@echo "The application was packaged into tar archive "

deb-arm : clean configure-arm build-go-arm package-deb-doc
	mv package/debian.deb package/build/fh-tibber_$(version)_armhf.deb

deb-amd : configure-amd64 build-go-amd package-deb-doc
	mv debian.deb fh-tibber_$(version)_amd64.deb

run :
	cd ./src; go run service.go -c testdata/var/config.json;cd ../


.phony : clean
