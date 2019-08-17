VERSION = $(shell git describe --abbrev=0 --tags)
DOCKER_BUILD_ARGS = --network host --build-arg https_proxy=${https_proxy}

build:: build-api build-broker build-listener build-sessions
release:: release-api release-broker release-listener release-sessions
deploy: deploy-api deploy-broker deploy-listener deploy-sessions

build-api:: build-common
	docker build ${DOCKER_BUILD_ARGS} --target api -t quay.io/vxlabs/mqtt-api:${VERSION} .
build-broker:: build-common
	docker build ${DOCKER_BUILD_ARGS} --target broker -t quay.io/vxlabs/mqtt-broker:${VERSION} .
build-listener:: build-common
	docker build ${DOCKER_BUILD_ARGS} --target listener -t quay.io/vxlabs/mqtt-listener:${VERSION} .
build-sessions:: build-common
	docker build ${DOCKER_BUILD_ARGS} --target sessions -t quay.io/vxlabs/mqtt-sessions:${VERSION} .

release-api:: build-api
	docker push quay.io/vxlabs/mqtt-api:${VERSION}
release-broker:: build-broker
	docker push quay.io/vxlabs/mqtt-broker:${VERSION}
release-listener:: build-listener
	docker push quay.io/vxlabs/mqtt-listener:${VERSION}
release-sessions:: build-sessions
	docker push quay.io/vxlabs/mqtt-sessions:${VERSION}

deploy-api:: release-api
	cd terraform/api && terraform init      && terraform apply -auto-approve -var api_version=${VERSION}
deploy-broker:: release-broker
	cd terraform/broker && terraform init   && terraform apply -auto-approve -var broker_version=${VERSION}
deploy-listener:: release-listener
	cd terraform/listener && terraform init && terraform apply -auto-approve -var listener_version=${VERSION}
deploy-sessions:: release-sessions
	cd terraform/sessions && terraform init && terraform apply -auto-approve -var sessions_version=${VERSION}

nuke:
	cd terraform/api && terraform init && terraform destroy -auto-approve -var api_version=${VERSION}
	cd terraform/broker && terraform init && terraform destroy -auto-approve -var broker_version=${VERSION}
	cd terraform/listener && terraform init && terraform destroy -auto-approve -var listener_version=${VERSION}
	cd terraform/sessions && terraform init && terraform destroy -auto-approve -var sessions_version=${VERSION}

build-common::
	docker build ${DOCKER_BUILD_ARGS} --target builder .
