# SănDeal (shopass) - Local Development Guide

This guide provides step-by-step instructions for engineers to set up and run the SănDeal platform in a local development environment.

## 1. Prerequisites

Before you begin, ensure you have the following installed on your machine:
- **Go** (1.22+)
- **Python** (3.12+) & **uv** (for the ML service)
- **Node.js** (18+) & **npm** (for the web frontend)
- **PostgreSQL** (15+)
- **Playwright** (for the scraper service)
- **Git**

## 2. Database Setup (PostgreSQL)

SănDeal uses a unified PostgreSQL database for all services.

1. **Start PostgreSQL** locally (via Docker, Postgres.app, or homebrew).
2. **Create the database**:
   ```bash
   createdb shopass_db
   ```
3. **Run Migrations**:
   The migration files are located in `db/migrations/`. You can apply them using your preferred migration tool (e.g., `golang-migrate`) or manually apply the SQL scripts in order:
   ```bash
   # Example using golang-migrate
   migrate -path db/migrations -database "postgres://localhost:5432/shopass_db?sslmode=disable" up
   ```
4. **Seed the Database** (Optional but recommended):
   Apply the seed data in `db/seed/` to populate platforms, initial policies, etc.

## 3. Backend Services (Go)

The backend consists of multiple Go microservices under the `services/` directory.

### Environment Variables
For local development, you will typically need the following environment variables exported (you can create a `.env` file or export them in your terminal):
```bash
export DATABASE_URL="postgres://localhost:5432/shopass_db?sslmode=disable"
export JWT_SECRET="your-local-dev-jwt-secret"
export CGO_ENABLED=0 # Recommended for macOS
```

### Running Core Services
Each service can be run independently. Open multiple terminal tabs or use a multiplexer (like `tmux`) to run the necessary services:

**1. API Gateway** (Routes requests & validates JWTs):
```bash
go run services/gateway/cmd/gateway/main.go
```
*Note: Make sure the gateway port (default usually 8080) is accessible.*

**2. Auth Service** (Handles login/registration):
```bash
go run services/auth/cmd/auth/main.go
```

**3. Deal Service** (Manages deals & HTTP fanout to Notifsvc):
```bash
go run services/deal/cmd/dealsvc/main.go
```

**4. Notification Service** (Handles email/push notifications):
```bash
go run services/notif/cmd/notifsvc/main.go
```

**5. Scrape Service** (Local Playwright scraping for TikTok/Shopee):
```bash
# Ensure playwright browsers are installed first!
go run github.com/playwright-community/playwright-go/cmd/playwright@v0.2000.1 install --with-deps

go run services/scrape/cmd/scrape/main.go
```

*(You can run other services like `bill`, `cart`, `comply`, `affil`, `track`, `price` similarly as needed for your specific task).*

## 4. Machine Learning Service (Python)

The ML service handles price prediction and NLP tasks.

1. Navigate to the ML directory:
   ```bash
   cd services/ml
   ```
2. Create and activate a virtual environment using `uv`:
   ```bash
   uv venv
   source .venv/bin/activate
   ```
3. Install dependencies:
   ```bash
   uv pip install -r requirements.txt
   ```
4. Run the ML API (assuming a standard runner like uvicorn):
   ```bash
   python main.py # or uvicorn main:app --reload
   ```

## 5. Web Frontend (React/Next.js/Vite)

The web application is located in the `web/` directory.

1. Navigate to the web directory:
   ```bash
   cd web
   ```
2. Install dependencies:
   ```bash
   npm install
   ```
3. Start the development server:
   ```bash
   npm run dev
   ```
4. Open your browser to `http://localhost:3000` (or the port specified in the console).

## 6. Chrome Extension

To test the SănDeal Chrome extension locally:

1. Open Google Chrome.
2. Navigate to `chrome://extensions/`.
3. Enable **Developer mode** in the top right corner.
4. Click **Load unpacked**.
5. Select the `extension/` directory in the `shopass` repository.

Whenever you make changes to the extension files, you may need to click the "Refresh" icon on the extension card in `chrome://extensions/`.

---

### Troubleshooting

- **CGO/SQLite Issues**: Ensure `CGO_ENABLED=0` is set if you encounter binary compilation issues on macOS, unless a specific module (like `go-sqlite3`) strictly requires CGO. (We use Postgres, so CGO can safely be disabled).
- **Playwright Mismatch**: Ensure the Playwright Go package version in `scrape` matches the installed browsers (`v0.2000.1`).
- **Missing Environment Variables**: Check the logs of failing services. If a service panics on startup, it is likely missing a `DATABASE_URL` or a configuration port.
