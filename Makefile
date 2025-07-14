PAYMENT_PROCESSORS_PATH := ./payment-processors
PORT ?= 9999

.DEFAULT_GOAL := help
.PHONY: reset down clean-docker load-test healthcheck create-payment summary purge help

help:
	@echo ""
	@echo "Available targets:"
	@echo "  make reset            # Stop, clean volumes, and start both docker-compose setups (main + processor)."
	@echo "  make down             # Stop and remove containers and volumes from both setups."
	@echo "  make clean-docker     # Remove all Docker containers, images, volumes, and networks."
	@echo "  make load-test        # Run the load test using k6."
	@echo "  make healthcheck      # Hit the /healthcheck endpoint."
	@echo "  make create-payment   # Send a POST request to /payments."
	@echo "  make summary          # Fetch the payment summary."
	@echo "  make purge            # POST to /purge-payments."
	@echo ""

reset:
	@echo "Stopping and cleaning up existing containers and volumes..."
	docker-compose down -v
	docker-compose -f $(PAYMENT_PROCESSORS_PATH)/docker-compose.yml down -v
	docker-compose -f $(PAYMENT_PROCESSORS_PATH)/docker-compose.yml up -d --build
	docker-compose up --build

down:
	@echo "Stopping and removing containers and volumes..."
	docker-compose down -v
	docker-compose -f $(PAYMENT_PROCESSORS_PATH)/docker-compose.yml down -v

clean-docker:
	@echo "Cleaning up all Docker containers, images, volumes, and networks..."
	@if [ -n "$$(docker ps -q)" ]; then docker stop $$(docker ps -q); fi
	@if [ -n "$$(docker ps -aq)" ]; then docker rm $$(docker ps -aq); fi
	@if [ -n "$$(docker images -q)" ]; then docker rmi -f $$(docker images -q); fi
	@if [ -n "$$(docker volume ls -q)" ]; then docker volume rm $$(docker volume ls -q); fi
	@if [ -n "$$(docker network ls --filter "type=custom" -q)" ]; then docker network rm $$(docker network ls --filter "type=custom" -q); fi
	docker system prune -a --volumes -f

load-test:
	@echo "Running load test with k6..."
	@command -v k6 >/dev/null 2>&1 || { \
		echo >&2 "Error: k6 is not installed."; \
		exit 1; \
	}
	K6_WEB_DASHBOARD=true \
	K6_WEB_DASHBOARD_PORT=5665 \
	K6_WEB_DASHBOARD_PERIOD=2s \
	K6_WEB_DASHBOARD_OPEN=true \
	K6_WEB_DASHBOARD_EXPORT=report.html \
	k6 run -e MAX_REQUESTS=850 load-test/rinha.js

healthcheck: check-curl check-jq
	@echo "Checking health of the application..."
	curl -s http://localhost:$(PORT)/healthcheck | jq .

create-payment: check-curl check-jq check-uuidgen
	@echo "Creating a payment..."
	curl -s -X POST http://localhost:$(PORT)/payments \
	  -H "Content-Type: application/json" \
	  -d '{"correlationId": "'$$(uuidgen)'", "amount": 10.12}' | jq .

summary: check-curl check-jq
	@echo "Fetching payment summary for today..."
	@FROM=$$(date -u +"%Y-%m-%dT00:00:00Z"); \
	TO=$$(date -u +"%Y-%m-%dT23:59:59Z"); \
	curl -s -G "http://localhost:$(PORT)/payments-summary" \
	  --data-urlencode "from=$$FROM" \
	  --data-urlencode "to=$$TO" | jq .

purge: check-curl check-jq
	@echo "Purging payments..."
	curl -s -X POST http://localhost:$(PORT)/purge-payments | jq .

check-curl:	
	@command -v curl >/dev/null 2>&1 || { \
		echo >&2 "Error: curl is not installed."; \
		exit 1; \
	}

check-jq:
	@command -v jq >/dev/null 2>&1 || { \
		echo >&2 "Error: jq is not installed."; \
		exit 1; \
	}

check-uuidgen:
	@command -v uuidgen >/dev/null 2>&1 || { \
		echo >&2 "Error: uuidgen is not installed."; \
		exit 1; \
	}
