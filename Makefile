VERSION = $(shell git rev-parse --short HEAD)
DOCKER_BUILD_ARGS = --network host --build-arg https_proxy=${https_proxy} --build-arg BUILT_VERSION=${VERSION}

build::
	docker build ${DOCKER_BUILD_ARGS} -t quay.io/vxlabs/mqtt-broker:${VERSION} .
release:: build release-nodep
deploy: release deploy-nodep
deploy-nodep:
		cd terraform/ && terraform init && terraform apply -auto-approve -var image_tag=${VERSION}

release-nodep:
	docker push quay.io/vxlabs/mqtt-broker:${VERSION}

deploy-api:: release deploy-api-nodep
deploy-auth:: release deploy-auth-nodep
deploy-broker:: release deploy-broker-nodep
deploy-listener:: release deploy-listener-nodep
deploy-sessions:: release deploy-sessions-nodep
deploy-subscriptions:: release deploy-subscriptions-nodep
deploy-queues:: release deploy-queues-nodep
deploy-messages:: release deploy-messages-nodep
deploy-kv:: release deploy-kv-nodep
deploy-router:: release deploy-router-nodep
deploy-topics:: release deploy-topics-nodep
deploy-endpoints:: release deploy-endpoints-nodep

deploy-api-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var image_tag=${VERSION} -target nomad_job.api
deploy-auth-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var image_tag=${VERSION} -target module.auth
deploy-broker-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var image_tag=${VERSION} -target module.broker
deploy-listener-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var image_tag=${VERSION} -target nomad_job.listener
deploy-sessions-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var image_tag=${VERSION} -target module.sessions
deploy-subscriptions-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var image_tag=${VERSION} -target module.subscriptions
deploy-queues-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var image_tag=${VERSION} -target module.queues
deploy-messages-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var image_tag=${VERSION} -target module.messages
deploy-kv-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var image_tag=${VERSION} -target module.kv
deploy-router-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var image_tag=${VERSION} -target module.router
deploy-topics-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var image_tag=${VERSION} -target module.topics
deploy-endpoints-nodep::
	cd terraform/ && terraform init && terraform apply -auto-approve -var image_tag=${VERSION} -target module.endpoints
nuke:
	cd terraform/ && terraform init && terraform destroy -auto-approve -var image_tag=${VERSION}
