# Build IPA Downloader cho Windows (chạy trong PowerShell tại thư mục dự án)
# Yêu cầu: Go >= 1.26, Node.js >= 18
$ErrorActionPreference = "Stop"

Write-Host "[1/3] Build frontend (Tailwind)..." -ForegroundColor Cyan
Push-Location frontend
if (-not (Test-Path node_modules)) { npm install }
npm run build
Pop-Location

Write-Host "[2/3] Tạo resource Windows (icon, manifest)..." -ForegroundColor Cyan
go run github.com/tc-hib/go-winres@v0.3.3 make --in build/windows/winres.json --out rsrc --arch amd64

Write-Host "[3/3] Build exe..." -ForegroundColor Cyan
$env:CGO_ENABLED = "0"
go build -tags desktop,production -trimpath -ldflags "-w -s -H windowsgui" -o "build/bin/IPA Downloader.exe" .

Write-Host "Xong: build\bin\IPA Downloader.exe" -ForegroundColor Green
