# Application Autostart Toggle Design

## Goal

Let users enable or disable Windows sign-in startup from the application's existing settings panel.

## Source of truth

- Use `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`, value name `ClashTrafficMonitor`.
- Read the registry whenever settings are loaded; do not store a duplicate state in SQLite or browser storage.
- Enabling writes the current executable path as a quoted command.
- Disabling removes the value. A missing value means disabled.
- The installer continues using the same value name and key, so installer and application settings remain synchronized.

## API and interface

- Add `GET /api/settings/autostart` returning `{ "enabled": true|false }`.
- Add `PUT /api/settings/autostart` accepting the same shape and immediately applying it.
- Add `开机自动启动` to the existing settings form.
- Opening settings shows the actual registry state. Saving settings applies the selected state together with existing Mihomo, grouping, and retention settings.
- Changing this setting does not stop the current process and does not affect traffic collection or history.

## Portability and errors

- Windows implements registry access through `golang.org/x/sys/windows/registry`.
- Non-Windows builds report autostart as unavailable/disabled and reject attempts to enable it, preserving compilation of inherited server targets.
- Registry errors return an API error and leave the settings panel open with the existing status message behavior.

## Verification

- Unit tests substitute an in-memory autostart backend and verify GET, enable, disable, and error responses without changing the real registry.
- Windows-specific tests verify command quoting and value-name/key constants.
- Embedded frontend tests verify the switch, API load/save calls, and state synchronization.
- Live acceptance toggles the setting off and on and checks the actual HKCU value after each operation, ending in the user's selected enabled state.
