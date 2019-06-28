# default make -s
# run `make VERBOSE=1` to make it print verbose
# ifndef VERBOSE
# .SILENT:
# endif

# Deprecated
# always `make -s`
# MAKEFLAGS += --silent

.PHONY: default all prepare build product push local-cluster rm-local-cluster integration-test clean

PWD=$(shell pwd)
CACHEDIR=$(shell go env GOCACHE)
ifeq ($(CACHEDIR),)  # ---> CACHEDIR == ""
CACHEDIR="/tmp/.gocache"
endif
PRJNAME := "adbot"
Compose := "https://github.com/docker/compose/releases/download/1.14.0/docker-compose"
ImgMaster := "bbklab/adbot-master"
ImgAgent  := "bbklab/adbot-agent"

# Used to populate version variable in main package.
VERSION=$(shell cat VERSION.txt)
BUILD_TIME=$(shell date -u +%Y-%m-%d_%H:%M:%S_%Z)
PKG := "github.com/bbklab/adbot"
gopkgs=$(shell go list ./... | grep -v "integration-test")
gitCommit=$(shell git rev-parse --short HEAD)
gitDirty=$(shell git status --porcelain --untracked-files=no)
GIT_COMMIT=$(gitCommit)
ifneq ($(gitDirty),)  # ---> gitDirty != ""
GIT_COMMIT=$(gitCommit)-dirty
endif
BUILD_FLAGS=-X $(PKG)/version.version=$(VERSION) -X $(PKG)/version.gitCommit=$(GIT_COMMIT) -X $(PKG)/version.buildAt=$(BUILD_TIME) -w -s

default: binary

nolint: binary-nolint

all: integration-test push

# product rpm build steps:
#   binary  -> master rpm (with agent rpm within it)
prod: product

prod-full: product-full

product: product-fast

product-full: product-fast geo-rpm dep-rpm
	echo "result @ ./product/"

product-fast: clean binary master-rpm
	echo "result @ ./product/"

# anonymous product build
#   - hide or obfuscate all of personal identifiers for security concerns
#   - same as product build but with env ANONYMOUS=1
prod-anonymous: product-anonymous

prod-anonymous-full: product-anonymous-full

product-anonymous:
	@make ANONYMOUS=1 product

product-anonymous-full:
	@make ANONYMOUS=1 product-full

# image build:
image: clean binary agent-image master-image

prepare:
	mkdir -p $(CACHEDIR)
	mkdir -p bin/
	mkdir -p product/

#
# gofmt & golint & govet
#
gocheck: gometalinter

# use gometalinter to replace all of above go linters
gometalinter:
	docker run --name gocheck-adbot --rm \
		-v $(PWD):/go/src/${PKG}:ro \
		bbklab/gometalinter:latest \
		gometalinter --skip=integration-test --skip=vendor \
			--deadline=120s \
			--disable-all \
			--enable=gofmt --enable=golint \
			--enable=vet --enable=goconst \
			--enable=deadcode \
			/go/src/${PKG}/...
	echo "  --- Gometalinter Passed!"

#
# binary
#
binary: prepare
ifeq (${NOLINT},)   # NOLINT == ""
	@make gocheck
endif
ifneq (${ANONYMOUS},)   # ANONYMOUS != ""
	@anonymous-build.sh
else
	@make binary-nolint
endif

binary-nolint:
ifeq (${ENV_CIRCLECI}, true)
	@make host-binary-build     # circleCI
else
	@make docker-binary-build   # travisCI / host
endif

#
# docker build binary (build binary via docker container)
#
docker-binary-build:
	docker run --rm \
		--name buildadbot \
		-w /go/src/${PKG} \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-e GOCACHE=/go/cache \
		-v $(PWD):/go/src/${PKG}:rw \
		-v $(CACHEDIR):/go/cache/:rw \
		golang:1.10-alpine \
		sh -c 'go build -ldflags "${BUILD_FLAGS}" -o bin/adbot ${PKG}/cmd/adbot'
	echo "Binary Built!"

#
# host build binary(direct build binary by using system installed golang)
#  mainly used for CI env
#
host-binary-build:
	env CGO_ENABLED=0 GOOS=linux go build -ldflags "${BUILD_FLAGS}" -o bin/adbot ${PKG}/cmd/adbot
	echo "Binary Built!"

#
# agent
#
agent-image:
	docker build --force-rm -t $(ImgAgent):$(gitCommit) -f Dockerfile.agent .   # optional: --no-cache
	docker tag $(ImgAgent):$(gitCommit) $(ImgAgent):latest
	echo "Agent Image Built!"

#
# master
# 
master-image: agent-pkg
	docker build --force-rm -t $(ImgMaster):$(gitCommit) -f Dockerfile.master .   # optional: --no-cache
	docker tag $(ImgMaster):$(gitCommit) $(ImgMaster):latest
	echo "Master Image Built!"

master-rpm:
	mkdir -p contrib/rpm/bin/
	cp -avf bin/* contrib/rpm/bin/
	./contrib/rpm/makedist.sh master
	mv -fv ./contrib/rpm/dist/adbot-master-*.rpm ./product/
	echo "Master RPM Built!"

agent-pkg:
	mkdir -p contrib/rpm/bin/
	cp -avf bin/* contrib/rpm/bin/
	./contrib/rpm/makedist.sh agent
	echo "Agent PKG Built! (only for master docker image)"

#
# geolite2
#
geo-rpm:
	./contrib/rpm/makedist.sh geolite2
	mv -fv ./contrib/rpm/dist/adbot-geolite2-*.rpm ./product/
	echo "GeoLite2 RPM Built!"

#
# dependency
#
dep-rpm:
	./contrib/rpm/makedist.sh dependency
	mv -fv ./contrib/rpm/dist/adbot-dependency-*.rpm ./product/
	echo "Dependency RPM Built!"

# update dep
#  - HTTP[S]_PROXY:  golang use proxy
#  - ALL_PROXY:   git use proxy
update-dep:
	cd ${GOPATH}/src/$(PKG) && \
		env \
		HTTP_PROXY=socks5://127.0.0.1:1080 HTTPS_PROXY=socks5://127.0.0.1:1080 \
		ALL_PROXY=socks5://127.0.0.1:1080 \
		dep ensure -v

# push
# 
push: agent-image master-image
	docker push $(ImgMaster):$(gitCommit)
	docker push $(ImgMaster):latest
	echo "Images Pushed!"


# unit test & coverage
#
unit-test:
	@echo "mode: count" > coverage.txt    # this file could be picked up by https://codecov.io/bash
	@for pkg in $(gopkgs); do \
		echo "go test $${pkg} ..."; \
		go test -v -coverprofile=coverage.profile -covermode=count $${pkg}; \
		tail -n +2 coverage.profile >> coverage.txt; \
		rm -f coverage.profile; \
	done
	@echo "result @ coverage.txt"

unit-test-codecov: unit-test
	bash -c 'env CODECOV_TOKEN="99121f5d-25ba-4dff-af61-fb7f9454024b" bash <(curl -s https://codecov.io/bash) -X fix'

unit-test-html: unit-test
	go tool cover -html=coverage.txt -o coverage.html
	rm -f coverage.txt
	echo "result @ coverage.html"

prepare-gocov-unit-test:
	@if ! command -v gocov > /dev/null 2>&1; then \
		go get -u github.com/axw/gocov/gocov; \
		echo "gocov installed!"; \
	fi
	@if ! command -v gocov-html > /dev/null 2>&1; then \
		go get -u go get gopkg.in/matm/v1/gocov-html; \
		echo "gocov-html installed!"; \
	fi

gocov-unit-test: prepare-gocov-unit-test
	gocov test $(gopkgs) | gocov report

gocov-unit-test-html: prepare-gocov-unit-test
	gocov test $(gopkgs) | gocov-html > coverage-pretty.html
	echo "result @ coverage-pretty.html"

#
# local cluster via docker-compose
# integration test via local cluster
#
prepare-docker-compose:
	@if ! command -v docker-compose > /dev/null 2>&1; then \
		echo "docker-compose downloading ..."; \
		curl --progress-bar -L $(Compose)-$(shell uname -s)-$(shell uname -m) -o \
			/usr/local/bin/docker-compose; \
		chmod +x /usr/local/bin/docker-compose; \
		echo "docker-compose downloaded!"; \
	fi

# apply a new dev license
prepare-dev-license:
	@cd ./contrib/tools/ && make license && sleep 1

# require docker daemon >= 1.13 to support docker compose v3
local-cluster: image prepare-docker-compose rm-local-cluster prepare-dev-license
	docker-compose -p ${PRJNAME} up -d --scale master=1 --scale agent=0
	docker-compose -p ${PRJNAME} ps

rm-local-cluster: prepare-docker-compose
	docker-compose -p ${PRJNAME} stop
	docker-compose -p ${PRJNAME} rm -f

integration-test: local-cluster run-integration-test rm-local-cluster

run-integration-test:
	docker run --name=testadbot --rm \
		--net=host \
		-w /go/src/${PKG}/integration-test \
		-e API_HOST=127.0.0.1:8008 \
		-e "TESTON=" \
		-e "CGO_ENABLED=0" \
		-v $(shell pwd):/go/src/${PKG} \
		golang:1.10-alpine \
		sh -c 'go test -check.v -test.timeout=5m ${PKG}/integration-test'

#
# local circleci full cluster (with mongodb container) via docker-compose
# mainly used under circleci env
#
local-circleci-cluster: prepare-docker-compose
	docker-compose -p ${PRJNAME} -f docker-compose.circleci.yml up -d
	docker-compose -p ${PRJNAME} -f docker-compose.circleci.yml ps

# mainly for debug startup failure
local-circleci-cluster-logs:
	docker-compose -p ${PRJNAME} -f docker-compose.circleci.yml logs --tail=300

rm-local-circleci-cluster:
	docker-compose -p ${PRJNAME} -f docker-compose.circleci.yml stop
	docker-compose -p ${PRJNAME} -f docker-compose.circleci.yml rm -f

run-circleci-integration-test:
	docker run -d --name=testadbot \
		--net=host \
        -e API_HOST=127.0.0.1:8008 \
        -e "TESTON=" \
        -e "CGO_ENABLED=0" \
        golang:1.10-alpine \
		sleep 1000000000
	docker exec testadbot sh -c 'mkdir -p /go/src/${PKG}'
	docker cp -a . testadbot:/go/src/${PKG}/    # copy source codes to testadbot container
	docker exec testadbot sh -c 'cd /go/src/${PKG}/integration-test && go test -check.v -test.timeout=5m .'
	docker rm -f testadbot

#
# clean up outdated
# 
clean: rpmclean
	rm -fv  product/*
	rm -fv  coverage.txt
	rm -fv  coverage.html
	rm -fv  coverage-pretty.html

rpmclean:
	./contrib/rpm/makedist.sh clean
