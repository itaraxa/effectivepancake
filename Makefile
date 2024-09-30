BUILD_SCRIPT = ./scripts/build.sh
SERVER_APP_NAME = cmd/server/server
AGENT_APP_NAME = cmd/agent/agent

.PHONY: all build clean run

all: build

build:
	@echo "Building the server"
	$(BUILD_SCRIPT)

clean:
	@echo "Cleaning up"
	@rm -f $(SERVER_APP_NAME) $(AGENT_APP_NAME)

run: build
	@echo "Running server"
	./$(SERVER_APP_NAME)
	sleep 2
	./$(AGENT_APP_NAME)
