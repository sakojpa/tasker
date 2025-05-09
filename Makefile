GOOS		:= $(shell go env GOOS)
GOARCH		:= $(shell go env GOARCH)
GO			:= $(shell which go)
GO_TAGS		?=
LDFLAGS		?=


BUILD_DIR	?= ./build

DOCKER_RUN_OPTS		?=
DOCKER_BUILD_OPTS	?=
REGISTRY			?= docker.io
IMAGE				?= sakojpa/tasker
TAG					?= $(shell git describe --tags --always)

RUN_PORT		?= 7540
RUN_DBFILE		?= scheduler.db
RUN_PASSWORD	?= 123

TEST_FULL		?= false
TEST_SEARCH		?= false

.SILENT:
.PHONY: build docker run run-with-auth run-n-test test-with-auth

ifeq ($(strip $(DOCKER_RUN_OPTS)),)
PORT_OPTION = -p 7540:7540
else
PORT_OPTION =
endif

build:
	$(GO) build -tags $(GO_TAGS) -ldflags $(LDFLAGS) -o $(BUILD_DIR)/tasker main.go

build-docker:
	docker buildx build \
		$(DOCKER_BUILD_OPTS) \
		-t $(IMAGE):$(TAG) \
		.

run-docker: build-docker
	docker run -d --rm \
		$(DOCKER_RUN_OPTS) \
		$(PORT_OPTION) \
		-t $(IMAGE):$(TAG) \
		.

run:
	TODO_PORT=$(RUN_PORT) \
	TODO_DBFILE=$(RUN_DBFILE) \
	$(GO) run main.go

run-with-auth:
	TODO_PASSWORD=$(RUN_PASSWORD) \
	TODO_PORT=$(RUN_PORT) \
	TODO_DBFILE=$(RUN_DBFILE) \
	$(GO) run main.go

PID_FILE = .app.pid
run-n-test:
	$(GO) run main.go > /dev/null 2>&1 & echo $$! > $(PID_FILE)
	sleep 1
	TODO_FULLNEXTDATE=$(TEST_FULL) TODO_SEARCH=$(TEST_SEARCH) $(GO) test -v ./tests -count=1
	pkill -KILL -P $$(cat $(PID_FILE))

run-n-test-with-auth:
	TODO_PASSWORD=$(RUN_PASSWORD) \
	$(GO) run main.go  > /dev/null 2>&1 & echo $$! > $(PID_FILE)
	sleep 1
	TODO_TOKEN=$$(curl -s 'http://localhost:7540/api/signin' -H 'Content-Type: application/json' -d '{"password":"123"}' | jq -r '.token') \
	TODO_FULLNEXTDATE=true TODO_SEARCH=true $(GO) test -v ./tests -count=1
	pkill -KILL -P $$(cat $(PID_FILE))
