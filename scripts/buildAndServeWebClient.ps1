Push-Location $PSScriptRoot\..
$env:GOOS = "js"
$env:GOARCH = "wasm"
go build -o web/geomyidae.wasm ./cmd/client
Remove-Item Env:\GOOS; Remove-Item Env:\GOARCH
Pop-Location
Push-Location $PSScriptRoot\..\web
$goroot = go env GOROOT
Copy-Item $goroot\lib\wasm\wasm_exec.js .
Pop-Location
Write-Host "Serving on http://127.0.0.1:8081/"
python -m http.server 8081 -d $PSScriptRoot\..\web
