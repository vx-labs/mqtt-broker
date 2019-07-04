VERSION = $(shell git describe --abbrev=0 --tags)

build:
	docker build -t quay.io/vxlabs/mqtt-broker:${VERSION} .
	docker build -f Dockerfile.listener  -t quay.io/vxlabs/mqtt-listener:${VERSION} .
release: build
	docker push quay.io/vxlabs/mqtt-broker:${VERSION}
	docker push quay.io/vxlabs/mqtt-listener:${VERSION}
deploy: release
	cd terraform && terraform apply -var broker_version=${VERSION}
nuke:
	cd terraform && terraform destroy
