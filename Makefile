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

tunnel:
	ssh -R 80:localhost:55060 localhost.run
