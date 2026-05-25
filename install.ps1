param(
    [string] $InstallDir = "$env:USERPROFILE\.codex\remote-control",
    [string] $OutputName = "codex-remote-control-shim.exe",
    [switch] $Force
)

$ErrorActionPreference = 'Stop'

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "Go is required to build this shim. Install Go, then rerun this script."
}

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$outputPath = Join-Path $InstallDir $OutputName
go build -trimpath -ldflags="-s -w" -o $outputPath .

$existing = [Environment]::GetEnvironmentVariable('CODEX_CLI_PATH', 'User')
if ($existing -and $existing -ne $outputPath -and -not $Force) {
    throw "User CODEX_CLI_PATH is already set to '$existing'. Re-run with -Force to replace it, or set CODEX_CLI_PATH manually after reviewing the existing value."
}

if ($existing -and $existing -ne $outputPath) {
    [Environment]::SetEnvironmentVariable('CODEX_CLI_PATH_BEFORE_REMOTE_CONTROL_SHIM', $existing, 'User')
    Write-Host "Saved previous CODEX_CLI_PATH to CODEX_CLI_PATH_BEFORE_REMOTE_CONTROL_SHIM."
}

[Environment]::SetEnvironmentVariable('CODEX_CLI_PATH', $outputPath, 'User')

Write-Host "Built shim: $outputPath"
Write-Host "Set user CODEX_CLI_PATH to: $outputPath"
Write-Host "Restart Codex Desktop so it inherits the new environment variable."
