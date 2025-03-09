# Makefile for Medusa Services
# Provides commands for managing Docker containers, builds, and cleanup

# Configuration
DOCKER_COMPOSE := docker-compose
MONGODB_PORT := 27017
MONGODB_HOST := localhost

# Colors for prettier output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

.PHONY: help build up down restart clean clean-all check-mongodb update-db-config logs ps status

# Display help information
help:
	@echo "${BLUE}Medusa Services Makefile${NC}"
	@echo "${YELLOW}Available commands:${NC}"
	@echo "  ${GREEN}build${NC}         - Build all service images"
	@echo "  ${GREEN}build-nocache${NC} - Build all service images no cache"
	@echo "  ${GREEN}build-auth${NC}    - Build only the auth service"
	@echo "  ${GREEN}build-core${NC}    - Build only the core service"
	@echo "  ${GREEN}up${NC}            - Start all services"
	@echo "  ${GREEN}up-auto${NC}            - Start all services"
	@echo "  ${GREEN}up-external-db${NC} - Start services using external MongoDB"
	@echo "  ${GREEN}down${NC}          - Stop all services"
	@echo "  ${GREEN}restart${NC}       - Restart all services"
	@echo "  ${GREEN}restart-auth${NC}  - Restart only the auth service"
	@echo "  ${GREEN}restart-core${NC}  - Restart only the core service"
	@echo "  ${GREEN}clean${NC}         - Stop and remove containers"
	@echo "  ${GREEN}clean-volumes${NC} - Stop and remove containers and volumes"
	@echo "  ${GREEN}clean-all${NC}     - Stop and remove containers, volumes, and images"
	@echo "  ${GREEN}logs${NC}          - View logs from all services"
	@echo "  ${GREEN}logs-auth${NC}     - View logs from auth service"
	@echo "  ${GREEN}logs-core${NC}     - View logs from core service"
	@echo "  ${GREEN}ps${NC}            - List running containers"
	@echo "  ${GREEN}status${NC}        - Show detailed status of all services"

# Build all service images
build:
	@echo "${BLUE}Building all service images...${NC}"
	$(DOCKER_COMPOSE) build  --no-cache

build-nocache:
	@echo "${BLUE}Building all service images...${NC}"
	$(DOCKER_COMPOSE) build  --no-cache

# Build specific services
build-auth:
	@echo "${BLUE}Building auth service image...${NC}"
	$(DOCKER_COMPOSE) build auth-service

build-core:
	@echo "${BLUE}Building core service image...${NC}"
	$(DOCKER_COMPOSE) build core-service

# Start all services
up:
	@echo "${BLUE}Starting all services...${NC}"
	$(DOCKER_COMPOSE) up -d
	@echo "${GREEN}Services started!${NC}"
	@echo "Auth Service: http://localhost:50051"
	@echo "Core Service: http://localhost:8080"

up-with-db:
	@echo "${BLUE}Starting all services with MongoDB...${NC}"
	$(DOCKER_COMPOSE) --profile database up -d
	@echo "${GREEN}Services started!${NC}"

# Start services using external MongoDB
up-without-db:
	@echo "${BLUE}Starting services without MongoDB...${NC}"
	@echo "${YELLOW}Make sure MongoDB is running at $(MONGODB_HOST):$(MONGODB_PORT)${NC}"
	$(DOCKER_COMPOSE) up -d
	@echo "${GREEN}Services started!${NC}"

up-auto:
	@if nc -z $(MONGODB_HOST) $(MONGODB_PORT) 2>/dev/null; then \
		echo "${GREEN}MongoDB is already running on $(MONGODB_HOST):$(MONGODB_PORT)${NC}"; \
		$(MAKE) up-without-db; \
	else \
		echo "${YELLOW}MongoDB is not running on $(MONGODB_HOST):$(MONGODB_PORT), starting it...${NC}"; \
		$(MAKE) up-with-db; \
	fi

# Check if MongoDB is already running
check-mongodb:
	@echo "${BLUE}Checking if MongoDB is already running...${NC}"
	@if nc -z $(MONGODB_HOST) $(MONGODB_PORT) 2>/dev/null; then \
		echo "${GREEN}MongoDB is already running on $(MONGODB_HOST):$(MONGODB_PORT)${NC}"; \
		echo "MONGODB_RUNNING=true" > .mongodb-status; \
	else \
		echo "${YELLOW}MongoDB is not running on $(MONGODB_HOST):$(MONGODB_PORT)${NC}"; \
		echo "MONGODB_RUNNING=false" > .mongodb-status; \
	fi

# Update database configuration in environment files
update-db-config:
	@echo "${BLUE}Updating database configuration for external MongoDB...${NC}"
	@sed -i "s|MONGO_URI=.*|MONGO_URI=mongodb://$(MONGODB_HOST):$(MONGODB_PORT)|g" auth-service/.env || true
	@sed -i "s|MONGO_URI=.*|MONGO_URI=mongodb://$(MONGODB_HOST):$(MONGODB_PORT)|g" core-service/.env || true
	@echo "${GREEN}Database configuration updated!${NC}"

# Start services using external MongoDB
up-external-db: check-mongodb update-db-config
	@echo "${BLUE}Starting services with external MongoDB...${NC}"
	@source .mongodb-status && \
	if [ "$$MONGODB_RUNNING" = "true" ]; then \
		echo "${GREEN}Using existing MongoDB at $(MONGODB_HOST):$(MONGODB_PORT)${NC}"; \
		$(DOCKER_COMPOSE) up -d auth-service core-service; \
	else \
		echo "${RED}No MongoDB found at $(MONGODB_HOST):$(MONGODB_PORT). Starting all services including MongoDB...${NC}"; \
		$(DOCKER_COMPOSE) up -d; \
	fi
	@echo "${GREEN}Services started!${NC}"

# Stop all services
down:
	@echo "${BLUE}Stopping all services...${NC}"
	$(DOCKER_COMPOSE) down
	@echo "${GREEN}Services stopped!${NC}"

# Restart all services
restart:
	@echo "${BLUE}Restarting all services...${NC}"
	$(DOCKER_COMPOSE) restart
	@echo "${GREEN}Services restarted!${NC}"

# Restart specific services
restart-auth:
	@echo "${BLUE}Restarting auth service...${NC}"
	$(DOCKER_COMPOSE) restart auth-service

restart-core:
	@echo "${BLUE}Restarting core service...${NC}"
	$(DOCKER_COMPOSE) restart core-service

# Stop and remove containers
clean:
	@echo "${BLUE}Cleaning up containers...${NC}"
	$(DOCKER_COMPOSE) down
	@echo "${GREEN}Containers removed!${NC}"

# Stop and remove containers and volumes
clean-volumes:
	@echo "${BLUE}Cleaning up containers and volumes...${NC}"
	$(DOCKER_COMPOSE) down -v
	@echo "${GREEN}Containers and volumes removed!${NC}"

# Stop and remove containers, volumes, and images
clean-all:
	@echo "${BLUE}Cleaning up everything (containers, volumes, images)...${NC}"
	$(DOCKER_COMPOSE) down -v --rmi all --remove-orphans
	@echo "${GREEN}All resources removed!${NC}"

# View logs from all services
logs:
	$(DOCKER_COMPOSE) logs -f

# View logs from specific services
logs-auth:
	$(DOCKER_COMPOSE) logs -f auth-service

logs-core:
	$(DOCKER_COMPOSE) logs -f core-service

# List running containers
ps:
	@echo "${BLUE}Running containers:${NC}"
	$(DOCKER_COMPOSE) ps

# Show detailed status of all services
status:
	@echo "${BLUE}===== MEDUSA SERVICES STATUS =====${NC}"
	@echo "${YELLOW}Container Status:${NC}"
	@$(DOCKER_COMPOSE) ps
	@echo "\n${YELLOW}MongoDB Status:${NC}"
	@if nc -z $(MONGODB_HOST) $(MONGODB_PORT) 2>/dev/null; then \
		echo "${GREEN}MongoDB is running on $(MONGODB_HOST):$(MONGODB_PORT)${NC}"; \
	else \
		echo "${RED}MongoDB is NOT running on $(MONGODB_HOST):$(MONGODB_PORT)${NC}"; \
	fi
	@echo "\n${YELLOW}Auth Service:${NC}"
	@if $(DOCKER_COMPOSE) ps | grep -q "auth-service.*Up"; then \
		echo "${GREEN}Auth Service is running${NC}"; \
	else \
		echo "${RED}Auth Service is NOT running${NC}"; \
	fi
	@echo "\n${YELLOW}Core Service:${NC}"
	@if $(DOCKER_COMPOSE) ps | grep -q "core-service.*Up"; then \
		echo "${GREEN}Core Service is running${NC}"; \
	else \
		echo "${RED}Core Service is NOT running${NC}"; \
	fi