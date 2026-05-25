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
- Go, if building from source

This is an unofficial workaround. It depends on Codex Desktop honoring `CODEX_CLI_PATH` and on the local Codex CLI supporting `app-server --remote-control`.

## Install

Clone the repo and build the shim:

```powershell
git clone https://github.com/feketegabor/codex-windows-remote-control-shim.git
cd codex-windows-remote-control-shim
.\install.ps1
```

Then fully restart Codex Desktop.

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

Clear the user environment variable and restart Codex Desktop:

```powershell
[Environment]::SetEnvironmentVariable("CODEX_CLI_PATH", $null, "User")
```

You can then delete the shim executable from `%USERPROFILE%\.codex\remote-control`.

## Security notes

- The shim is small enough to audit directly.
- It does not bind a TCP port or expose a WebSocket listener.
- It only changes app-server launches; other Codex CLI commands are forwarded unchanged.
- Remote-control access is still governed by Codex/OpenAI account and remote-control pairing behavior.
- This relies on an internal/hidden flag and may break when Codex changes its desktop launch path or app-server flags.

## License

MIT
