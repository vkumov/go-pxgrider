GOPATH := $(shell go env GOPATH)
PATH := $(PATH):$(GOPATH)/bin
PROTO_PREFIX := github.com/vkumov/go-pxgrider/pxgrider_proto
BOIL_VER := v4.18.0
BOIL_EXT_VER := v0.9.0
SSH_PRIVATE_KEY := $(shell cat ~/.ssh/id_rsa | base64)

BUILD_STAMP := $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
GIT_HASH := $(shell git rev-parse HEAD)
V := $(shell git describe --always --tags --dirty)

gen-proto:
	protoc --go_out=./pkg --go_opt=module=$(PROTO_PREFIX) \
    --go-grpc_out=./pkg --go-grpc_opt=module=$(PROTO_PREFIX) ./proto/*.proto

prepare:
	@go get -u github.com/tiendc/sqlboiler-extensions@$(BOIL_EXT_VER)
	@go install github.com/volatiletech/sqlboiler/v4@$(BOIL_VER)
	@go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@$(BOIL_VER)
	@go install github.com/glerchundi/sqlboiler-crdb/v4@latest

wipe-models:
	@sqlboiler psql --wipe

gen-models:
	@sqlboiler psql -c sqlboiler.toml \
 		--templates $(GOPATH)/pkg/mod/github.com/volatiletech/sqlboiler/v4@$(BOIL_VER)/templates/main \
 		--templates $(GOPATH)/pkg/mod/github.com/tiendc/sqlboiler-extensions@$(BOIL_EXT_VER)/templates/boilv4/postgres

build:
	ENV=production go build --trimpath -ldflags "\
    -X github.com/vkumov/go-pxgrider/server/internal/config.BuildStamp=`date -u '+%Y-%m-%d_%I:%M:%S%p'` \
    -X github.com/vkumov/go-pxgrider/server/internal/config.GitHash=`git rev-parse HEAD` \
    -X github.com/vkumov/go-pxgrider/server/internal/config.V=`git describe --always --tags --dirty`\
  " -buildvcs=false -o ./bin/pxgrider ./server/bin/*.go

release-proto:
	git tag -a pkg/v$(VERSION) -m "Release proto v$(VERSION)"
	git push origin master pkg/v$(VERSION)
	go get github.com/vkumov/go-pxgrider/pkg@v$(VERSION)

dev:
	air

build-docker:
	docker build -t pxgrider \
		--build-arg="BUILD_STAMP=$(BUILD_STAMP)" \
		--build-arg="GIT_HASH=$(GIT_HASH)" \
		--build-arg="VERSION_TAG=$(V)" .

run-docker-dev:
	cd ./docker-compose && docker-compose up

run-docker-detach-dev:
	cd ./docker-compose && docker-compose ud -d