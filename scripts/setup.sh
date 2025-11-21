#!/bin/bash

# Scrapp'd Setup Script
# This script initializes the development environment

set -e

echo "🎨 Scrapp'd Development Environment Setup"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

echo "1️⃣  Checking prerequisites..."
echo ""

# Check Docker
if command_exists docker; then
    print_success "Docker is installed"
else
    print_error "Docker is not installed. Please install Docker first."
    exit 1
fi

# Check Docker Compose
if command_exists docker-compose; then
    print_success "Docker Compose is installed"
else
    print_error "Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Check Go
if command_exists go; then
    GO_VERSION=$(go version | awk '{print $3}')
    print_success "Go is installed ($GO_VERSION)"
else
    print_warning "Go is not installed. API service development will require Go 1.21+"
fi

# Check Python
if command_exists python3; then
    PYTHON_VERSION=$(python3 --version)
    print_success "Python is installed ($PYTHON_VERSION)"
else
    print_warning "Python is not installed. ML service development will require Python 3.10+"
fi

# Check Flutter
if command_exists flutter; then
    FLUTTER_VERSION=$(flutter --version | head -n 1)
    print_success "Flutter is installed ($FLUTTER_VERSION)"
else
    print_warning "Flutter is not installed. Mobile app development will require Flutter 3.16+"
fi

echo ""
echo "2️⃣  Setting up environment files..."
echo ""

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    cp .env.example .env
    print_success "Created root .env file"
else
    print_warning ".env file already exists, skipping"
fi

# Create API .env file
if [ ! -f services/api/.env ]; then
    cat > services/api/.env << EOF
ENV=development
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=scrappd
DB_PASSWORD=scrappd_dev_password
DB_NAME=scrappd
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=scrappd_redis_password
JWT_SECRET=$(openssl rand -base64 32)
ML_SERVICE_URL=http://localhost:8000
EOF
    print_success "Created API .env file"
else
    print_warning "API .env file already exists, skipping"
fi

# Create ML service .env file
if [ ! -f services/ml-service/.env ]; then
    cat > services/ml-service/.env << EOF
ENV=development
PORT=8000
MODEL_PATH=./models
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=scrappd_redis_password
EOF
    print_success "Created ML service .env file"
else
    print_warning "ML service .env file already exists, skipping"
fi

# Create mobile .env file
if [ ! -f mobile/.env ]; then
    cat > mobile/.env << EOF
API_URL=http://localhost:8080
ML_SERVICE_URL=http://localhost:8000
EOF
    print_success "Created mobile .env file"
else
    print_warning "Mobile .env file already exists, skipping"
fi

echo ""
echo "3️⃣  Starting infrastructure services..."
echo ""

# Start PostgreSQL and Redis
docker-compose up -d postgres redis

echo "Waiting for services to be ready..."
sleep 5

# Check if PostgreSQL is ready
until docker-compose exec -T postgres pg_isready -U scrappd > /dev/null 2>&1; do
    echo "Waiting for PostgreSQL..."
    sleep 2
done
print_success "PostgreSQL is ready"

# Check if Redis is ready
until docker-compose exec -T redis redis-cli -a scrappd_redis_password ping > /dev/null 2>&1; do
    echo "Waiting for Redis..."
    sleep 2
done
print_success "Redis is ready"

echo ""
echo "4️⃣  Setting up API service..."
echo ""

if command_exists go; then
    cd services/api
    
    # Install Go dependencies
    go mod download
    print_success "Go dependencies installed"
    
    # Install development tools
    if ! command_exists air; then
        go install github.com/cosmtrek/air@latest
        print_success "Installed Air for hot reload"
    fi
    
    if ! command_exists migrate; then
        go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
        print_success "Installed golang-migrate"
    fi
    
    # Run migrations
    make migrate-up
    print_success "Database migrations completed"
    
    cd ../..
else
    print_warning "Skipping API setup (Go not installed)"
fi

echo ""
echo "5️⃣  Setting up ML service..."
echo ""

if command_exists python3; then
    cd services/ml-service
    
    # Create virtual environment
    if [ ! -d "venv" ]; then
        python3 -m venv venv
        print_success "Created Python virtual environment"
    fi
    
    # Activate virtual environment and install dependencies
    source venv/bin/activate 2>/dev/null || . venv/Scripts/activate 2>/dev/null
    pip install --upgrade pip > /dev/null
    pip install -r requirements.txt > /dev/null
    pip install -r requirements-dev.txt > /dev/null
    print_success "Python dependencies installed"
    
    # Create models directory
    mkdir -p models
    print_success "Created models directory"
    
    deactivate 2>/dev/null || true
    cd ../..
else
    print_warning "Skipping ML service setup (Python not installed)"
fi

echo ""
echo "6️⃣  Setting up mobile app..."
echo ""

if command_exists flutter; then
    cd mobile
    
    # Get Flutter dependencies
    flutter pub get > /dev/null
    print_success "Flutter dependencies installed"
    
    cd ..
else
    print_warning "Skipping mobile app setup (Flutter not installed)"
fi

echo ""
echo "=========================================="
echo "✨ Setup Complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo ""
echo "1. Configure your environment variables in .env files"
echo "2. Download ML models (if needed):"
echo "   cd services/ml-service && make download-models"
echo ""
echo "3. Start the services:"
echo "   - API:        cd services/api && make dev"
echo "   - ML Service: cd services/ml-service && make dev"
echo "   - Mobile:     cd mobile && flutter run"
echo ""
echo "4. Access the services:"
echo "   - API:        http://localhost:8080"
echo "   - ML Service: http://localhost:8000"
echo "   - API Docs:   http://localhost:8000/docs"
echo ""
echo "5. Optional - Start admin tools:"
echo "   docker-compose --profile tools up -d"
echo "   - pgAdmin:         http://localhost:5050"
echo "   - Redis Commander: http://localhost:8081"
echo ""
echo "For more information, see README.md"
echo ""