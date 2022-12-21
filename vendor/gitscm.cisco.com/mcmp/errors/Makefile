MAKEFLAGS += -r --warn-undefined-variables
SHELL := /bin/bash
.SHELLFLAGS := -o pipefail -euc
.DEFAULT_GOAL := help

include Makefile.variables
include Makefile.local

.PHONY: help clean veryclean dockerclean tag-build vendor format check test race cover todo adhoc next-dev start-release finish-release

## display this help message
help:
	@echo 'Management commands for errors:'
	@echo
	@echo 'Usage:'
	@echo '  ## Build Commands'
	@echo '    tag-build       Add git tag for latest build.'
	@echo
	@echo '  ## Develop / Test Commands'
	@echo '    vendor          Install dependencies using go mod if go.mod changed.'
	@echo '    format          Run code formatter.'
	@echo '    check           Run static code analysis (lint).'
	@echo '    test            Run tests on project.'
	@echo '    race            Run tests on project with race condition detector enabled.'
	@echo '    cover           Run tests and capture code coverage metrics on project.'
	@echo '    todo            Generate a TODO list for project.'
	@echo '    clean           Clean the directory tree of produced artifacts.'
	@echo '    veryclean       Same as clean but also removes cached dependencies.'
	@echo '    dockerclean     Same as veryclean but also performs Docker environment cleanup.'
	@echo
	@echo '  ## Release Commands'
	@echo '    next-dev        Prepare project for development of next version.'
	@echo '    start-release   Start release of version of project.'
	@echo '    finish-release  Finish release of version of project.'
	@echo
	@echo '  ## Local Commands'
	@echo '    drma            Removes all stopped containers.'
	@echo '    drmia           Removes all unlabelled images.'
	@echo '    drmvu           Removes all unused container volumes.'
	@echo

.ci-clean:
ifeq ($(CI_ENABLED),1)
	@rm -f tmp/dev_image_id || :
endif

## Clean the directory tree of produced artifacts.
clean: .ci-clean prepare
	@${DOCKERRUN} bash -c 'rm -rf cover *.out *.xml'

## Same as clean but also removes cached dependencies.
veryclean: clean
	@${DOCKERRUN} bash -c 'rm -rf tmp vendor .mod'

## Same as veryclean but also performs Docker environment cleanup.
dockerclean: veryclean drma drmia drmvu

## builds the dev container
prepare: tmp/dev_image_id
tmp/dev_image_id: Dockerfile.dev
	@mkdir -p tmp
	@docker rmi -f ${DEV_IMAGE} > /dev/null 2>&1 || :
	@echo "## Building dev container"
	@docker build --quiet -t ${DEV_IMAGE} --build-arg DEVELOPER="${DEVELOPER}" -f Dockerfile.dev .
	@docker inspect -f "{{ .ID }}" ${DEV_IMAGE} > tmp/dev_image_id

# ----------------------------------------------
# build

## Add git tag for latest build.
tag-build: prepare
	${DOCKERRUN} bash ./scripts/version.sh build

# ----------------------------------------------
# dependencies

## Install dependencies using go mod if go.mod changed.
vendor: tmp/vendor-installed
tmp/vendor-installed: tmp/dev_image_id go.mod
	@mkdir -p .mod
	${DOCKERRUN} go mod tidy
	@date > tmp/vendor-installed
	@chmod 644 go.sum || :

# ----------------------------------------------
# develop and test

## print environment info about this dev environment
debug:
	@echo IMPORT_PATH="$(IMPORT_PATH)"
	@echo ROOT="$(ROOT)"
	@echo VERSION="$(VERSION)"
	@echo PRERELEASE="$(PRERELEASE)"
	@echo RELEASE_TYPE="$(RELEASE_TYPE)"
	@echo TAG_TYPE="$(TAG_TYPE)"
	@echo
	@echo docker commands run as:
	@echo "$(DOCKERRUN)"

## Run code formatter.
format: tmp/vendor-installed
	${DOCKERRUN} bash ./scripts/format.sh
	@if [[ -n "$$(git -c core.fileMode=false status --porcelain)" ]]; then \
		echo -e "\n\tgoimports modified code; requires attention!\n" ; \
		if [[ "${CI_ENABLED}" == "1" ]]; then \
			git status --short ; echo "" ; \
			exit 1 ; \
		fi ; \
	fi

## Run static code analysis (lint).
check: format
ifeq ($(CI_ENABLED),1)
	${DOCKERRUN} bash ./scripts/check.sh --jenkins
else
	${DOCKERRUN} bash ./scripts/check.sh
endif

## Run tests on project.
test: check
	${DOCKERRUN} bash ./scripts/test.sh

## Run tests on project with race condition detector enabled.
race: check
	${DOCKERRUN} bash ./scripts/test.sh --race

## Run tests and capture code coverage metrics on project.
cover: check
	@rm -rf cover/
	@mkdir -p cover
ifeq ($(CI_ENABLED),1)
	${DOCKERRUN} bash ./scripts/cover.sh --jenkins
else
	${DOCKERRUN} bash ./scripts/cover.sh
	@chmod 644 cover/coverage.html || :
endif

# generate a TODO.md file with a list of TODO and FIXME items sorted by file
# the string is case insensitive and is removed from the output. So the final output
# should provide the file, line number, username that added it, and message about what
# needs to be done.
# Excludes the Makefile from consideration. Only files that are being tracked in git are
# included by default, therefore external dependencies or anything that is part of gitignore
# is automatically excluded.
## Generate a TODO list for project.
todo: prepare
	${DOCKERRUN} bash ./scripts/todo.sh -e Makefile -e scripts/todo.sh -t '(FIXME|TODO)'

# usage: make adhoc RUNTHIS='command to run inside of dev container'
# example: make adhoc RUNTHIS='which git'
adhoc: prepare
	@${DOCKERRUN} ${RUNTHIS}

# ----------------------------------------------
# release

## Prepare project for development of next version.
next-dev: prepare
	@echo "## Incrementing version: v${VERSION}"
	${DOCKERRUN} bash ./scripts/version.sh dev

## Start release of version of project.
start-release: .pre-rel prepare
	@echo "## Starting release for version: v${VERSION}"
	@git flow release start "v${VERSION}"
	${DOCKERRUN} bash ./scripts/version.sh rel

## Finish release of version of project.
finish-release: .pre-rel
	@echo "## Releasing version: v${VERSION}"
	@git flow release finish -p -m "Release v${VERSION}" "v${VERSION}"

.pre-rel:
ifndef HAS_GITFLOW
	$(error "You must install git-flow")
endif
