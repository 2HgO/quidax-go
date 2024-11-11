-include env.MAK
export

.PHONY: start
start:
ifeq ($(shell which docker),)
$(error docker required for application)
endif
	@docker compose up datadb app txdbrepl1 txdbrepl2 txdbrepl3

.PHONY: run
run:
	@go run .
