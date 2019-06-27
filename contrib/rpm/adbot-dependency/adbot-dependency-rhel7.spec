Summary: 	dependency packages of adbot master
Name: 		adbot-dependency
Version: 	1.0.0
Release: 	rhel7
License: 	Commercial
Group:  	Extension
Vendor:		Coding Bot
Packager: 	Coding Bot <codingbot@gmail.com>
BuildRoot: 	/var/tmp/%{name}-%{version}-%{release}-root
Source0: 	%{name}-%{version}-%{release}.tgz
Requires(pre):		coreutils >= 8.22
Requires(post): 	coreutils >= 8.22
Requires(postun): 	coreutils >= 8.22

%description 
dependency packages of adbot master

%prep
%setup -q

%build

%install 
[ "$RPM_BUILD_ROOT" != "/" ] && [ -d $RPM_BUILD_ROOT ] && /bin/rm -rf $RPM_BUILD_ROOT

# /usr/share/adbot/dependency
mkdir -p $RPM_BUILD_ROOT/usr/share/adbot/dependency
cp -a share/dependency/*    $RPM_BUILD_ROOT/usr/share/adbot/dependency

%clean
[ "$RPM_BUILD_ROOT" != "/" ] && [ -d $RPM_BUILD_ROOT ] && /bin/rm -rf $RPM_BUILD_ROOT

%files
%attr(0755, root, root) /usr/share/adbot/dependency/
%attr(0644, root, root)/usr/share/adbot/dependency/*

%config

%doc

%pre

%post

%preun

%postun

%changelog
* Sat Dec  2 2017 Coding Bot <codingbot@gmail.com>
- 1.0.0 rpm release
