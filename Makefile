SOURCES=$(shell find . -name "*.go" | grep -v vendor/)
PACKAGES=$(shell go list ./... | grep -v vendor/)

NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

check: fmt imports lint vet errcheck test

fmt:
	@echo "$(WARN_COLOR)+ $@$(NO_COLOR)"
	gofmt -s -w $(SOURCES)

fmt-ci:
	@echo "$(WARN_COLOR)+ $@$(NO_COLOR)"
	fgt gofmt -s -l -d $(SOURCES)

imports:
	@echo "$(WARN_COLOR)+ $@$(NO_COLOR)"
	goimports -w $(SOURCES)

imports-ci:
	@echo "$(WARN_COLOR)+ $@$(NO_COLOR)"
	goimports -d -e $(SOURCES)

lint:
	@echo "$(WARN_COLOR)+ $@$(NO_COLOR)"
	echo $(PACKAGES) | xargs -n1 golint -set_exit_status

vet:
	@echo "$(WARN_COLOR)+ $@$(NO_COLOR)"
	go vet $(PACKAGES)

errcheck:
	@echo "$(WARN_COLOR)+ $@$(NO_COLOR)"
	errcheck -ignore Close $(PACKAGES)

test:
	@echo "$(WARN_COLOR)+ $@$(NO_COLOR)"
	go test ${PACKAGES}

test-ci:
	@echo "$(WARN_COLOR)+ $@$(NO_COLOR)"
	go test -race ${PACKAGES}
