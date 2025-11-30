#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================${NC}"
echo -e "${BLUE}🖼️  Scrapp'd Image Upload Tests${NC}"
echo -e "${BLUE}================================${NC}\n"

# Check if image file is provided
if [ -z "$1" ]; then
    echo -e "${RED}❌ Error: No image file provided${NC}"
    echo -e "${YELLOW}Usage: $0 <path-to-image.jpg>${NC}"
    echo -e "${YELLOW}Example: $0 ~/Pictures/test.jpg${NC}"
    exit 1
fi

IMAGE_FILE="$1"

# Check if file exists
if [ ! -f "$IMAGE_FILE" ]; then
    echo -e "${RED}❌ Error: File not found: $IMAGE_FILE${NC}"
    exit 1
fi

# Check file size
FILE_SIZE=$(stat -f%z "$IMAGE_FILE" 2>/dev/null || stat -c%s "$IMAGE_FILE" 2>/dev/null)
FILE_SIZE_MB=$(echo "scale=2; $FILE_SIZE / 1048576" | bc)
echo -e "${YELLOW}📊 Image Info:${NC}"
echo -e "   File: $IMAGE_FILE"
echo -e "   Size: ${FILE_SIZE_MB}MB ($FILE_SIZE bytes)"

# Check file type
FILE_TYPE=$(file -b --mime-type "$IMAGE_FILE")
echo -e "   Type: $FILE_TYPE"

# Validate file size (max 10MB)
MAX_SIZE=10485760
if [ "$FILE_SIZE" -gt "$MAX_SIZE" ]; then
    echo -e "${RED}❌ Error: File too large. Maximum size is 10MB${NC}"
    exit 1
fi

# Validate file type
if [[ ! "$FILE_TYPE" =~ ^image/(jpeg|png|webp)$ ]]; then
    echo -e "${YELLOW}⚠️  Warning: File type may not be supported. Supported: JPEG, PNG, WEBP${NC}"
fi

echo ""

# Test 1: Check if services are running
echo -e "${YELLOW}Test 1: Checking if services are running...${NC}"

ML_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8000/health 2>/dev/null)
BACKEND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health 2>/dev/null)

if [ "$ML_STATUS" != "200" ]; then
    echo -e "${RED}❌ ML Service not responding (port 8000)${NC}"
    echo -e "${YELLOW}   Start it with: cd ml-service && make dev${NC}"
    exit 1
fi

if [ "$BACKEND_STATUS" != "200" ]; then
    echo -e "${RED}❌ Backend API not responding (port 8080)${NC}"
    echo -e "${YELLOW}   Start it with: cd backend && make dev${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Both services are running${NC}\n"

# Test 2: Multipart Upload to Backend API
echo -e "${YELLOW}Test 2: Upload via Backend API (Multipart)...${NC}"
echo -e "   Processing image with BiRefNet..."
echo -e "   ${YELLOW}This may take 10-15 seconds...${NC}\n"

START_TIME=$(date +%s)

HTTP_CODE=$(curl -X POST "http://localhost:8080/api/v1/ml/process" \
    -F "image=@$IMAGE_FILE" \
    --output backend_result.png \
    -w "%{http_code}" \
    -s)

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

if [ "$HTTP_CODE" = "200" ] && [ -f "backend_result.png" ]; then
    RESULT_SIZE=$(stat -f%z "backend_result.png" 2>/dev/null || stat -c%s "backend_result.png" 2>/dev/null)
    RESULT_SIZE_KB=$(echo "scale=2; $RESULT_SIZE / 1024" | bc)
    
    if [ "$RESULT_SIZE" -gt 1000 ]; then
        echo -e "${GREEN}✓ Backend upload successful!${NC}"
        echo -e "   HTTP Status: $HTTP_CODE"
        echo -e "   Processing time: ${DURATION}s"
        echo -e "   Output size: ${RESULT_SIZE_KB}KB"
        echo -e "   Saved to: ${BLUE}backend_result.png${NC}"
    else
        echo -e "${RED}❌ Output file too small, might be an error${NC}"
        cat backend_result.png
        exit 1
    fi
else
    echo -e "${RED}❌ Backend upload failed (HTTP $HTTP_CODE)${NC}"
    if [ -f "backend_result.png" ]; then
        cat backend_result.png
    fi
    exit 1
fi

echo ""

# Test 3: Base64 Upload to Backend API
echo -e "${YELLOW}Test 3: Upload via Backend API (JSON/Base64)...${NC}"
echo -e "   Encoding image to base64..."

# Create base64 encoded image
if command -v base64 > /dev/null; then
    BASE64_IMAGE=$(base64 -i "$IMAGE_FILE" 2>/dev/null || base64 "$IMAGE_FILE" 2>/dev/null | tr -d '\n')
    
    echo -e "   Sending to backend..."
    echo -e "   ${YELLOW}This may take 10-15 seconds...${NC}\n"
    
    START_TIME=$(date +%s)
    
    HTTP_CODE=$(curl -X POST "http://localhost:8080/api/v1/ml/process" \
        -H "Content-Type: application/json" \
        -d "{\"image\":\"data:$FILE_TYPE;base64,$BASE64_IMAGE\"}" \
        --output backend_result_base64.png \
        -w "%{http_code}" \
        -s)
    
    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))
    
    if [ "$HTTP_CODE" = "200" ] && [ -f "backend_result_base64.png" ]; then
        RESULT_SIZE=$(stat -f%z "backend_result_base64.png" 2>/dev/null || stat -c%s "backend_result_base64.png" 2>/dev/null)
        RESULT_SIZE_KB=$(echo "scale=2; $RESULT_SIZE / 1024" | bc)
        
        if [ "$RESULT_SIZE" -gt 1000 ]; then
            echo -e "${GREEN}✓ Base64 upload successful!${NC}"
            echo -e "   HTTP Status: $HTTP_CODE"
            echo -e "   Processing time: ${DURATION}s"
            echo -e "   Output size: ${RESULT_SIZE_KB}KB"
            echo -e "   Saved to: ${BLUE}backend_result_base64.png${NC}"
        else
            echo -e "${RED}❌ Output file too small, might be an error${NC}"
            exit 1
        fi
    else
        echo -e "${RED}❌ Base64 upload failed (HTTP $HTTP_CODE)${NC}"
        if [ -f "backend_result_base64.png" ]; then
            cat backend_result_base64.png
        fi
        exit 1
    fi
else
    echo -e "${YELLOW}⚠️  Skipped: base64 command not found${NC}"
fi

echo ""

# Test 4: Direct ML Service Upload (for comparison)
echo -e "${YELLOW}Test 4: Direct ML Service Upload (bypassing backend)...${NC}"
echo -e "   Processing with ML service directly..."
echo -e "   ${YELLOW}This may take 10-15 seconds...${NC}\n"

START_TIME=$(date +%s)

HTTP_CODE=$(curl -X POST "http://localhost:8000/process" \
    -F "image=@$IMAGE_FILE" \
    --output ml_direct_result.png \
    -w "%{http_code}" \
    -s)

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

if [ "$HTTP_CODE" = "200" ] && [ -f "ml_direct_result.png" ]; then
    RESULT_SIZE=$(stat -f%z "ml_direct_result.png" 2>/dev/null || stat -c%s "ml_direct_result.png" 2>/dev/null)
    RESULT_SIZE_KB=$(echo "scale=2; $RESULT_SIZE / 1024" | bc)
    
    if [ "$RESULT_SIZE" -gt 1000 ]; then
        echo -e "${GREEN}✓ Direct ML upload successful!${NC}"
        echo -e "   HTTP Status: $HTTP_CODE"
        echo -e "   Processing time: ${DURATION}s"
        echo -e "   Output size: ${RESULT_SIZE_KB}KB"
        echo -e "   Saved to: ${BLUE}ml_direct_result.png${NC}"
    else
        echo -e "${RED}❌ Output file too small, might be an error${NC}"
        exit 1
    fi
else
    echo -e "${RED}❌ Direct ML upload failed (HTTP $HTTP_CODE)${NC}"
    if [ -f "ml_direct_result.png" ]; then
        cat ml_direct_result.png
    fi
    exit 1
fi

echo ""
echo -e "${BLUE}================================${NC}"
echo -e "${GREEN}🎉 All tests passed!${NC}"
echo -e "${BLUE}================================${NC}"
echo ""
echo -e "${YELLOW}Output files created:${NC}"
echo -e "   ${BLUE}backend_result.png${NC} - Via backend multipart upload"
echo -e "   ${BLUE}backend_result_base64.png${NC} - Via backend JSON/base64"
echo -e "   ${BLUE}ml_direct_result.png${NC} - Direct ML service upload"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo -e "   • Open the PNG files to verify background removal quality"
echo -e "   • Test with different image types (receipts, tickets, packaging)"
echo -e "   • Try edge cases (very small images, complex backgrounds)"
echo ""