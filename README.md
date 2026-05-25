# Codex Remote Control Shim For The Codex App For Windows

Small shim for the Codex app for Windows that makes the app-owned local app-server start with Codex remote control enabled.

## What Problem This Solves

The Codex app for Windows starts a private local app-server process. That app-server is normally launched by the app with arguments similar to:

```text
app-server --analytics-default-enabled
```

Newer Codex builds expose remote control through a hidden app-server flag:

```text
--remote-control
```

The Codex app for Windows can be pointed at a different Codex executable through `CODEX_CLI_PATH`, but that setting chooses the executable only; it does not let you append extra arguments. This shim uses that executable hook to insert `--remote-control` only when the app launches `codex app-server`.

After setup, the expected process chain is:

```text
Codex.exe
  -> codex-remote-control-shim.exe app-server --analytics-default-enabled
    -> codex.exe app-server --analytics-default-enabled --remote-control
```

## How It Works

The shim:

- receives the arguments from the Codex app for Windows,
- detects `app-server` launches,
- appends `--remote-control` if a remote-control flag is not already present,
- resolves the real Codex CLI from `%LOCALAPPDATA%\OpenAI\Codex\bin\*\codex.exe`,
- optionally uses `CODEX_REMOTE_CONTROL_REAL_CODEX` if you want to pin the exact real `codex.exe`,
- removes `CODEX_CLI_PATH` from the child process environment so it does not recursively call itself,
- places the child process in a Windows Job Object so the real `codex.exe` exits when the shim exits.

It does not modify `config.toml`, patch the Codex app for Windows, replace files in `WindowsApps`, or create a network listener by itself.

## Install With An AI Coding Agent

Copy this prompt into your AI coding agent if you want the agent to install the shim for you:

```text
Install the Codex Remote Control Shim for the Codex app for Windows from https://github.com/feketegabor/codex-windows-remote-control-shim.

Use the latest GitHub Release asset named codex-remote-control-shim-windows-amd64.exe unless I explicitly ask you to build from source. Do not download or run unrelated binaries.

Before changing anything, check the current user-level CODEX_CLI_PATH. If it is already set and does not point to $env:USERPROFILE\.codex\remote-control\codex-remote-control-shim.exe, stop and explain what it currently points to. Ask me before replacing it.

If I approve replacing CODEX_CLI_PATH, save the previous value in CODEX_CLI_PATH_BEFORE_REMOTE_CONTROL_SHIM, copy the downloaded executable to $env:USERPROFILE\.codex\remote-control\codex-remote-control-shim.exe, then set user-level CODEX_CLI_PATH to that path.

Do not modify Codex config.toml, do not patch files inside WindowsApps, do not create scheduled tasks, and do not start a separate app-server. This setup should make the Codex app for Windows use the shim the next time the app is fully restarted.

After installation, tell me exactly what path CODEX_CLI_PATH is set to and ask me to fully restart the Codex app for Windows.
```

## Requirements

- Codex app for Windows
- Go only when building from source

This is an unofficial workaround. It depends on the Codex app for Windows honoring `CODEX_CLI_PATH` and on the local Codex CLI supporting `app-server --remote-control`.

## Install From A Release

This is the simplest install path. It uses the executable built by GitHub Actions from the tagged source.

1. Open the latest release:

   [Latest release](https://github.com/feketegabor/codex-windows-remote-control-shim/releases/latest)

2. Download `codex-remote-control-shim-windows-amd64.exe`.

3. Optional: download `SHA256SUMS.txt` and compare it with the executable hash:

   ```powershell
   Get-FileHash .\codex-remote-control-shim-windows-amd64.exe -Algorithm SHA256
   ```

   The hash printed by PowerShell should match the hash listed in `SHA256SUMS.txt`.

4. Open PowerShell in the folder where you downloaded the `.exe`.

5. Create the install folder:

   ```powershell
   New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.codex\remote-control"
   ```

6. Copy the executable into the install folder:

   ```powershell
   Copy-Item .\codex-remote-control-shim-windows-amd64.exe "$env:USERPROFILE\.codex\remote-control\codex-remote-control-shim.exe"
   ```

7. Check whether you already have a user-level `CODEX_CLI_PATH`:

   ```powershell
   [Environment]::GetEnvironmentVariable("CODEX_CLI_PATH", "User")
   ```

   If this prints an existing path, decide whether you want to replace it. To keep a backup of the old value before replacing it, run:

   ```powershell
   [Environment]::SetEnvironmentVariable("CODEX_CLI_PATH_BEFORE_REMOTE_CONTROL_SHIM", [Environment]::GetEnvironmentVariable("CODEX_CLI_PATH", "User"), "User")
   ```

8. Point the Codex app for Windows to the shim:

   ```powershell
   [Environment]::SetEnvironmentVariable("CODEX_CLI_PATH", "$env:USERPROFILE\.codex\remote-control\codex-remote-control-shim.exe", "User")
   ```

9. Fully restart the Codex app for Windows.

## Build From Source

Use this path if you prefer to compile the executable yourself instead of downloading the release build.

1. Install Go if you do not already have it.

2. Clone this repository:

   ```powershell
   git clone https://github.com/feketegabor/codex-windows-remote-control-shim.git
   ```

3. Enter the repository folder:

   ```powershell
   cd codex-windows-remote-control-shim
   ```

4. Create the install folder:

   ```powershell
   New-Item -ItemType Directory -Force -Path "$env:USERPROFILE\.codex\remote-control"
   ```

5. Build the shim into that folder:

   ```powershell
   go build -trimpath -ldflags="-s -w" -o "$env:USERPROFILE\.codex\remote-control\codex-remote-control-shim.exe" .
   ```

6. Check whether you already have a user-level `CODEX_CLI_PATH`:

   ```powershell
   [Environment]::GetEnvironmentVariable("CODEX_CLI_PATH", "User")
   ```

   If this prints an existing path, decide whether you want to replace it. To keep a backup of the old value before replacing it, run:

   ```powershell
   [Environment]::SetEnvironmentVariable("CODEX_CLI_PATH_BEFORE_REMOTE_CONTROL_SHIM", [Environment]::GetEnvironmentVariable("CODEX_CLI_PATH", "User"), "User")
   ```

7. Point the Codex app for Windows to the shim:

   ```powershell
   [Environment]::SetEnvironmentVariable("CODEX_CLI_PATH", "$env:USERPROFILE\.codex\remote-control\codex-remote-control-shim.exe", "User")
   ```

8. Fully restart the Codex app for Windows.

If your endpoint protection flags locally built Go executables, review the source, build from a trusted environment, or allow-list your own build output according to your organization's security policy. Do not run random binaries from the internet.

## Scripted Source Install

The repository also includes `install.ps1`, which builds from source and sets `CODEX_CLI_PATH`.

Use it only after reviewing what it does. First clone the repository and enter the repository folder:

```powershell
git clone https://github.com/feketegabor/codex-windows-remote-control-shim.git
```

```powershell
cd codex-windows-remote-control-shim
```

Then run:

```powershell
.\install.ps1
```

If `CODEX_CLI_PATH` is already set, the installer stops instead of overwriting it. Review the current value first, then either set `CODEX_CLI_PATH` manually or rerun:

```powershell
.\install.ps1 -Force
```

When `-Force` replaces an existing value, the installer saves the previous value in `CODEX_CLI_PATH_BEFORE_REMOTE_CONTROL_SHIM`.

## Uninstall

To uninstall, restore `CODEX_CLI_PATH` to what it was before installing the shim, or clear it if there was no previous value.

1. Close the Codex app for Windows.

2. Check whether the installer saved a previous value:

   ```powershell
   [Environment]::GetEnvironmentVariable("CODEX_CLI_PATH_BEFORE_REMOTE_CONTROL_SHIM", "User")
   ```

3. If the command prints a previous path, restore it:

   ```powershell
   [Environment]::SetEnvironmentVariable("CODEX_CLI_PATH", [Environment]::GetEnvironmentVariable("CODEX_CLI_PATH_BEFORE_REMOTE_CONTROL_SHIM", "User"), "User")
   ```

4. Clear the saved backup value after restoring it:

   ```powershell
   [Environment]::SetEnvironmentVariable("CODEX_CLI_PATH_BEFORE_REMOTE_CONTROL_SHIM", $null, "User")
   ```

5. If there was no previous value, only clear `CODEX_CLI_PATH` if it currently points to the shim:

   ```powershell
   if ([Environment]::GetEnvironmentVariable("CODEX_CLI_PATH", "User") -eq "$env:USERPROFILE\.codex\remote-control\codex-remote-control-shim.exe") { [Environment]::SetEnvironmentVariable("CODEX_CLI_PATH", $null, "User") }
   ```

6. Optionally delete the shim executable:

   ```powershell
   Remove-Item "$env:USERPROFILE\.codex\remote-control\codex-remote-control-shim.exe"
   ```

7. Restart the Codex app for Windows.

If you want a single copy-paste uninstall script instead, use this:

```powershell
$previous = [Environment]::GetEnvironmentVariable("CODEX_CLI_PATH_BEFORE_REMOTE_CONTROL_SHIM", "User")
$shim = "$env:USERPROFILE\.codex\remote-control\codex-remote-control-shim.exe"
$current = [Environment]::GetEnvironmentVariable("CODEX_CLI_PATH", "User")
if ($previous) {
  [Environment]::SetEnvironmentVariable("CODEX_CLI_PATH", $previous, "User")
  [Environment]::SetEnvironmentVariable("CODEX_CLI_PATH_BEFORE_REMOTE_CONTROL_SHIM", $null, "User")
} elseif ($current -eq $shim) {
  [Environment]::SetEnvironmentVariable("CODEX_CLI_PATH", $null, "User")
} else {
  Write-Host "CODEX_CLI_PATH does not point to this shim; leaving it unchanged."
}
Remove-Item $shim -ErrorAction SilentlyContinue
```

## Security Notes

- The shim is small enough to audit directly.
- The shim itself does not bind a TCP port or expose a WebSocket listener.
- It only changes app-server launches; other Codex CLI commands are forwarded unchanged.
- It enables Codex app-server remote-control behavior. The authentication, authorization, pairing, and network behavior of remote control belong to the Codex/OpenAI implementation in the Codex version you are running, not to this shim.
- Review the Codex app-server behavior for your installed version before enabling this on a machine with sensitive access.
- This relies on an internal/hidden flag and may break when Codex changes its desktop launch path or app-server flags.

## Publishing Releases

Maintainers publish a release by pushing a semver-style tag:

```powershell
git tag v0.1.0
git push origin v0.1.0
```

The GitHub Actions workflow builds `codex-remote-control-shim-windows-amd64.exe`, creates `SHA256SUMS.txt`, and uploads both files to the matching GitHub Release.

## License

MIT
