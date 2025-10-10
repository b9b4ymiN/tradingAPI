#!/bin/bash

# API Key Generator for Crypto Trading API
# This script generates secure random API keys

echo "🔐 Crypto Trading API - API Key Generator"
echo "=========================================="
echo ""

# Method 1: Using OpenSSL (Most secure, works on Linux/Mac/Git Bash)
if command -v openssl &> /dev/null; then
    echo "Method 1: OpenSSL (Recommended)"
    echo "--------------------------------"

    # Generate 32-byte random key, base64 encoded
    API_KEY_32=$(openssl rand -base64 32)
    echo "32-byte API Key (Strong):"
    echo "$API_KEY_32"
    echo ""

    # Generate 48-byte random key, base64 encoded
    API_KEY_48=$(openssl rand -base64 48)
    echo "48-byte API Key (Very Strong):"
    echo "$API_KEY_48"
    echo ""

    # Generate hex format (alternative)
    API_KEY_HEX=$(openssl rand -hex 32)
    echo "Hex format (64 chars):"
    echo "$API_KEY_HEX"
    echo ""
fi

# Method 2: Using /dev/urandom (Linux/Mac)
if [ -f /dev/urandom ]; then
    echo "Method 2: /dev/urandom"
    echo "----------------------"
    API_KEY_URANDOM=$(cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 48 | head -n 1)
    echo "Random alphanumeric (48 chars):"
    echo "$API_KEY_URANDOM"
    echo ""
fi

# Method 3: Using uuidgen (works on most systems)
if command -v uuidgen &> /dev/null; then
    echo "Method 3: UUID-based"
    echo "--------------------"
    UUID1=$(uuidgen)
    UUID2=$(uuidgen)
    API_KEY_UUID="${UUID1}${UUID2}"
    echo "UUID-based key:"
    echo "$API_KEY_UUID"
    echo ""
fi

# Method 4: Using date + random (fallback for Windows)
echo "Method 4: Date + Random (Basic)"
echo "--------------------------------"
API_KEY_BASIC="api_$(date +%s)_$RANDOM$RANDOM$RANDOM"
echo "Basic key (fallback):"
echo "$API_KEY_BASIC"
echo ""

echo "=========================================="
echo "✅ Choose one of the above keys and add it to your .env file"
echo "⚠️  NEVER commit the .env file to version control"
echo "💡 Recommendation: Use the 48-byte OpenSSL key for production"
echo ""
echo "To add to .env file:"
echo "  echo 'API_KEY=YOUR_CHOSEN_KEY' >> .env"
