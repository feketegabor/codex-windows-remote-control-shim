# Codex Windows Remote Control Shim

Small Windows shim for Codex Desktop that makes the Desktop-owned local app-server start with Codex remote control enabled.

## What problem this solves

Codex Desktop starts a private local app-server process. On Windows, that app-server is normally launched by Desktop with arguments similar to:

```text
app-server --analytics-default-enabled
```

Newer Codex builds expose remote control through a hidden app-server flag:

```text
--remote-control
```

The desktop app can be pointed at a different Codex executable through `CODEX_CLI_PATH`, but that setting chooses the executable only; it does not let you append extra arguments. This shim uses that executable hook to insert `--remote-control` only when Desktop launches `codex app-server`.

After setup, the expected process chain is:

```text
Codex.exe
  -> codex-remote-control-shim.exe app-server --analytics-default-enabled
    -> codex.exe app-server --analytics-default-enabled --remote-control
```

## How it works

The shim:

- receives the arguments from Codex Desktop,
- detects `app-server` launches,
- appends `--remote-control` if a remote-control flag is not already present,
- resolves the real Codex CLI from `%LOCALAPPDATA%\OpenAI\Codex\bin\*\codex.exe`,
- optionally uses `CODEX_REMOTE_CONTROL_REAL_CODEX` if you want to pin the exact real `codex.exe`,
- removes `CODEX_CLI_PATH` from the child process environment so it does not recursively call itself,
- places the child process in a Windows Job Object so the real `codex.exe` exits when the shim exits.

It does not modify `config.toml`, patch Codex Desktop, replace files in `WindowsApps`, or create a network listener by itself.

## Requirements

- Windows
- Codex Desktop
- Go only when building from source

This is an unofficial workaround. It depends on Codex Desktop honoring `CODEX_CLI_PATH` and on the local Codex CLI supporting `app-server --remote-control`.

## Install from a release

Download `codex-remote-control-shim-windows-amd64.exe` from the latest GitHub Release, then install it somewhere stable and point `CODEX_CLI_PATH` to it:

```powershell
$installDir = "$env:USERPROFILE\.codex\remote-control"
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
$target = "$installDir\codex-remote-control-shim.exe"

$existing = [Environment]::GetEnvironmentVariable("CODEX_CLI_PATH", "User")
if ($existing -and $existing -ne $target) {
  throw "CODEX_CLI_PATH is already set to '$existing'. Review that value before replacing it."
}

Copy-Item .\codex-remote-control-shim-windows-amd64.exe $target

[Environment]::SetEnvironmentVariable(
  "CODEX_CLI_PATH",
  $target,
  "User"
)
```

Then fully restart Codex Desktop.

Each release also includes `SHA256SUMS.txt` so you can verify the downloaded executable:

```powershell
$actual = (Get-FileHash .\codex-remote-control-shim-windows-amd64.exe -Algorithm SHA256).Hash.ToLowerInvariant()
$expected = (Get-Content .\SHA256SUMS.txt | Select-String "codex-remote-control-shim-windows-amd64.exe").Line.Split(" ")[0]

if ($actual -ne $expected) {
  throw "Checksum mismatch. Expected $expected but got $actual."
}
```

## Build from source

Clone the repo and build the shim:

```powershell
git clone https://github.com/feketegabor/codex-windows-remote-control-shim.git
cd codex-windows-remote-control-shim
.\install.ps1
```

Then fully restart Codex Desktop.

If `CODEX_CLI_PATH` is already set, the installer stops instead of overwriting it. Review the current value first, then either set `CODEX_CLI_PATH` manually or rerun:

```powershell
.\install.ps1 -Force
```

When `-Force` replaces an existing value, the installer saves the previous value in `CODEX_CLI_PATH_BEFORE_REMOTE_CONTROL_SHIM`.

If your endpoint protection flags locally built Go executables, review the source, build from a trusted environment, or allow-list your own build output according to your organization's security policy. Do not run random binaries from the internet.

## Manual install

```powershell
git clone https://github.com/feketegabor/codex-windows-remote-control-shim.git
cd codex-windows-remote-control-shim

$installDir = "$env:USERPROFILE\.codex\remote-control"
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

go build -trimpath -ldflags="-s -w" -o "$installDir\codex-remote-control-shim.exe" .
[Environment]::SetEnvironmentVariable(
  "CODEX_CLI_PATH",
  "$installDir\codex-remote-control-shim.exe",
  "User"
)
```

Restart Codex Desktop after setting the environment variable.

## Verify

After restarting Codex Desktop, run:

```powershell
Get-CimInstance Win32_Process |
  Where-Object {
    $_.Name -in @("codex-remote-control-shim.exe", "codex.exe") -and
    $_.CommandLine -like "*app-server*"
  } |
  Select-Object ProcessId, ParentProcessId, Name, CommandLine
```

You should see the shim process and a child `codex.exe` process whose command line includes:

```text
--remote-control
```

## Uninstall

Restore the previous `CODEX_CLI_PATH` value if the installer saved one:

```powershell
$previous = [Environment]::GetEnvironmentVariable(
  "CODEX_CLI_PATH_BEFORE_REMOTE_CONTROL_SHIM",
  "User"
)

if ($previous) {
  [Environment]::SetEnvironmentVariable("CODEX_CLI_PATH", $previous, "User")
  [Environment]::SetEnvironmentVariable(
    "CODEX_CLI_PATH_BEFORE_REMOTE_CONTROL_SHIM",
    $null,
    "User"
  )
} else {
  [Environment]::SetEnvironmentVariable("CODEX_CLI_PATH", $null, "User")
}
```

Then restart Codex Desktop.

If you installed manually and know there was no previous `CODEX_CLI_PATH`, you can clear it directly:

```powershell
[Environment]::SetEnvironmentVariable("CODEX_CLI_PATH", $null, "User")
```

You can then delete the shim executable from `%USERPROFILE%\.codex\remote-control`.

## Security notes

- The shim is small enough to audit directly.
- The shim itself does not bind a TCP port or expose a WebSocket listener.
- It only changes app-server launches; other Codex CLI commands are forwarded unchanged.
- It enables Codex app-server remote-control behavior. The authentication, authorization, pairing, and network behavior of remote control belong to the Codex/OpenAI implementation in the Codex version you are running, not to this shim.
- Review the Codex app-server behavior for your installed version before enabling this on a machine with sensitive access.
- This relies on an internal/hidden flag and may break when Codex changes its desktop launch path or app-server flags.

## Publishing releases

Maintainers publish a release by pushing a semver-style tag:

```powershell
git tag v0.1.0
git push origin v0.1.0
```

The GitHub Actions workflow builds `codex-remote-control-shim-windows-amd64.exe`, creates `SHA256SUMS.txt`, and uploads both files to the matching GitHub Release.

## License

MIT
