BUILD_SCRIPT = ./scripts/build.sh
APP_NAME = cmd/server/server

.PHONY: all build clean run

all: build

build:
	@echo "Building the server"
	$(BUILD_SCRIPT)

clean:
	@echo "Cleaning up"
	@rm -f $(APP_NAME)

run: build
	@echo "Running server"
	./$(APP_NAME)
