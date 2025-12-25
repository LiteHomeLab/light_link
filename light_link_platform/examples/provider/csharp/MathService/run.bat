@echo off
REM Skip TLS certificate verification for self-signed certificates
REM The connection is still encrypted with TLS
set DOTNET_SSL_SKIP_CERT_VALIDATION=1
dotnet run
