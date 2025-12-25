# Skip TLS certificate verification for self-signed certificates
# The connection is still encrypted with TLS
$env:DOTNET_SSL_SKIP_CERT_VALIDATION = "1"

# Change to script directory
$scriptPath = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptPath

# Run the project
dotnet run
