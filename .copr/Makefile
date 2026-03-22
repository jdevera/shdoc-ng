.PHONY: srpm

# Extract version from the latest git tag (strips leading 'v')
VERSION := $(shell git describe --tags --abbrev=0 | sed 's/^v//')
SPEC := shdoc-ng.spec

srpm:
	mkdir -p rpmbuild/{SOURCES,SPECS}
	git archive --format=tar.gz --prefix=shdoc-ng-$(VERSION)/ HEAD > rpmbuild/SOURCES/shdoc-ng-$(VERSION).tar.gz
	sed 's/%{version}/$(VERSION)/' $(SPEC) > rpmbuild/SPECS/$(SPEC)
	rpmbuild -bs rpmbuild/SPECS/$(SPEC) \
		--define "_topdir $(CURDIR)/rpmbuild" \
		--define "_sourcedir $(CURDIR)/rpmbuild/SOURCES"
	mv rpmbuild/SRPMS/*.src.rpm $(outdir)
