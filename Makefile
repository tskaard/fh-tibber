version="0.0.7"
version_file=VERSION
working_dir=$(shell pwd)
arch="armhf"

clean:
	-rm ./src/tibber

build-go:
	cd ./src;go build -o tibber service.go;cd ../

build-go-arm:
	cd ./src;GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="-s -w" -o tibber service.go;cd ../

build-go-amd:
	cd ./src;GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o tibber src/service.go;cd ../


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
	mkdir -p package/debian/var/log/futurehome/tibber package/debian/var/lib/futurehome/tibber/data
	cp ./src/tibber package/debian/usr/bin/tibber
	cp $(version_file) package/debian/var/lib/futurehome/tibber
	docker run --rm -v ${working_dir}:/build -w /build --name debuild debian dpkg-deb --build package/debian
	@echo "Done"

deb-arm: clean configure-arm build-go-arm package-deb-doc
	@echo "Building Futurehome ARM package"
	mv package/debian.deb package/build/tibber_$(version)_armhf.deb

deb-amd : configure-amd64 build-go-amd package-deb-doc
	@echo "Building Thingsplex AMD package"
	mv package/debian.deb tibber_$(version)_amd64.deb

run :
	cd ./src; go run service.go -c testdata/var/config.json;cd ../


.phony : clean
