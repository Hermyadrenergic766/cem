.PHONY: build dev install clean tidy test

BUILD_DIR := build
BINS      := cem cemi cemir
LDFLAGS   := -s -w
GOFLAGS   := -trimpath

build:
	@mkdir -p $(BUILD_DIR)
	@for name in $(BINS); do \
		echo "→ $$name"; \
		go build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$$name . ; \
	done
	@echo "✓ build/ → $(BINS)"

dev: build
	@mkdir -p $(HOME)/.local/bin
	@for name in $(BINS); do install -m 0755 $(BUILD_DIR)/$$name $(HOME)/.local/bin/$$name; done
	@echo "✓ ~/.local/bin → $(BINS)"

install: build
	@for name in $(BINS); do sudo install -m 0755 $(BUILD_DIR)/$$name /usr/local/bin/$$name; done
	@echo "✓ /usr/local/bin → $(BINS)"

tidy:
	go mod tidy

test:
	go test ./...

clean:
	rm -rf $(BUILD_DIR)
