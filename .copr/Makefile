.PHONY: srpm

SPEC := shdoc-ng.spec

srpm:
ifndef VERSION
	$(error VERSION is required, e.g. make srpm VERSION=0.6.3 outdir=/tmp)
endif
	mkdir -p rpmbuild/SOURCES rpmbuild/SPECS
	tar czf rpmbuild/SOURCES/shdoc-ng-$(VERSION).tar.gz \
		--transform 's,^\.,shdoc-ng-$(VERSION),' \
		--exclude=.git --exclude=rpmbuild .
	rpmbuild -bs $(SPEC) \
		--define "_topdir $(CURDIR)/rpmbuild" \
		--define "_sourcedir $(CURDIR)/rpmbuild/SOURCES" \
		--define "version $(VERSION)"
	mv rpmbuild/SRPMS/*.src.rpm $(outdir)
