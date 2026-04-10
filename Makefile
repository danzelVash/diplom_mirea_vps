SERVICES := \
	api-gateway \
	edge-bridge-service \
	device-service \
	context-service \
	scenario-service \
	voice-service \
	vision-service \
	notification-service \
	edge-agent

.PHONY: list
list:
	@printf "%s\n" $(SERVICES)

.PHONY: tree
tree:
	@find services -maxdepth 2 -type d | sort

