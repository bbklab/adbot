Summary: 	geolite2 data of adbot master
Name: 		adbot-geolite2
Version: 	1.0.0
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
geolite2 data of adbot master

%prep
%setup -q

%build

%install 
[ "$RPM_BUILD_ROOT" != "/" ] && [ -d $RPM_BUILD_ROOT ] && /bin/rm -rf $RPM_BUILD_ROOT

# /usr/share/adbot/geo
mkdir -p $RPM_BUILD_ROOT/usr/share/adbot/geo
cp -a share/geo/*    $RPM_BUILD_ROOT/usr/share/adbot/geo

%clean
[ "$RPM_BUILD_ROOT" != "/" ] && [ -d $RPM_BUILD_ROOT ] && /bin/rm -rf $RPM_BUILD_ROOT

%files
%attr(0755, root, root) /usr/share/adbot/geo/
%attr(0644, root, root)/usr/share/adbot/geo/*

%config

%doc

%pre

%post

%preun

%postun

%changelog
* Sat Dec  2 2017 Guangzheng Zhang <zhang.elinks@gmail.com>
- 1.0.0 rpm release
