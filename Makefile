BUILD_SCRIPT = ./scripts/build.sh
SERVER_APP_NAME = cmd/server/server
AGENT_APP_NAME = cmd/agent/agent
METRIC_TEST = ./test/metricstest-darwin-arm64

.PHONY: all build clean run

all: build

build:
	@echo "Building the server and agent binary"
	$(BUILD_SCRIPT)

clean:
	@echo "Cleaning up"
	@rm -f $(SERVER_APP_NAME) $(AGENT_APP_NAME)

run: build
	@echo "Running server"
	./$(SERVER_APP_NAME) &
	sleep 2
	@echo "Running agent"
	./$(AGENT_APP_NAME)

test: build
	@echo "Increment 1 test"
	$(METRIC_TEST) -test.v -test.run=^TestIteration1$ -binary-path=$(SERVER_APP_NAME) && fg
	sleep 2
	@echo "Increment 2 test"
	$(METRIC_TEST) -test.v -test.run=^TestIteration2[AB]*$ -source-path=. -agent-binary-path=cmd/agent/agent
	sleep 2
	@echo "Increment 3 test"
	$(METRIC_TEST) -test.v -test.run=^TestIteration3[AB]*$ -source-path=. -agent-binary-path=cmd/agent/agent -binary-path=cmd/server/server
	sleep 2
	@echo "Increment 4 test"
	SERVER_PORT=$(random unused-port)
	ADDRESS="localhost:${SERVER_PORT}"
	TEMP_FILE=$(random tempfile)
	$(METRIC_TEST) -test.v -test.run=^TestIteration4$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-server-port=$SERVER_PORT \
	-source-path=.
	sleep 2
	@echo "Increment 5 test"
	SERVER_PORT=$(random unused-port)
	ADDRESS="localhost:${SERVER_PORT}"
	TEMP_FILE=$(random tempfile)
	$(METRIC_TEST) -test.v -test.run=^TestIteration5$ \
	-agent-binary-path=cmd/agent/agent \
	-binary-path=cmd/server/server \
	-server-port=$SERVER_PORT \
	-source-path=.
