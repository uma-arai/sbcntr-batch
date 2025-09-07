.PHONY: all build test clean validate install-tools

# ビルド後の出力先ディレクトリ
BUILD_DIR     = bin

# バッチ処理の全量
BATCHES = reservation notification

# allターゲットでは「validate → build → run」を一括実行
all: validate build run

local-all: validate build
	@echo "==> Running with ENV=LOCAL..."
	@ENV=LOCAL $(BUILD_DIR)/reservation-batch
	@ENV=LOCAL $(BUILD_DIR)/notification-batch

# ビルド
build:
	@if [ -z "$(BATCH)" ]; then \
		echo "==> Building all batch processes"; \
		for batch_type in $(BATCHES); do \
			echo "Building $$batch_type batch..."; \
			go build -ldflags "-s -w" -o $(BUILD_DIR)/$$batch_type-batch cmd/batch/$$batch_type/main.go; \
		done; \
	elif [ -f "cmd/batch/$(BATCH)/main.go" ]; then \
		echo "==> Building $(BATCH) batch only"; \
		go build -ldflags "-s -w" -o $(BUILD_DIR)/$(BATCH)-batch cmd/batch/$(BATCH)/main.go; \
	else \
		echo "==> Unknown batch type: $(BATCH). Please ensure cmd/batch/$(BATCH)/main.go exists."; \
		echo "==> Available batch types:"; \
		find cmd/batch -mindepth 1 -maxdepth 1 -type d -exec basename {} \; | sort; \
		exit 1; \
	fi

# クリーンアップ
clean:
	@echo "==> Cleaning build outputs"
	rm -rf $(BUILD_DIR)/

##
# 検証系: fmt, vet, test
##
validate:
	@echo "==> Running go fmt"
	go fmt ./...

	@echo "==> Running go vet"
	go vet ./...

	@echo "==> Running golangci-lint"
	golangci-lint run

#   必要に応じてテストを実行する
# 	@echo "==> Running tests"
# 	go test -v ./...

##
# 実行
##
run:
	@if [ -z "$(BATCH)" ]; then \
		echo "==> Running all batch processes"; \
		for batch_type in $(BATCHES); do \
			echo "Running $$batch_type batch..."; \
			$(BUILD_DIR)/$$batch_type-batch; \
		done; \
	elif [ -f "$(BUILD_DIR)/$(BATCH)-batch" ]; then \
		echo "==> Running $(BATCH) batch only"; \
		$(BUILD_DIR)/$(BATCH)-batch; \
	else \
		echo "==> Unknown batch type: $(BATCH). Please ensure $(BUILD_DIR)/$(BATCH)-batch exists."; \
		echo "==> Available batch types:"; \
		for batch_type in $(BATCHES); do \
			echo "  - $$batch_type"; \
		done; \
		exit 1; \
	fi

##
# テスト (validate でも実行しているが、個別でも呼び出せるように)
##
test:
	@echo "==> Running tests"
	go test -v ./...

# 開発ツールのインストール
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

##
# 依存関係の更新: go mod tidy
##
update-deps:
	@echo "==> Updating dependencies"
	go mod tidy

##
# Linux向けクロスコンパイル
##
build-linux:
	@echo "==> Cross compiling for Linux (amd64)"
	@GOOS=linux GOARCH=arm64 CGO_ENABLED=0 make build 