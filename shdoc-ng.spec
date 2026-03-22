%global goipath github.com/jdevera/shdoc-ng

Name:           shdoc-ng
Version:        @@VERSION@@
Release:        1%{?dist}
Summary:        Documentation generator for shell scripts

License:        MIT
URL:            https://%{goipath}
Source0:        %{name}-%{version}.tar.gz

BuildRequires:  golang >= 1.25

%description
shdoc-ng reads annotated shell scripts and produces Markdown, HTML, or JSON
documentation. It is a Go reimplementation of shdoc.

%prep
%autosetup -n %{name}-%{version}

%build
GOFLAGS=-mod=vendor go build -ldflags "-s -w -X main.version=%{version}" -o %{name} ./cmd/shdoc-ng

%install
install -Dpm 0755 %{name} %{buildroot}%{_bindir}/%{name}

%files
%license LICENSE
%{_bindir}/%{name}

%changelog
* Sun Mar 22 2026 Jacobo de Vera <73069+jdevera@users.noreply.github.com> - @@VERSION@@-1
- See https://github.com/jdevera/shdoc-ng/releases for release notes
