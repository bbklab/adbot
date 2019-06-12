#!/bin/bash
#
# this script will be running in a temporary rpm building container
# which is built from `Dockerfile.buildrpm`
#
set -ex

# def
base=$(cd $(dirname ${BASH_SOURCE[*]}); pwd -P)
spec="${base}/adbot-dependency-rhel7.spec"
rpmbuild="/usr/bin/rpmbuild"
product="/product"

clean() {
	rm -rf "${base}"/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS} 2>&-
}

prepare() {
	test -e share/dependency/mongod.pkg -a -r share/dependency/mongod.pkg
	test -e share/dependency/prometheus.pkg -a -r share/dependency/prometheus.pkg 
	mkdir -p $product
}


build() {
	local specfile=$1

	if [ ! -f "${specfile}" -o ! -s "${specfile}" ]; then
		echo "spec file: ${specfile} not prepared."
		return 1
	fi

	mkdir -p "${base}"/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS} 2>&-

	local name="$(awk '($1~/^Name:/){print $2;exit}' ${specfile} 2>&-)"
	local version="$(awk '($1~/^Version:/){print $2;exit}' ${specfile} 2>&-)"
	local release="$(awk '($1~/^Release:/){print $2;exit}' ${specfile} 2>&-)"
	local tgzname="${name}-${version}-${release}.tgz"

	pushd "${base}/SOURCES/"
	mkdir -p "${name}-${version}"
	cp -a ${base}/share ${name}-${version}
	tar -c --remove-files -zf "$tgzname" ${name}-${version}
	popd

	cp -a "${specfile}" "${base}/SPECS/${name}.spec"

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
