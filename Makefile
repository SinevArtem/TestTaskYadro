
BINARY_NAME=app
CONFIG_PATH=./config/config.json
EVENTS_PATH=./events

# CONFIG_PATH=./common/test_files/test_config.json
# EVENTS_PATH=./common/test_files/events_test_3

.PHONY: all build run test test-cover clean

all: run

build:
	go build -o $(BINARY_NAME) cmd/app/main.go

run: build
	./$(BINARY_NAME) --config=$(CONFIG_PATH) < $(EVENTS_PATH)

test:
	go test ./... 

test-cover:
	go test ./... -cover 


clean:
	rm -f $(BINARY_NAME)
