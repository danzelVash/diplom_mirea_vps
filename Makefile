SERVICES := \
	platform-service \
	context-service \
	voice-recognition-service \
	vision-service \
	notification-service

.PHONY: list
list:
	@printf "%s\n" $(SERVICES)

.PHONY: tree
tree:
	@find services -maxdepth 2 -type d | sort
