.PHONY: build test clean help fmt lint run

VERSION := 0.1.0-alpha
BINARY := jcc
MAIN := cmd/jcc/main.go
BUILD_DIR := bin

# 빌드 설정
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

# 색상 출력
GREEN := \033[0;32m
BLUE := \033[0;34m
RED := \033[0;31m
NC := \033[0m # No Color

help:
	@echo "$(BLUE)JuliaCC Makefile Commands:$(NC)"
	@echo "  $(GREEN)make build$(NC)       - 바이너리 빌드"
	@echo "  $(GREEN)make test$(NC)        - 테스트 실행"
	@echo "  $(GREEN)make clean$(NC)       - 빌드 산출물 제거"
	@echo "  $(GREEN)make fmt$(NC)         - 코드 포맷팅"
	@echo "  $(GREEN)make lint$(NC)        - 코드 검사"
	@echo "  $(GREEN)make run$(NC)         - 예제 코드 실행"

build:
	@echo "$(BLUE)🔨 빌드 시작...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) $(MAIN)
	@echo "$(GREEN)✅ 빌드 완료: $(BUILD_DIR)/$(BINARY)$(NC)"

test:
	@echo "$(BLUE)🧪 테스트 실행...$(NC)"
	@go test -v ./... -cover

test-verbose:
	@echo "$(BLUE)🧪 상세 테스트 실행...$(NC)"
	@go test -v ./... -race -cover

clean:
	@echo "$(BLUE)🧹 정리 중...$(NC)"
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "$(GREEN)✅ 정리 완료$(NC)"

fmt:
	@echo "$(BLUE)📝 포맷팅 중...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✅ 포맷팅 완료$(NC)"

lint:
	@echo "$(BLUE)🔍 검사 중...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✅ 검사 완료$(NC)"

run: build
	@echo "$(BLUE)▶️  실행 중...$(NC)"
	@echo "test" > /tmp/test.jl
	@$(BUILD_DIR)/$(BINARY) /tmp/test.jl

# 의존성 설치
deps:
	@echo "$(BLUE)📦 의존성 설치...$(NC)"
	@go mod download
	@go mod verify
	@echo "$(GREEN)✅ 의존성 설치 완료$(NC)"

# 전체 체크
check: fmt lint test
	@echo "$(GREEN)✅ 모든 체크 완료$(NC)"

# Phase별 타겟
phase0:
	@echo "$(BLUE)Phase 0: 초기화 완료$(NC)"
	@echo "상태: ✅ 프로젝트 구조 생성"
	@echo "      ✅ Go 모듈 초기화"
	@echo "      ✅ Main 프로그램 작성"

phase1-prepare:
	@echo "$(BLUE)Phase 1 준비: Lexer 구현$(NC)"
	@echo "대상: internal/lexer/lexer.go"
	@echo "예상 코드량: 500-800줄"

.DEFAULT_GOAL := help
