#!/bin/bash
set -ex

#
# def
#
BASEDIR=$(cd $(dirname ${BASH_SOURCE[*]}); pwd -P)
DISTDIR="${BASEDIR}/dist"   # product rpm save dir
BINDIR="${BASEDIR}/bin"     # binary file dir
IMAGE="bbklab/centos7-buildrpm"

function clean() {
	rm -fv ${DISTDIR}/*
	rm -fv ${BINDIR}/*
	rm -fv ${BASEDIR}/adbot-agent/bin/*
	rm -fv ${BASEDIR}/adbot-master/bin/*
	rm -fv ${BASEDIR}/adbot-master/share/agent.pkg 
	rm -fv ${BASEDIR}/adbot-master/share/agent.pkg.sha1sum
}

function buildTarget() {
	local target=$1
	local cname="build_adbot_${target}_rpm_dist"

	if [ "$(docker inspect  -f {{.Name}} ${cname} 2>&-)" == "/${cname}" ]; then
		echo "a previous rpm building is running, abort."
		exit 1
	fi

	local dirname=
	if [ "${target}" == "agent" ]; then
		dirname="adbot-agent"
	elif [ "${target}" == "master" ]; then
		dirname="adbot-master"
	elif [ "${target}" == "geolite2" ]; then
		dirname="adbot-geolite2"
	elif [ "${target}" == "dependency" ]; then
		dirname="adbot-dependency"
	fi

	# local build
	if [ "${ENV_CIRCLECI}" != "true" ]; then
		docker run --rm --name=${cname} \
			-e IN_CONTAINER=yes \
			-v ${BASEDIR}/${dirname}:/buildrpm \
			-v ${DISTDIR}:/product \
			-v /usr/bin/docker:/usr/bin/docker \
			-v /var/run/docker.sock:/var/run/docker.sock \
			-w /buildrpm \
			${IMAGE} ./build.sh
		return
	fi

	# circleci env build
	# note: circleci is using remote docker env
	# See: https://circleci.com/docs/2.0/building-docker-images/#mounting-folders
	docker run -d --rm --name=${cname} \
		-e IN_CONTAINER=yes \
		${IMAGE} sleep 1000000000

	docker cp -a ${BASEDIR}/${dirname} ${cname}:/buildrpm   # copy source codes to remote container

	docker exec ${cname} sh -c 'mkdir -p /product && cd /buildrpm && ./build.sh'

	docker cp -a ${cname}:/product/. ${DISTDIR}
	docker rm -f ${cname}
}

#
# main
#

pushd ${BASEDIR}

# trap clean
if [ "$1" == "clean" ]; then
	clean
	exit
fi

# check parameter must be  master|geolite2|dependency
target=$1
if [ "$target" != "master" -a "$target" != "agent" -a "$target" != "geolite2" -a "$target" != "dependency" ]; then
	echo "parameter must be one of master|agent|geolite2|dependency"
	exit 1
fi

# check binary ready if build master
if [ "$target" == "master" -o "$target" == "agent" ]; then
	if [ ! -e "${BINDIR}/adbot" ]; then
		echo "binary adbot not ready yet before building master or agent"
		exit 1
	fi

	# install binary to agent & master directory
	mkdir -p adbot-agent/bin/
	cp -avf ${BINDIR}/adbot adbot-agent/bin/

	mkdir -p adbot-master/bin/
	cp -avf ${BINDIR}/adbot adbot-master/bin/
fi

# ready to fly
mkdir -p $DISTDIR

if [ "${target}" == "master" ]; then 

	# first build agent rpm and save to masater's share directory
	buildTarget "agent"
	mkdir -p adbot-master/share/
	mv -fv ${DISTDIR}/adbot-agent*.rpm adbot-master/share/agent.pkg
	sha1sum adbot-master/share/agent.pkg | cut -d " " -f1 > adbot-master/share/agent.pkg.sha1sum

	if [ ! -e adbot-master/share/agent.pkg ]; then
		echo "pls build agent rpm firstly."
		exit 1
	fi

	# build master rpm with agent rpm within it
	buildTarget "master"
	rm -fv adbot-master/share/agent.pkg adbot-master/share/agent.pkg.sha1sum

	# create latest link
	for rpm in $(ls ${DISTDIR}/*); do
		if [[ "${rpm}" =~ master ]]; then
			ln -svf ${rpm##*/} ${DISTDIR}/adbot-master-latest-rhel7.x86_64.rpm
		fi
	done

elif [ "${target}" == "agent" ]; then
	# mostly we build agent rpm independently only for master docker image which
	# will be used by local cluster containerized env setup
	buildTarget "agent"
	mv -fv ${DISTDIR}/adbot-agent*.rpm ${DISTDIR}/agent.pkg
	sha1sum ${DISTDIR}/agent.pkg | cut -d " " -f1 > ${DISTDIR}/agent.pkg.sha1sum

elif [ "${target}" == "geolite2" ]; then
	buildTarget "geolite2"

elif [ "${target}" == "dependency" ]; then
	buildTarget "dependency"

fi
