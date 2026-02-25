BINARY := tk
REAL_BINARY := $(BINARY)-real
GOBIN := $(shell go env GOPATH)/bin

.PHONY: build install install-direct clean

build:
	go build -o $(BINARY) .

install: build
	cp $(BINARY) $(GOBIN)/$(REAL_BINARY)
	@printf '#!/bin/sh\nexec "%s/%s" "$$@"\n' "$(GOBIN)" "$(REAL_BINARY)" > "$(GOBIN)/$(BINARY)"
	chmod +x "$(GOBIN)/$(BINARY)"
	@echo "Installed $(REAL_BINARY) to $(GOBIN)/$(REAL_BINARY)"
	@echo "Installed wrapper $(BINARY) to $(GOBIN)/$(BINARY)"

install-direct: build
	cp $(BINARY) $(GOBIN)/$(BINARY)
	@echo "Installed $(BINARY) to $(GOBIN)/$(BINARY)"

clean:
	rm -f $(BINARY)
