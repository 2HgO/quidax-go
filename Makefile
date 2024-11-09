ifeq ($(shell which docker),)
$(error docker required for application)
endif

start:
	@docker compose up datadb app txdbrepl1 txdbrepl2 txdbrepl3
