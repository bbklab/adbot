sslkeys:
	sh new-sslkey.sh

# generate a new dev license and save as local filename 'naive-license.pem'
license:
ifeq (${ENV_CIRCLECI}, true)
	@make host-license   # circleCI
else
	@make docker-license # travisCI / host
endif

host-license:
	go run new-devlicense.go

docker-license:
	docker run --name=new-infdev-license --rm \
		-w /tmp/go \
		-v $(shell pwd):/tmp/go \
		golang:1.10-alpine \
		sh -c 'go run new-devlicense.go'
