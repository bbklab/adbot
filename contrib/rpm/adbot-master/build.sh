#!/bin/bash
#
# this script will be running in a temporary rpm building container
# which is built from `Dockerfile.buildrpm`
#
set -ex

# def
base=$(cd $(dirname ${BASH_SOURCE[*]}); pwd -P)
spec="${base}/adbot-master-rhel7.spec"
rpmbuild="/usr/bin/rpmbuild"
product="/product"
version=""

clean() {
	rm -rf "${base}"/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS} 2>&-
}

prepare() {
	test -e bin/adbot -a -x bin/adbot
	test -e systemd/adbot-master.service -a -r systemd/adbot-master.service
	test -e etc/master.env.example -a -r etc/master.env.example

	mkdir -p $product

	version=$(bin/adbot version 2>&- | awk '/Version:/{print $2;exit;}' | sed -e 's/-/./g')
	if [ "${version}" == "" ]; then
		echo "adbot binary version not detected, abort"
		exit 1
	fi
}


build() {
	local specfile=$1

	if [ ! -f "${specfile}" -o ! -s "${specfile}" ]; then
		echo "spec file: ${specfile} not prepared."
		return 1
	fi

	mkdir -p "${base}"/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS} 2>&-

	local name="$(awk '($1~/^Name:/){print $2;exit}' ${specfile} 2>&-)"
	local release="$(awk '($1~/^Release:/){print $2;exit}' ${specfile} 2>&-)"
	local tgzname="${name}-${version}-${release}.tgz"

	pushd "${base}/SOURCES/"
	mkdir -p "${name}-${version}"
	cp -a ${base}/{bin,systemd,etc} ${name}-${version}
	if [[ "${specfile}" =~ master ]]; then
		cp -a ${base}/share ${name}-${version}
	fi
	tar -c --remove-files -zf "$tgzname" ${name}-${version}
	popd

	# note: keep original spec file unchanged and generate 
	# new spec file to the right place 
	sed -e  's/{PRODUCT_VERSION}/'$version'/g' $specfile > "${base}/SPECS/${name}.spec"  

	cat > ~/.rpmmacros <<EOF
%_topdir ${base}/
%debug_package %{nil}
EOF

	# time to glue all of them together   
	$rpmbuild -bb "${base}/SPECS/${name}.spec"
	mv -fv $base/RPMS/$(uname -m)/${name}-${version}-${release}.$(uname -m).rpm  $product

	clean
}

#
# main
#
if [ ! -f $rpmbuild -o ! -x $rpmbuild ]; then
	echo "/usr/bin/rpmbuild not prepared"
	exit 1
fi

if [ "${IN_CONTAINER}" != "yes" ]; then
	echo "intended to be running in a container"
	exit 1
fi

pushd $base
prepare
build $spec
