# Compiler
GO := go

# Build directory
BIN_DIR := bin

# Detect current OS
ifeq ($(OS),Windows_NT)
    CURRENT_OS := windows
    EXE_EXT := .exe
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
        CURRENT_OS := linux
        EXE_EXT :=
    else ifeq ($(UNAME_S),Darwin)
        CURRENT_OS := darwin
        EXE_EXT :=
    else ifeq ($(UNAME_S),FreeBSD)
        CURRENT_OS := freebsd
        EXE_EXT :=
    else ifeq ($(UNAME_S),OpenBSD)
        CURRENT_OS := openbsd
        EXE_EXT :=
    else ifeq ($(UNAME_S),NetBSD)
        CURRENT_OS := netbsd
        EXE_EXT :=
    else
        CURRENT_OS := unknown
        EXE_EXT :=
    endif
endif

# Detect current ARCH
CURRENT_ARCH ?= $(shell go env GOARCH)

# Target OS and ARCH (default to current)
TARGET_OS ?= $(CURRENT_OS)
TARGET_ARCH ?= $(CURRENT_ARCH)

# Set executable extension based on target OS
ifeq ($(TARGET_OS),windows)
    TARGET_EXT := .exe
else
    TARGET_EXT :=
endif

# Map interval names to config constants
INTERVAL_MAP := \
	loader-1min=CANDLE_INTERVAL_1_MIN \
	loader-2min=CANDLE_INTERVAL_2_MIN \
	loader-3min=CANDLE_INTERVAL_3_MIN \
	loader-5min=CANDLE_INTERVAL_5_MIN \
	loader-10min=CANDLE_INTERVAL_10_MIN \
	loader-15min=CANDLE_INTERVAL_15_MIN \
	loader-30min=CANDLE_INTERVAL_30_MIN \
	loader-1hour=CANDLE_INTERVAL_HOUR \
	loader-2hour=CANDLE_INTERVAL_2_HOUR \
	loader-4hour=CANDLE_INTERVAL_4_HOUR \
	loader-1day=CANDLE_INTERVAL_DAY \
	loader-1week=CANDLE_INTERVAL_WEEK \
	loader-1month=CANDLE_INTERVAL_MONTH

# All interval-based loaders
INTERVAL_LOADERS := loader-1min loader-2min loader-3min loader-5min \
                    loader-10min loader-15min loader-30min \
                    loader-1hour loader-2hour loader-4hour \
                    loader-1day loader-1week loader-1month

# Other loaders (not interval-based)
OTHER_LOADERS := loader-instruments loader-dividends loader-arch loader-cli

# Default target
.PHONY: all
all: build

# Create bin directory
$(BIN_DIR):
	mkdir -p $(BIN_DIR)

# Build all loaders
.PHONY: build
build: $(BIN_DIR)
	@echo "Building interval loaders for $(TARGET_OS)/$(TARGET_ARCH)..."
	@for loader in $(INTERVAL_LOADERS); do \
		interval=$$(echo "$(INTERVAL_MAP)" | tr ' ' '\n' | grep "^$$loader=" | cut -d= -f2); \
		if [ -z "$$interval" ]; then \
			echo " Неизвестный интервал для $$loader"; \
			exit 1; \
		fi; \
		echo " Building $$loader ($$interval)..."; \
		GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) $(GO) build \
			-ldflags "-X main.MAININTERVAL=$$interval" \
			-o $(BIN_DIR)/$$loader$(TARGET_EXT) \
			cmd/loader-interval/main.go || exit 1; \
	done
	@echo ""
	@echo "Building other loaders..."
	@for loader in $(OTHER_LOADERS); do \
		echo " Building $$loader..."; \
		GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) $(GO) build \
			-o $(BIN_DIR)/$$loader$(TARGET_EXT) \
			cmd/$$loader/main.go || exit 1; \
	done
	@echo ""
	@echo "Build completed. Executables are in $(BIN_DIR)/"

# Build individual interval loader
.PHONY: build-%
build-%: $(BIN_DIR)
	@echo "Building $* for $(TARGET_OS)/$(TARGET_ARCH)..."
	@loader="$*"; \
	if echo "$(INTERVAL_LOADERS)" | tr ' ' '\n' | grep -q "^$$loader$$"; then \
		interval=$$(echo "$(INTERVAL_MAP)" | tr ' ' '\n' | grep "^$$loader=" | cut -d= -f2); \
		GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) $(GO) build \
			-ldflags "-X main.MAININTERVAL=$$interval" \
			-o $(BIN_DIR)/$$loader$(TARGET_EXT) \
			cmd/loader-interval/main.go; \
	elif echo "$(OTHER_LOADERS)" | tr ' ' '\n' | grep -q "^$$loader$$"; then \
		GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) $(GO) build \
			-o $(BIN_DIR)/$$loader$(TARGET_EXT) \
			cmd/$$loader/main.go; \
	else \
		echo "Неизвестный загрузчик: $$loader"; \
		exit 1; \
	fi

# Cross-compilation OS targets
.PHONY: build-windows
build-windows:
	$(MAKE) TARGET_OS=windows TARGET_ARCH=amd64 build

.PHONY: build-linux
build-linux:
	$(MAKE) TARGET_OS=linux TARGET_ARCH=amd64 build

.PHONY: build-darwin
build-darwin:
	$(MAKE) TARGET_OS=darwin TARGET_ARCH=amd64 build

.PHONY: build-darwin-arm64
build-darwin-arm64:
	$(MAKE) TARGET_OS=darwin TARGET_ARCH=arm64 build

.PHONY: build-freebsd
build-freebsd:
	$(MAKE) TARGET_OS=freebsd TARGET_ARCH=amd64 build

.PHONY: build-openbsd
build-openbsd:
	$(MAKE) TARGET_OS=openbsd TARGET_ARCH=amd64 build

.PHONY: build-netbsd
build-netbsd:
	$(MAKE) TARGET_OS=netbsd TARGET_ARCH=amd64 build

# Individual cross-compilation
.PHONY: build-windows-%
build-windows-%:
	$(MAKE) TARGET_OS=windows TARGET_ARCH=amd64 build-$*

.PHONY: build-linux-%
build-linux-%:
	$(MAKE) TARGET_OS=linux TARGET_ARCH=amd64 build-$*

.PHONY: build-darwin-%
build-darwin-%:
	$(MAKE) TARGET_OS=darwin TARGET_ARCH=amd64 build-$*

.PHONY: build-darwin-arm64-%
build-darwin-arm64-%:
	$(MAKE) TARGET_OS=darwin TARGET_ARCH=arm64 build-$*

.PHONY: build-freebsd-%
build-freebsd-%:
	$(MAKE) TARGET_OS=freebsd TARGET_ARCH=amd64 build-$*

.PHONY: build-openbsd-%
build-openbsd-%:
	$(MAKE) TARGET_OS=openbsd TARGET_ARCH=amd64 build-$*

.PHONY: build-netbsd-%
build-netbsd-%:
	$(MAKE) TARGET_OS=netbsd TARGET_ARCH=amd64 build-$*

# Clean
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR)
	@echo "Clean completed."

# Lint
.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	golangci-lint run
	@echo "Linting completed."

# Help
.PHONY: help
help:
	@echo ""
	@echo "Available targets:"
	@echo ""
	@echo "  all                         - Build all loaders for current OS/ARCH"
	@echo "  build                       - Same as 'all'"
	@echo ""
	@echo "  build-NAME                  - Build specific loader (e.g. build-loader-1min)"
	@echo ""
	@echo "  build-windows               - Build all for Windows (amd64)"
	@echo "  build-linux                 - Build all for Linux (amd64)"
	@echo "  build-darwin                - Build all for macOS Intel (amd64)"
	@echo "  build-darwin-arm64          - Build all for macOS Apple Silicon (arm64)"
	@echo "  build-freebsd               - Build all for FreeBSD (amd64)"
	@echo "  build-openbsd               - Build all for OpenBSD (amd64)"
	@echo "  build-netbsd                - Build all for NetBSD (amd64)"
	@echo ""
	@echo "  build-windows-NAME          - e.g. build-windows-loader-1min"
	@echo "  build-linux-NAME            - e.g. build-linux-loader-1day"
	@echo "  build-darwin-NAME           - macOS Intel"
	@echo "  build-darwin-arm64-NAME     - macOS Apple Silicon"
	@echo ""
	@echo "  clean                       - Remove bin/ directory"
	@echo "  lint                        - Run golangci-lint"
	@echo "  help                        - Show this message"
	@echo ""
	@echo "Current: OS=$(CURRENT_OS), ARCH=$(CURRENT_ARCH)"
	@echo "Target:  OS=$(TARGET_OS), ARCH=$(TARGET_ARCH)"
	@echo ""
	@echo "Examples:"
	@echo "  make"
	@echo "  make build-loader-15min"
	@echo "  make build-darwin-arm64"
	@echo "  make build-linux-loader-1hour"
	@echo ""