IMAGE?=322456/signalonebackend
TAG?=dev

SOLUTIONS_AGENT_IMAGE?=322456/solutions-agent
SOLUTIONS_AGENT_TAG?=dev

dev-build-backend: ## Build service image - dev version
	docker build --tag=$(IMAGE):$(TAG) -f dev.Dockerfile .
	docker build --tag=$(SOLUTIONS_AGENT_IMAGE):$(SOLUTIONS_AGENT_TAG) -f ../solutionAgent/dev.Dockerfile ../solutionAgent

prod-build-backend: ## Build service image - prod version
	docker build --tag=$(IMAGE):$(TAG) .
	docker build --tag=$(SOLUTIONS_AGENT_IMAGE):$(SOLUTIONS_AGENT_TAG) ../solutionAgent

start-backend: ## Start service locally (local development only)
	docker compose -f ./docker-compose.dev.yaml up -d

start-backend-with-init-windows: ## Start service locally with mock data initialization
	docker compose -f ./docker-compose.dev.yaml up -d
	python -m venv ./devenv/initenv.venv
	powershell ./devenv/initenv.venv/Scripts/activate.ps1
	pip install -r ./devenv/requirements.txt
	python ./devenv/init.py

start-backend-with-init-macos: ## Start service locally with mock data initialization
	docker compose -f ./docker-compose.dev.yaml up -d
	python3 -m venv ./devenv/initenv.venv
	. ./devenv/initenv.venv/bin/activate
	pip3 install -r ./devenv/requirements.txt
	python3 ./devenv/init.py

deploy-backend: ## Deploy service to production
	docker compose -f ./docker-compose.prd.yaml up -d