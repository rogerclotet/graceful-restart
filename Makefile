SOURCES=$(shell find . -name "*.go" | grep -v vendor/)
PACKAGES=$(shell go list ./... | grep -v vendor/)

NO_COLOR=\033[0m
OK_COLOR=\033[32;01m
ERROR_COLOR=\033[31;01m
WARN_COLOR=\033[33;01m

check: fmt lint vet errcheck test

fmt:
	@echo "$(WARN_COLOR)+ $@$(NO_COLOR)"
	gofmt -s -w $(SOURCES)

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
