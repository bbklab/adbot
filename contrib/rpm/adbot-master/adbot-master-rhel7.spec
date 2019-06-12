Summary: 	master of adbot
Name: 		adbot-master
Version: 	{PRODUCT_VERSION}
Release: 	rhel7
License: 	Commercial
Group:  	Extension
Vendor:		BBKLAB
Packager: 	Guangzheng Zhang <zhang.elinks@gmail.com>
BuildRoot: 	/var/tmp/%{name}-%{version}-%{release}-root
Source0: 	%{name}-%{version}-%{release}.tgz
Requires(pre):		coreutils >= 8.22, adbot-dependency >= 1.0.0, adbot-geolite2 >= 1.0.0
Requires(post): 	coreutils >= 8.22
Requires(postun): 	coreutils >= 8.22

%description 
master of adbot

%prep
%setup -q

%build

%install 
[ "$RPM_BUILD_ROOT" != "/" ] && [ -d $RPM_BUILD_ROOT ] && /bin/rm -rf $RPM_BUILD_ROOT

# /usr/bin/adbot
mkdir -p $RPM_BUILD_ROOT/usr/bin
cp -a bin/adbot    $RPM_BUILD_ROOT/usr/bin/adbot

# /etc/adbot/env
mkdir -p $RPM_BUILD_ROOT/etc/adbot
cp -a etc/master.env.example  $RPM_BUILD_ROOT/etc/adbot/master.env.example
# /etc/adbot/node_hooks
mkdir -p $RPM_BUILD_ROOT/etc/adbot/node_hooks
mkdir -p $RPM_BUILD_ROOT/etc/adbot/node_hooks/post_install
mkdir -p $RPM_BUILD_ROOT/etc/adbot/node_hooks/pre_uninstall
# product public keys
mkdir -p $RPM_BUILD_ROOT/etc/adbot/keys
cp -a etc/keys/public.key.pem $RPM_BUILD_ROOT/etc/adbot/keys/public.key.pem

# /usr/share/adbot/
mkdir -p $RPM_BUILD_ROOT/usr/share/adbot
cp -a share/*    $RPM_BUILD_ROOT/usr/share/adbot/

# adbot-master.service
mkdir -p $RPM_BUILD_ROOT/usr/lib/systemd/system
cp -a systemd/adbot-master.service $RPM_BUILD_ROOT/usr/lib/systemd/system/

%clean
[ "$RPM_BUILD_ROOT" != "/" ] && [ -d $RPM_BUILD_ROOT ] && /bin/rm -rf $RPM_BUILD_ROOT

%files
%defattr(-, root, root)
%attr(0755, root, root) /usr/bin/adbot
%attr(0644, root, root) /etc/adbot/master.env.example
%attr(0755, root, root) /etc/adbot/node_hooks/post_install/
%attr(0755, root, root) /etc/adbot/node_hooks/pre_uninstall/
%attr(0755, root, root) /etc/adbot/keys/
%attr(0644, root, root) /etc/adbot/keys/public.key.pem
%attr(-, root, root)    /usr/share/adbot/*
%attr(0644, root, root) /usr/lib/systemd/system/adbot-master.service

%config

%doc

%pre
:

%post
exec >/dev/null 2>&1
systemctl enable adbot-master.service
systemctl daemon-reload
:

%preun
if [ "$1" == "0" ]; then	# if uninstall indeed
	systemctl stop adbot-master
fi
:

%postun
exec >/dev/null 2>&1
systemctl disable adbot-master.service
systemctl daemon-reload
:

%changelog
* Sat Dec  2 2017 Guangzheng Zhang <zhang.elinks@gmail.com>
- 0.1 rpm release
