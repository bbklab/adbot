Summary: 	agent of adbot
Name: 		adbot-agent
Version: 	{PRODUCT_VERSION}
Release: 	rhel7
License: 	Commercial
Group:  	Extension
Vendor:		BBKLAB
Packager: 	Guangzheng Zhang <zhang.elinks@gmail.com>
BuildRoot: 	/var/tmp/%{name}-%{version}-%{release}-root
Source0: 	%{name}-%{version}-%{release}.tgz
Requires(pre):		coreutils >= 8.22
Requires(post): 	coreutils >= 8.22
Requires(postun): 	coreutils >= 8.22

%description 
agent of adbot

%prep
%setup -q

%build

%install 
[ "$RPM_BUILD_ROOT" != "/" ] && [ -d $RPM_BUILD_ROOT ] && /bin/rm -rf $RPM_BUILD_ROOT

# /usr/bin/adbot
mkdir -p $RPM_BUILD_ROOT/usr/bin
cp -a bin/adbot $RPM_BUILD_ROOT/usr/bin/adbot-agent

# /etc/adbot/env
mkdir -p $RPM_BUILD_ROOT/etc/adbot
cp -a etc/agent.env.example $RPM_BUILD_ROOT/etc/adbot/agent.env.example

# adbot-agent.service
mkdir -p $RPM_BUILD_ROOT/usr/lib/systemd/system
cp -a systemd/adbot-agent.service $RPM_BUILD_ROOT/usr/lib/systemd/system/

%clean
[ "$RPM_BUILD_ROOT" != "/" ] && [ -d $RPM_BUILD_ROOT ] && /bin/rm -rf $RPM_BUILD_ROOT

%files
%defattr(-, root, root)
%attr(0755, root, root) /usr/bin/adbot-agent
%attr(0644, root, root) /etc/adbot/agent.env.example
%attr(0644, root, root) /usr/lib/systemd/system/adbot-agent.service

%config

%doc

%pre
:

%post
exec >/dev/null 2>&1
systemctl enable adbot-agent.service
systemctl daemon-reload
:

%preun
if [ "$1" == "0" ]; then	# if uninstall indeed
	systemctl stop adbot-agent
	if [ -e /etc/.adbot.uuid ]; then
		rm -f /etc/.adbot.uuid # remove adbot agent uuid
		rm -f /etc/.adbot-agent.shutdown
	fi
fi
:

%postun
exec >/dev/null 2>&1
systemctl disable adbot-agent.service
systemctl daemon-reload
:

%changelog
* Sat Dec  2 2017 Guangzheng Zhang <zhang.elinks@gmail.com>
- 0.1 rpm release
