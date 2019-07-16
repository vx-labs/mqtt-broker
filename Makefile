VERSION = $(shell git describe --abbrev=0 --tags)

build:
	docker build -t quay.io/vxlabs/mqtt-broker:${VERSION} .
	docker build -f Dockerfile.listener  -t quay.io/vxlabs/mqtt-listener:${VERSION} .
	docker build -f Dockerfile.api  -t quay.io/vxlabs/mqtt-api:${VERSION} .
release: build
	docker push quay.io/vxlabs/mqtt-broker:${VERSION}
	docker push quay.io/vxlabs/mqtt-listener:${VERSION}
	docker push quay.io/vxlabs/mqtt-api:${VERSION}
deploy: release
	cd terraform && terraform apply -var broker_version=${VERSION} -var listener_version=${VERSION} -var api_version=${VERSION}
nuke:
	cd terraform && terraform destroy
