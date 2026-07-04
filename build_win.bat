@echo off
setlocal enabledelayedexpansion

for /f "delims=" %%i in ('git describe --tags --exact-match 2^>nul') do set "tag=%%i"
for /f "delims=" %%i in ('git rev-parse --short HEAD 2^>nul') do set "commit=%%i"
for /f "delims=" %%i in ('go list -m') do set "mod=%%i"

set "version=%tag%"
if defined tag if "!version:~0,1!"=="v" set "version=!version:~1!"

if not exist .\dist mkdir .\dist

go build ^
    -ldflags "-X %mod%/internal/buildinfo.Version=%version% -X %mod%/internal/buildinfo.Commit=%commit%" ^
    -o .\dist\boring.exe ^
    .\cmd\boring
