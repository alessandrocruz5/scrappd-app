#!/usr/bin/env python3
"""
Test script for the ML service
Run the service first: uvicorn app.main:app --reload
Then run this script: python test_service.py
"""
import requests
import sys

def test_service():
    base_url = "http://localhost:8000"
    
    # Test 1: Health check
    print("Test 1: Health check...")
    response = requests.get(f"{base_url}/health")
    print(f"Status: {response.status_code}")
    print(f"Response: {response.json()}\n")
    
    # Test 2: Process image
    print("Test 2: Process image...")
    
    # Get image path from command line or use default
    image_path = sys.argv[1] if len(sys.argv) > 1 else "../scrappd-ml-experiments/test_images/test1.jpg"
    
    try:
        with open(image_path, "rb") as f:
            files = {"file": (image_path, f, "image/jpeg")}
            response = requests.post(f"{base_url}/process", files=files)
        
        if response.status_code == 200:
            # Save processed image
            output_path = "test_output.png"
            with open(output_path, "wb") as f:
                f.write(response.content)
            
            processing_time = response.headers.get("X-Processing-Time")
            print(f"✓ Success!")
            print(f"  Processing time: {processing_time}s")
            print(f"  Output saved to: {output_path}\n")
        else:
            print(f"✗ Failed: {response.status_code}")
            print(f"  Response: {response.json()}\n")
    
    except FileNotFoundError:
        print(f"✗ Image not found: {image_path}")
        print("Usage: python test_service.py <path-to-image>")

if __name__ == "__main__":
    test_service()