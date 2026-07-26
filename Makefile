# SănDeal - one-command dev + ops. Wraps the Docker Compose stack.
# Requires Docker with the Compose plugin. Run `make help` to list targets.

COMPOSE := docker compose -f deploy/docker-compose.yml
DBEXEC  := $(COMPOSE) exec -T db psql -U $(or $(POSTGRES_USER),postgres) -d $(or $(POSTGRES_DB),shopass)

.DEFAULT_GOAL := help
.PHONY: help env up down restart logs ps migrate seed scrape forecast deal-once smoke simulate-prices grant-premium psql dev-db clean test test-go test-web test-ml

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | sort | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-11s\033[0m %s\n",$$1,$$2}'

# ---------- setup ----------
env: ## Create deploy/.env from the example (edit secrets before deploying)
	@test -f deploy/.env || cp deploy/.env.example deploy/.env
	@echo 'deploy/.env ready - edit POSTGRES_PASSWORD before any non-local deploy'

# ---------- stack lifecycle ----------
up: ## Build + start the core stack (db, redis, gateway, services, web)
	$(COMPOSE) up -d --build

down: ## Stop the stack (keeps the data volume)
	$(COMPOSE) down

restart: ## Restart all services
	$(COMPOSE) restart

logs: ## Tail logs from all services
	$(COMPOSE) logs -f --tail=100

ps: ## Show service status
	$(COMPOSE) ps

migrate: ## Re-run database migrations (idempotent)
	$(COMPOSE) run --rm migrate

dev-db: ## Start only the database + migrations (for running services locally)
	$(COMPOSE) up -d db migrate

clean: ## Stop the stack and DELETE the data volume (destroys all data)
	$(COMPOSE) down -v

# ---------- demo / loop jobs ----------
seed: ## Seed a demo product (100) + user + alert rule + a mature forecast
	$(DBEXEC) -c "INSERT INTO app_user(id) VALUES (999) ON CONFLICT DO NOTHING; \
	  INSERT INTO tracked_product(id, platform_id, platform_item_id, first_seen) VALUES (100,1,'555:777', now() - INTERVAL '100 days') ON CONFLICT DO NOTHING; \
	  INSERT INTO alert_rule(user_id,product_id,rule_type,active) VALUES (999,100,'bottom_predicted',true) ON CONFLICT DO NOTHING; \
	  INSERT INTO price_forecast(product_id,run_date,horizon_day,yhat,yhat_lower,yhat_upper,p_bottom_14d,model_kind,scored_at) VALUES (100,current_date,14,100000,90000,110000,0.85,'lgbm',now()) ON CONFLICT DO NOTHING;"

scrape: ## Run the scraper once (SCRAPE_SEED=100:555:777, SHOPEE_BASE_URL=fixture for a dry run)
	$(COMPOSE) run --rm scrapesvc

forecast: ## Run the ml forecast job once
	$(COMPOSE) run --rm mlforecast

deal-once: ## Run the deal nightly bottom-price score once
	$(COMPOSE) run --rm -e RUN_ONCE=1 dealsvc

smoke: ## Demo the loop via containers: seed -> scrape -> deal -> show rows
	$(MAKE) seed
	SCRAPE_SEED=$(or $(SCRAPE_SEED),100:555:777) $(COMPOSE) run --rm scrapesvc
	$(COMPOSE) run --rm -e RUN_ONCE=1 dealsvc
	@echo '--- price_snapshot(100) ---'; $(DBEXEC) -c 'SELECT product_id, price FROM price_snapshot WHERE product_id=100;'
	@echo '--- bottom_alert_log(100) ---'; $(DBEXEC) -c 'SELECT user_id, product_id, p_bottom FROM bottom_alert_log WHERE product_id=100;'

simulate-prices: ## Post dropping price series via pricesvc ingest (no live scrape)
	@bash scripts/simulate_price_series.sh

grant-premium: ## Temporary local Premium grant (USER_ID= required; skips checkout/IPN)
	@test -n "$(USER_ID)" || (echo 'usage: make grant-premium USER_ID=<id>' >&2; exit 1)
	@USER_ID=$(USER_ID) bash scripts/grant_premium_local.sh

psql: ## Open a psql shell on the database
	$(COMPOSE) exec db psql -U $(or $(POSTGRES_USER),postgres) -d $(or $(POSTGRES_DB),shopass)

# ---------- local test suites (need Go / Node / Python, not Docker) ----------
test: test-go test-web test-ml ## Run every test suite locally

test-go: ## Go unit tests for every service
	@for m in db obs region secrets services/gateway services/auth services/price services/scrape services/deal services/notif services/track services/cart services/affil services/bill services/comply; do echo "== $$m =="; (cd $$m && go test ./...) || exit 1; done

test-web: ## Extension + web jest suites
	cd extension && npm install && npm test
	cd web && npm install && npm test

test-ml: ## ml pytest suite
	cd services/ml && pip install -r requirements.txt && pytest tests/
