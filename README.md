# Wealth Management

A local web service for personal wealth management, built with a Go backend API and a React.js frontend. Runs as a Docker container for portability.

## Features

- **Dashboard** — net worth snapshot, account balances, recent transactions
- **Accounts** — manage bank, savings, investment, and credit accounts
- **Transactions** — record and filter income / expense transactions
- **Portfolio** — track investment holdings with gain / loss calculations

## Tech Stack

| Layer     | Technology              |
|-----------|-------------------------|
| Backend   | Go 1.21 + Gin           |
| Frontend  | React 18 + react-scripts |
| Proxy     | Nginx                   |
| Container | Docker + Docker Compose |

## Quick Start — Docker (recommended)

```bash
# Clone the repo
git clone https://github.com/ginohsieh/wealth-management.git
cd wealth-management

# Build and start both services
docker compose up --build
```

| Service  | URL                      |
|----------|--------------------------|
| Frontend | http://localhost:3000     |
| Backend  | http://localhost:8080/api |

Stop with `docker compose down`.

## Local Development (without Docker)

### Backend

```bash
cd backend
go run .
# API available at http://localhost:8080
```

### Frontend

```bash
cd frontend
npm install
npm start
# UI available at http://localhost:3000
```

The `.env.development` file already points the frontend at `http://localhost:8080`, so no extra configuration is needed.

## API Reference

| Method | Path               | Description           |
|--------|--------------------|-----------------------|
| GET    | /api/health        | Health check          |
| GET    | /api/summary       | Financial summary     |
| GET    | /api/accounts      | List accounts         |
| POST   | /api/accounts      | Create account        |
| GET    | /api/accounts/:id  | Get account           |
| PUT    | /api/accounts/:id  | Update account        |
| DELETE | /api/accounts/:id  | Delete account        |
| GET    | /api/transactions  | List transactions     |
| POST   | /api/transactions  | Create transaction    |
| GET    | /api/portfolio     | List portfolio assets |
| POST   | /api/portfolio     | Add portfolio asset   |
| PUT    | /api/portfolio/:id | Update portfolio asset |
| DELETE | /api/portfolio/:id | Remove portfolio asset |

## Project Structure

```
wealth-management/
├── docker-compose.yml
├── backend/
│   ├── Dockerfile
│   ├── go.mod / go.sum
│   ├── main.go
│   ├── handlers/        # HTTP handlers
│   └── store/           # In-memory data store
└── frontend/
    ├── Dockerfile
    ├── nginx.conf
    ├── package.json
    ├── public/
    └── src/
        ├── App.js
        └── components/  # Dashboard, Accounts, Transactions, Portfolio
```
