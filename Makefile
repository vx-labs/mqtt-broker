VERSION = $(shell git describe --abbrev=0 --tags)
DOCKER_BUILD_ARGS = --network host --build-arg https_proxy=${https_proxy}

build:
	docker build ${DOCKER_BUILD_ARGS} --target broker -t quay.io/vxlabs/mqtt-broker:${VERSION} .
	docker build ${DOCKER_BUILD_ARGS} --target api -t quay.io/vxlabs/mqtt-api:${VERSION} .
	docker build ${DOCKER_BUILD_ARGS} --target listener -t quay.io/vxlabs/mqtt-listener:${VERSION} .
release: build
	docker push quay.io/vxlabs/mqtt-broker:${VERSION}
	docker push quay.io/vxlabs/mqtt-listener:${VERSION}
	docker push quay.io/vxlabs/mqtt-api:${VERSION}
deploy: release
	cd terraform && terraform apply -var broker_version=${VERSION} -var listener_version=${VERSION} -var api_version=${VERSION}
deploy-nodep:
	cd terraform && terraform apply -var broker_version=${VERSION} -var listener_version=${VERSION} -var api_version=${VERSION}
nuke:
	cd terraform && terraform destroy
