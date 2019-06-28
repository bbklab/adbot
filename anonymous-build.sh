#!/bin/bash
#
# Most Same As Makefile
#
set -e

# setup env (note: same as Makefile)
PWD=$(pwd)
CACHEDIR=$(go env GOCACHE)
if [ "${CACHEDIR}" == "" ]; then
	CACHEDIR="/tmp/.gocache"
fi

VERSION=$(cat VERSION.txt)
BUILD_TIME=$(date -u +%Y-%m-%d_%H:%M:%S_%Z)
PKG="github.com/codingbot/adbot"   # anonymous github username

gitCommit=$(git rev-parse --short HEAD)
gitDirty=$(git status --porcelain --untracked-files=no)
GIT_COMMIT=${gitCommit}
if [ "${gitDirty}" != "" ]; then
	GIT_COMMIT="${gitCommit}-dirty"
fi

BUILD_FLAGS="-X ${PKG}/version.version=${VERSION} -X ${PKG}/version.gitCommit=${GIT_COMMIT} -X ${PKG}/version.buildAt=${BUILD_TIME} -w -s"


# prepare build dir
ANONYBUILDDIR="${PWD}/.anonymous-build"
rm -rf ${ANONYBUILDDIR}
mkdir -p ${ANONYBUILDDIR}
trap "rm -rf $ANONYBUILDDIR" EXIT 2 15

# make tar on every go files
pkgs=$(go list ./... )
dirlist=""
for x in `echo ${pkgs}`
do
	dir=$(echo $x | sed -e 's#github.com/bbklab/adbot/##g')
	dirlist="$dirlist ${dir}"
done
dirlist="$dirlist vendor"  # with vendor directory
tar -cf ${ANONYBUILDDIR}/.anony.tar $dirlist

# extract all of source files to build dir
tar -xf ${ANONYBUILDDIR}/.anony.tar -C ${ANONYBUILDDIR}

# modify every go files under build dir
gofiles=$(find ${ANONYBUILDDIR} -type f -iname "*.go")
for gf in `echo ${gofiles}`
do
	sed -i -e 's#github.com/bbklab/adbot#github.com/codingbot/adbot#g' $gf
	if [ $? -ne 0 ]; then
		echo -ERR $gf
		exit 1
	fi
done
echo +OK Modified Go Source Files!


# build
docker run --rm \
	--name buildadbot \
	-w /go/src/${PKG} \
	-e CGO_ENABLED=0 \
	-e GOOS=linux \
	-e GOCACHE=/go/cache \
	-v ${ANONYBUILDDIR}:/go/src/${PKG}:rw \
	-v ${CACHEDIR}:/go/cache/:rw \
	golang:1.10-alpine \
	sh -c "go build -ldflags \"${BUILD_FLAGS}\" -o bin/adbot ${PKG}/cmd/adbot"
echo +OK Binary Built!

# copy result out
cp -afv ${ANONYBUILDDIR}/bin/adbot  bin/adbot

# ensure binary file not contains keywords
keywords=(
  "bbklab"
  "zhang.elinks@gmail.com"
  "guangzheng"
  "bbklab.me"
  "bbklab.net"
)
for keyword in ${keywords[*]}
do
	if grep -E -q -i $keyword bin/adbot; then
		echo -ERR keyword: [$keyword] found in the binary
		exit 1
	fi
done

echo +OK Anonymous Binary Built!
