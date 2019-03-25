# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: all push container clean orphan all-push push-manifest annotate manifest

REGISTRY ?= staging-k8s.gcr.io
IMAGE = $(REGISTRY)/pause
IMAGE_WITH_OS_ARCH = $(IMAGE)-$(OS)-$(ARCH)

TAG = 3.1
REV = $(shell git describe --contains --always --match='v*')

# Architectures supported: amd64, arm, arm64, ppc64le and s390x
ARCH ?= amd64
# Operating systems supported: linux, windows
OS ?= linux

ALL_ARCH = amd64 arm arm64 ppc64le s390x
WINDOWS_ARCH = amd64

CFLAGS = -Os -Wall -Werror -static -DVERSION=v$(TAG)-$(REV)
KUBE_CROSS_IMAGE ?= k8s.gcr.io/kube-cross
KUBE_CROSS_VERSION ?= $(shell cat ../build-image/cross/VERSION)

LINUX-BIN = pause
LINUX-SRCS = linux/pause.c

WINDOWS-BIN = wincat
WINDOWS-SRCS = windows/wincat.go

# The manifest command is still experimental as of Docker 18.09.3
export DOCKER_CLI_EXPERIMENTAL=enabled

ifeq ($(ARCH),amd64)
	TRIPLE ?= x86_64-linux-gnu
endif

ifeq ($(ARCH),arm)
	TRIPLE ?= arm-linux-gnueabihf
endif

ifeq ($(ARCH),arm64)
	TRIPLE ?= aarch64-linux-gnu
endif

ifeq ($(ARCH),ppc64le)
	TRIPLE ?= powerpc64le-linux-gnu
endif

ifeq ($(ARCH),s390x)
	TRIPLE ?= s390x-linux-gnu
endif

# If you want to build AND push all containers, see the 'all-push' rule.
all: all-container

all-push: push-manifest

push-manifest: all-push-images all-annotate
	docker manifest push $(IMAGE):$(TAG)

sub-annotate-linux-%:
	$(MAKE) OS=linux ARCH=$* annotate
sub-annotate-windows-%:
	$(MAKE) OS=windows ARCH=$* annotate

sub-container-linux-%:
	$(MAKE) OS=linux ARCH=$* container
sub-container-windows-%:
	$(MAKE) OS=windows ARCH=$* container

sub-push-linux-%:
	$(MAKE) OS=linux ARCH=$* push
sub-push-windows-%:
	$(MAKE) OS=windows ARCH=$* push

all-annotate: $(addprefix sub-annotate-linux-,$(ALL_ARCH)) $(addprefix sub-annotate-windows-,$(WINDOWS_ARCH))
all-container: $(addprefix sub-container-linux-,$(ALL_ARCH)) $(addprefix sub-container-windows-,$(WINDOWS_ARCH))

all-push-images: $(addprefix sub-push-linux-,$(ALL_ARCH)) $(addprefix sub-push-windows-,$(WINDOWS_ARCH))

build: bin/$(LINUX-BIN)-linux-$(ARCH)

bin/$(LINUX-BIN)-linux-$(ARCH): $(LINUX-SRCS)
	mkdir -p bin
	docker run --rm -u $$(id -u):$$(id -g) -v $$(pwd):/build \
		$(KUBE_CROSS_IMAGE):$(KUBE_CROSS_VERSION) \
		/bin/bash -c "\
			cd /build && \
			$(TRIPLE)-gcc $(CFLAGS) -o $@ $^ && \
			$(TRIPLE)-strip $@"

bin/$(WINDOWS-BIN)-windows-$(ARCH): $(WINDOWS-SRCS)
	mkdir -p bin
	docker run --rm -e GOOS=windows -v $$(pwd):/build \
		$(KUBE_CROSS_IMAGE):$(KUBE_CROSS_VERSION) \
		/bin/bash -c "\
			cd /build && \
			go build -o $@ $^"

container: .container-$(OS)-$(ARCH)
.container-linux-$(ARCH): bin/$(LINUX-BIN)-linux-$(ARCH)
	docker build --pull -t $(IMAGE_WITH_OS_ARCH):$(TAG) --build-arg ARCH=$(ARCH) -f linux/Dockerfile .
	touch $@

.container-windows-$(ARCH): bin/$(WINDOWS-BIN)-windows-$(ARCH)
	mkdir -p bin
	DOCKER_HOST=${WINDOWS_DOCKER_HOST} \
		DOCKER_CERT_PATH=${WINDOWS_DOCKER_CERT_PATH} \
		DOCKER_TLS_VERIFY=${WINDOWS_DOCKER_TLS_VERIFY} \
		docker build --pull -t $(IMAGE_WITH_OS_ARCH):$(TAG) --build-arg ARCH=$(ARCH) -f windows/Dockerfile .
	touch $@

push: .push-$(OS)-$(ARCH)
.push-linux-$(ARCH): .container-$(OS)-$(ARCH)
	docker push $(IMAGE_WITH_OS_ARCH):$(TAG)
	touch $@
.push-windows-$(ARCH): .container-$(OS)-$(ARCH)
	DOCKER_HOST=${WINDOWS_DOCKER_HOST} \
		DOCKER_CERT_PATH=${WINDOWS_DOCKER_CERT_PATH} \
		DOCKER_TLS_VERIFY=${WINDOWS_DOCKER_TLS_VERIFY} \
		docker push $(IMAGE_WITH_OS_ARCH):$(TAG)
	touch $@

# Useful for testing, not automatically included in container image
orphan: bin/orphan-linux-$(ARCH)
bin/orphan-linux-$(ARCH): linux/orphan.c
	mkdir -p bin
	docker run -u $$(id -u):$$(id -g) -v $$(pwd):/build \
		$(KUBE_CROSS_IMAGE):$(KUBE_CROSS_VERSION) \
		/bin/bash -c "\
			cd /build && \
			$(TRIPLE)-gcc $(CFLAGS) -o $@ $^ && \
			$(TRIPLE)-strip $@"

annotate: .annotate-$(OS)-$(ARCH)
.annotate-$(OS)-$(ARCH): .container-$(OS)-$(ARCH) .push-$(OS)-$(ARCH) .manifest
	docker manifest create --amend $(IMAGE):$(TAG) $(IMAGE_WITH_OS_ARCH):$(TAG)
	docker manifest annotate $(IMAGE):$(TAG) $(IMAGE_WITH_OS_ARCH):$(TAG) --os $(OS) --arch $(ARCH)
	touch $@

manifest: .manifest
.manifest:
	docker manifest create $(IMAGE):$(TAG) $(IMAGE_WITH_OS_ARCH):$(TAG)
	touch $@

clean:
	rm -rf .container-* .push-* .annotate-* .manifest bin/
