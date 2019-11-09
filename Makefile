VERSION = $(shell git rev-parse --short HEAD)
DOCKER_BUILD_ARGS = --network host --build-arg https_proxy=${https_proxy} --build-arg BUILT_VERSION=${VERSION}

build:: build-api build-broker build-listener build-sessions build-subscriptions build-queues
release:: release-api release-broker release-listener release-sessions release-subscriptions release-queues
deploy: deploy-api deploy-broker deploy-listener deploy-sessions deploy-subscriptions deploy-queues
deploy-nodep: deploy-api-nodep deploy-broker-nodep deploy-listener-nodep deploy-sessions-nodep deploy-subscriptions-nodep deploy-queues-nodep deploy-messages-nodep deploy-kv-nodep

build-api:: build-common
	docker build ${DOCKER_BUILD_ARGS} --build-arg ARTIFACT=api -t quay.io/vxlabs/mqtt-api:${VERSION} .
build-broker:: build-common
	docker build ${DOCKER_BUILD_ARGS} --build-arg ARTIFACT=broker -t quay.io/vxlabs/mqtt-broker:${VERSION} .
build-listener:: build-common
	docker build ${DOCKER_BUILD_ARGS} --build-arg ARTIFACT=listener -t quay.io/vxlabs/mqtt-listener:${VERSION} .
build-sessions:: build-common
	docker build ${DOCKER_BUILD_ARGS} --build-arg ARTIFACT=sessions -t quay.io/vxlabs/mqtt-sessions:${VERSION} .
build-subscriptions:: build-common
	docker build ${DOCKER_BUILD_ARGS} --build-arg ARTIFACT=subscriptions -t quay.io/vxlabs/mqtt-subscriptions:${VERSION} .
build-queues:: build-common
	docker build ${DOCKER_BUILD_ARGS} --build-arg ARTIFACT=queues -t quay.io/vxlabs/mqtt-queues:${VERSION} .
build-messages:: build-common
	docker build ${DOCKER_BUILD_ARGS} --build-arg ARTIFACT=messages -t quay.io/vxlabs/mqtt-messages:${VERSION} .
build-kv:: build-common
	docker build ${DOCKER_BUILD_ARGS} --build-arg ARTIFACT=kv -t quay.io/vxlabs/mqtt-kv:${VERSION} .

release-api:: build-api
	docker push quay.io/vxlabs/mqtt-api:${VERSION}
release-broker:: build-broker
	docker push quay.io/vxlabs/mqtt-broker:${VERSION}
release-listener:: build-listener
	docker push quay.io/vxlabs/mqtt-listener:${VERSION}
release-sessions:: build-sessions
	docker push quay.io/vxlabs/mqtt-sessions:${VERSION}
release-subscriptions:: build-subscriptions
	docker push quay.io/vxlabs/mqtt-subscriptions:${VERSION}
release-queues:: build-queues
	docker push quay.io/vxlabs/mqtt-queues:${VERSION}
release-messages:: build-messages
	docker push quay.io/vxlabs/mqtt-messages:${VERSION}
release-kv:: build-kv
	docker push quay.io/vxlabs/mqtt-kv:${VERSION}

deploy-api:: release-api deploy-api-nodep
deploy-broker:: release-broker deploy-broker-nodep
deploy-listener:: release-listener deploy-listener-nodep
deploy-sessions:: release-sessions deploy-sessions-nodep
deploy-subscriptions:: release-subscriptions deploy-subscriptions-nodep
deploy-queues:: release-queues deploy-queues-nodep
deploy-messages:: release-messages deploy-messages-nodep
deploy-kv:: release-kv deploy-kv-nodep

deploy-api-nodep::
	cd terraform/api && terraform init      && terraform apply -auto-approve -var api_version=${VERSION}
deploy-broker-nodep::
	cd terraform/broker && terraform init   && terraform apply -auto-approve -var broker_version=${VERSION}
deploy-listener-nodep::
	cd terraform/listener && terraform init && terraform apply -auto-approve -var listener_version=${VERSION}
deploy-sessions-nodep::
	cd terraform/sessions && terraform init && terraform apply -auto-approve -var sessions_version=${VERSION}
deploy-subscriptions-nodep::
	cd terraform/subscriptions && terraform init && terraform apply -auto-approve -var subscriptions_version=${VERSION}
deploy-queues-nodep::
	cd terraform/queues && terraform init && terraform apply -auto-approve -var queues_version=${VERSION}
deploy-messages-nodep::
	cd terraform/messages && terraform init && terraform apply -auto-approve -var messages_version=${VERSION}
deploy-kv-nodep::
	cd terraform/kv && terraform init && terraform apply -auto-approve -var kv_version=${VERSION}
nuke:
	cd terraform/api && terraform init && terraform destroy -auto-approve -var api_version=${VERSION}
	cd terraform/broker && terraform init && terraform destroy -auto-approve -var broker_version=${VERSION}
	cd terraform/listener && terraform init && terraform destroy -auto-approve -var listener_version=${VERSION}
	cd terraform/sessions && terraform init && terraform destroy -auto-approve -var sessions_version=${VERSION}
	cd terraform/subscriptions && terraform init && terraform destroy -auto-approve -var subscriptions_version=${VERSION}
	cd terraform/queues && terraform init && terraform destroy -auto-approve -var queues_version=${VERSION}
	cd terraform/messages && terraform init && terraform destroy -auto-approve -var messages_version=${VERSION}
	cd terraform/kv && terraform init && terraform destroy -auto-approve -var kv=${VERSION}

build-common::
	docker build ${DOCKER_BUILD_ARGS} --target builder .

dockerfiles::
	for svc in api listener subscriptions sessions broker queues messages kv; do sed "s/###/$$svc/g" Dockerfile > Dockerfile.$$svc; done
