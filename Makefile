VERSION = $(shell git rev-parse --short HEAD)
DOCKER_BUILD_ARGS = --network host --build-arg https_proxy=${https_proxy} --build-arg BUILT_VERSION=${VERSION}

build::
	docker build ${DOCKER_BUILD_ARGS} -t quay.io/vxlabs/mqtt-broker:${VERSION} .
release:: build release-nodep
deploy: release
		cd terraform/ && terraform init && terraform apply -auto-approve -var version=${VERSION}
deploy-nodep: deploy-api-nodep deploy-broker-nodep deploy-listener-nodep deploy-sessions-nodep deploy-subscriptions-nodep deploy-queues-nodep deploy-messages-nodep deploy-kv-nodep deploy-router-nodep deploy-topics-nodep

release-nodep:
	docker push quay.io/vxlabs/mqtt-broker:${VERSION}

deploy-api:: release deploy-api-nodep
deploy-broker:: release deploy-broker-nodep
deploy-listener:: release deploy-listener-nodep
deploy-sessions:: release deploy-sessions-nodep
deploy-subscriptions:: release deploy-subscriptions-nodep
deploy-queues:: release deploy-queues-nodep
deploy-messages:: release deploy-messages-nodep
deploy-kv:: release deploy-kv-nodep
deploy-router:: release deploy-router-nodep
deploy-topics:: release deploy-topics-nodep

deploy-api-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var version=${VERSION}
deploy-broker-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var version=${VERSION}
deploy-listener-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var version=${VERSION}
deploy-sessions-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var version=${VERSION}
deploy-subscriptions-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var version=${VERSION}
deploy-queues-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var version=${VERSION}
deploy-messages-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var version=${VERSION}
deploy-kv-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var version=${VERSION}
deploy-router-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var version=${VERSION}
deploy-topics-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var version=${VERSION}
nuke:
	cd terraform/ && terraform init && terraform destroy -auto-approve -var version=${VERSION}
