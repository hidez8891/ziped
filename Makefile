NAME        := ziped
DESCRIPTION := zip editing tool
VERSION     := 1.0.1
REVISION    := $(shell git rev-parse --short HEAD)

LDFLAGS  := -ldflags="-s -w \
	-X \"main.name=$(NAME)\" \
	-X \"main.description=$(DESCRIPTION)\" \
	-X \"main.version=$(VERSION)-$(REVISION)\""

ifeq ($(OS),Windows_NT)
  TARGET := $(NAME).exe
  SRCS   := $(shell where.exe /r . '*.go')
else
  TARGET := $(NAME)
  SRCS   := $(shell find . -type f -name '*.go')
endif

$(TARGET):$(SRCS)
	go build $(LDFLAGS) -o $(TARGET)

.PHONY: install
install:
	go install $(LDFLAGS)

.PHONY: test
test:
	go test ./...
