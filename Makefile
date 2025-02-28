API_BINARY_PATH = bin/api
FIREWALL_BINARY_PATH = bin/firewall
PROTO_FILE = internal/shared/protocol/protocol.proto
PROTO_GO = internal/shared/protocol/protocol.pb.go
OPENSNITCH_PROTO_FILE = internal/shared/opensnitch/opensnitch.proto

# Protocol buffer generation
.PHONY: protoc
protoc: $(PROTO_GO)

$(PROTO_GO): $(PROTO_FILE)
	@echo "Generating protocol buffers..."
	@protoc --go_out=. --go_opt=paths=source_relative $<

# Client targets
.PHONY: api
api: protoc
		@echo "Building api..."
		@mkdir -p bin
		@go build -o $(API_BINARY_PATH) ./cmd/api

.PHONY: run-api
run-api: api
		@echo "Running api..."
		@./$(API_BINARY_PATH)

# Linux Agent targets
.PHONY: firewall
firewall: protoc
		@echo "Building firewall..."
		@mkdir -p bin
		@go build -o $(FIREWALL_BINARY_PATH) ./cmd/firewall

.PHONY: run-firewall
run-firewall: firewall
		@echo "Running firewall..."
		@./$(FIREWALL_BINARY_PATH)

.PHONY: clean
clean:
		@echo "Cleaning up..."
		@rm *.sqlite3