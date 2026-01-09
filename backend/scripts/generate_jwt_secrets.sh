#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔐 JWT Secret Generator${NC}\n"

# Check if openssl is available
if ! command -v openssl &> /dev/null; then
    echo -e "${YELLOW}⚠️  OpenSSL not found. Please install OpenSSL first.${NC}"
    exit 1
fi

# Generate secrets
ACCESS_SECRET=$(openssl rand -base64 32)
REFRESH_SECRET=$(openssl rand -base64 32)

echo -e "${GREEN}✅ Generated JWT Secrets:${NC}\n"
echo -e "${YELLOW}Access Token Secret:${NC}"
echo "$ACCESS_SECRET"
echo ""
echo -e "${YELLOW}Refresh Token Secret:${NC}"
echo "$REFRESH_SECRET"
echo ""

# Ask if user wants to update .env file
read -p "Do you want to update backend/.env file? (y/n): " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    ENV_FILE="backend/.env"
    
    # Create .env if it doesn't exist
    if [ ! -f "$ENV_FILE" ]; then
        echo -e "${YELLOW}Creating new .env file...${NC}"
        cat > "$ENV_FILE" << EOF
# Server
SERVER_PORT=8080
ENVIRONMENT=development

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=scrappd_app
DB_PASSWORD=scrappd-go
DB_NAME=scrappd
DB_SSLMODE=disable

# ML Service
ML_SERVICE_URL=http://localhost:8000

# JWT Secrets
JWT_ACCESS_SECRET=$ACCESS_SECRET
JWT_REFRESH_SECRET=$REFRESH_SECRET
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h
EOF
        echo -e "${GREEN}✅ Created .env file with JWT secrets${NC}"
    else
        # Update existing .env file
        if grep -q "JWT_ACCESS_SECRET" "$ENV_FILE"; then
            # Replace existing secrets
            if [[ "$OSTYPE" == "darwin"* ]]; then
                # macOS
                sed -i '' "s|JWT_ACCESS_SECRET=.*|JWT_ACCESS_SECRET=$ACCESS_SECRET|" "$ENV_FILE"
                sed -i '' "s|JWT_REFRESH_SECRET=.*|JWT_REFRESH_SECRET=$REFRESH_SECRET|" "$ENV_FILE"
            else
                # Linux
                sed -i "s|JWT_ACCESS_SECRET=.*|JWT_ACCESS_SECRET=$ACCESS_SECRET|" "$ENV_FILE"
                sed -i "s|JWT_REFRESH_SECRET=.*|JWT_REFRESH_SECRET=$REFRESH_SECRET|" "$ENV_FILE"
            fi
            echo -e "${GREEN}✅ Updated existing JWT secrets in .env${NC}"
        else
            # Append secrets
            cat >> "$ENV_FILE" << EOF

# JWT Secrets
JWT_ACCESS_SECRET=$ACCESS_SECRET
JWT_REFRESH_SECRET=$REFRESH_SECRET
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h
EOF
            echo -e "${GREEN}✅ Added JWT secrets to .env${NC}"
        fi
    fi
fi

echo ""
echo -e "${BLUE}📝 Important Security Notes:${NC}"
echo -e "  • Never commit .env file to version control"
echo -e "  • Use different secrets for development and production"
echo -e "  • Rotate secrets regularly in production"
echo -e "  • Store production secrets in a secure vault (AWS Secrets Manager, etc.)"
echo ""