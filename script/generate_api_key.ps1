# API Key Generator for Crypto Trading API (PowerShell)
# Works on Windows PowerShell and PowerShell Core

Write-Host "🔐 Crypto Trading API - API Key Generator" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# Method 1: Using .NET Cryptography (Most secure)
Write-Host "Method 1: .NET Cryptography (Recommended)" -ForegroundColor Green
Write-Host "------------------------------------------" -ForegroundColor Green

$bytes32 = New-Object byte[] 32
$rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
$rng.GetBytes($bytes32)
$apiKey32 = [Convert]::ToBase64String($bytes32)
Write-Host "32-byte API Key (Strong):" -ForegroundColor Yellow
Write-Host $apiKey32
Write-Host ""

$bytes48 = New-Object byte[] 48
$rng.GetBytes($bytes48)
$apiKey48 = [Convert]::ToBase64String($bytes48)
Write-Host "48-byte API Key (Very Strong):" -ForegroundColor Yellow
Write-Host $apiKey48
Write-Host ""

# Method 2: Hex format
$bytesHex = New-Object byte[] 32
$rng.GetBytes($bytesHex)
$apiKeyHex = [BitConverter]::ToString($bytesHex).Replace("-", "").ToLower()
Write-Host "Hex format (64 chars):" -ForegroundColor Yellow
Write-Host $apiKeyHex
Write-Host ""

# Method 3: UUID-based
Write-Host "Method 2: UUID-based" -ForegroundColor Green
Write-Host "--------------------" -ForegroundColor Green
$uuid1 = [guid]::NewGuid().ToString("N")
$uuid2 = [guid]::NewGuid().ToString("N")
$apiKeyUuid = "$uuid1$uuid2"
Write-Host "UUID-based key:" -ForegroundColor Yellow
Write-Host $apiKeyUuid
Write-Host ""

# Method 4: Alphanumeric (64 chars)
Write-Host "Method 3: Alphanumeric (64 chars)" -ForegroundColor Green
Write-Host "---------------------------------" -ForegroundColor Green
$chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
$apiKeyAlpha = -join ((1..64) | ForEach-Object { $chars[(Get-Random -Maximum $chars.Length)] })
Write-Host $apiKeyAlpha
Write-Host ""

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "✅ Choose one of the above keys and add it to your .env file" -ForegroundColor Green
Write-Host "⚠️  NEVER commit the .env file to version control" -ForegroundColor Red
Write-Host "💡 Recommendation: Use the 48-byte key for production" -ForegroundColor Yellow
Write-Host ""
Write-Host "To add to .env file:" -ForegroundColor Cyan
Write-Host "  Add-Content -Path .env -Value 'API_KEY=$apiKey48'" -ForegroundColor White
Write-Host ""
Write-Host "Or copy manually to your .env file" -ForegroundColor Cyan

$rng.Dispose()
