param(
    [string] $InstallDir = "$env:USERPROFILE\.codex\remote-control",
    [string] $OutputName = "codex-remote-control-shim.exe"
)

$ErrorActionPreference = 'Stop'

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "Go is required to build this shim. Install Go, then rerun this script."
}

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$outputPath = Join-Path $InstallDir $OutputName
go build -trimpath -ldflags="-s -w" -o $outputPath .

[Environment]::SetEnvironmentVariable('CODEX_CLI_PATH', $outputPath, 'User')

Write-Host "Built shim: $outputPath"
Write-Host "Set user CODEX_CLI_PATH to: $outputPath"
Write-Host "Restart Codex Desktop so it inherits the new environment variable."
