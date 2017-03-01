# Borrowed from: 
# https://github.com/silven/go-example/blob/master/Makefile
# https://vic.demuzere.be/articles/golang-makefile-crosscompile/
# https://ariejan.net/2015/10/03/a-makefile-for-golang-cli-tools/
# https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html

SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go' -maxdepth 1 | grep -v main.go | grep -v _test.go)
FILES = $(SOURCES)
BINARY = moresql
MAIN = cmds/moresql/main.go
DATE_COMPILED = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS_BASE = "-X main.version='$(shell git describe --abbrev=0 --tags --always)' -X main.BuildDate='$(DATE_COMPILED)' -X main.GitRef='$(shell git describe --tags --dirty --always)' -X main.GitSHA='$(shell git rev-parse --short HEAD)'"
LDFLAGS = -ldflags $(LDFLAGS_BASE)
# Symlink into GOPATH
GITHUB_USERNAME=zph
BUILD_DIR=${GOPATH}/src/github.com/${GITHUB_USERNAME}/${BINARY}
CURRENT_DIR=$(shell pwd)
BUILD_DIR_LINK=$(shell readlink ${BUILD_DIR})
GOARCH = amd64
.DEFAULT_GOAL := help

# Build the project
all: clean fmt test_full linux build docs

$(BINARY): $(FILES) $(MAIN) ## Build binary for current system architecture
	go build $(LDFLAGS) -o bin/$(BINARY) $(MAIN)

build: $(BINARY)

heroku: build ## Used by heroku build process

flags:
	@echo "$(LDFLAGS_BASE)"

test: ## Run tests
	go test -v

test_full: ## Test with race and coverage
	go test -v -race -cover

linux:
	GOOS=linux GOARCH=${GOARCH} go build $(LDFLAGS) -o bin/$(BINARY)-linux-${GOARCH} $(MAIN)

# darwin:
# 	cd ${BUILD_DIR}; \
# 	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o bin/${BINARY}-darwin-${GOARCH} . ; \
# 	cd - >/dev/null

fmt: ## Go fmt the code
	cd ${BUILD_DIR}; \
	go fmt $$(go list ./... | grep -v /vendor/) ; \
	cd - >/dev/null

clean: ## Clean out the generated binaries
	-rm -f bin/${BINARY}-*
	-rm -f bin/${BINARY}

docs: clean ## Regenerate README.md from template
	@./bin/update-readme
	@echo "If changes occured in README.md that you want in mkdocs run:"
	@echo "cp -f README.md docs/README.md"

docs-deploy:
	@git diff-index --quiet HEAD -- || (echo "Only allowed with clean working directory" && exit 1)
	@mkdocs gh-deploy

# Allows building whether in GOPATH or not
# link:
# 	BUILD_DIR=${BUILD_DIR}; \
# 	BUILD_DIR_LINK=${BUILD_DIR_LINK}; \
# 	CURRENT_DIR=${CURRENT_DIR}; \
# 	if [ "$${BUILD_DIR_LINK}" != "$${CURRENT_DIR}" ]; then \
# 	    echo "Fixing symlinks for build"; \
# 	    rm -f $${BUILD_DIR}; \
# 	    ln -s $${CURRENT_DIR} $${BUILD_DIR}; \
# 	fi

help: ## prints help
		@ cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_\-]*: *.*## *" | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: link linux darwin test fmt clean help
