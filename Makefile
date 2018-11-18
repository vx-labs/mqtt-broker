VERSION = $(shell git describe --abbrev=0 --tags)

build:
	docker build -t quay.io/vxlabs/mqtt-broker:${VERSION} .
release: build
	docker push quay.io/vxlabs/mqtt-broker:${VERSION}
deploy: release
	cd terraform && terraform apply -var broker_version=${VERSION}
nuke:
	cd terraform && terraform destroy
