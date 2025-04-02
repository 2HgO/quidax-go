-include env.MAK
export

.PHONY: start
start:
ifeq ($(shell which docker),)
$(error docker required for application)
endif
	@docker compose up datadb app

.PHONY: run
run:
	@go run .
