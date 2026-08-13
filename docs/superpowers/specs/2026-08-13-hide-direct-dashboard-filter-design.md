# Hide Direct Dashboard Filter Design

## Goal

Allow users to exclude `DIRECT` traffic from the dashboard without changing collection, persistence, or permanent accounting.

## Interaction

- Add a compact `隐藏直连` switch to the dashboard toolbar near the time controls.
- The default for a first-time user is off, so the existing all-traffic view remains unchanged.
- Persist the last switch state in browser `localStorage` and restore it on the next visit.
- Changing the switch immediately reloads the current dashboard while preserving the selected mode, dimension, and time range.

## Filtering semantics

- When the switch is off, queries include every recorded route type exactly as they do today.
- When the switch is on, dashboard queries add `route_type <> 'DIRECT'`.
- The filter affects all visible traffic results: total/upload/download cards, trend, primary ranking, secondary drilldown, and connection details.
- `REJECT` and `PROXY` remain visible. Only `DIRECT` is excluded.
- Filtering occurs in backend queries before aggregation so mixed direct/proxy traffic for one process or host is calculated correctly.

## Data integrity

- The collector continues recording all Mihomo-managed connections.
- Minute facts, connection sessions, and hour/day/week/month rollups retain direct traffic permanently.
- Enabling or disabling the switch performs no database writes and never deletes historical data.

## API

- Dashboard traffic endpoints accept an optional `excludeDirect=1` query parameter.
- Missing, empty, or other values preserve the existing include-all behavior.
- The parameter is applied consistently to detail and summary query paths, including the in-memory aggregate buffer.

## Verification

- Backend tests use mixed `DIRECT` and `PROXY` rows and prove that the parameter changes only returned totals.
- Frontend contract tests prove the switch exists, persists its last state, and sends the parameter to every dashboard traffic request.
- Existing behavior remains covered with the filter omitted.
- Visually verify both switch states in the running dashboard before rebuilding the installer.
