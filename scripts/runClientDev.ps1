Write-Host "This will run the client game in a loop so that it restarts if it is closed or crashes or if the server restarts."
Push-Location $PSScriptRoot\..\src
while ($true) {
    go run .\client
}