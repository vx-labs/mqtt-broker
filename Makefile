VERSION = $(shell git rev-parse --short HEAD)
DOCKER_BUILD_ARGS = --network host --build-arg https_proxy=${https_proxy} --build-arg BUILT_VERSION=${VERSION}

build::
	docker build ${DOCKER_BUILD_ARGS} -t quay.io/vxlabs/mqtt-broker:${VERSION} .
release:: build release-nodep
deploy: release deploy-nodep
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
deploy-router-nodep::
	cd terraform/router && terraform init && terraform apply -auto-approve -var router_version=${VERSION}
deploy-topics-nodep::
	cd terraform/topics && terraform init && terraform apply -auto-approve -var topics_version=${VERSION}
nuke:
	cd terraform/api && terraform init && terraform destroy -auto-approve -var api_version=${VERSION}
	cd terraform/broker && terraform init && terraform destroy -auto-approve -var broker_version=${VERSION}
	cd terraform/listener && terraform init && terraform destroy -auto-approve -var listener_version=${VERSION}
	cd terraform/sessions && terraform init && terraform destroy -auto-approve -var sessions_version=${VERSION}
	cd terraform/subscriptions && terraform init && terraform destroy -auto-approve -var subscriptions_version=${VERSION}
	cd terraform/queues && terraform init && terraform destroy -auto-approve -var queues_version=${VERSION}
	cd terraform/messages && terraform init && terraform destroy -auto-approve -var messages_version=${VERSION}
	cd terraform/kv && terraform init && terraform destroy -auto-approve -var kv_version=${VERSION}
	cd terraform/router && terraform init && terraform destroy -auto-approve -var router_version=${VERSION}
	cd terraform/topics && terraform init && terraform destroy -auto-approve -var topics_version=${VERSION}
