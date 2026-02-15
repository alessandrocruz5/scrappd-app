# Docker Setup Guide

## First Time Setup

1. **Start database services:**
```bash
   make db-up
```

2. **Verify services are running:**
```bash
   make db-ps
```

3. **Check database health:**
```bash
   make db-health
```

4. **Run migrations:**
```bash
   make migrate-up
```

## Daily Development
```bash
# Start databases
make db-up

# Start ML service
make ml-up

# Start backend API (in backend/)
# Ensure DB_PORT=5433 in your environment when using docker-compose
cd backend && make dev

# View logs
make db-logs        # Database logs
make ml-logs        # ML service logs
```

## Optional Tools
```bash
# Start pgAdmin
make tools-up

# Access:
# - pgAdmin: http://localhost:5050 (admin@scrappd.local / admin)
```

## Troubleshooting
```bash
# Reset everything (DESTRUCTIVE)
make db-reset

# View what's running
make db-ps

# Connect to database
make db-shell

# Stop everything
make services-down
```

## Port Reference

- **5433** - PostgreSQL (host → container 5432)
- **8000** - ML Service
- **8080** - Backend API
- **5050** - pgAdmin (with --profile tools)
