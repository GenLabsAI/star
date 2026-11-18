Create a timer that can be used by an agent to schedule an automatic wake up to continue an operation after a fixed delay.

One-shot usage:
- `delay_seconds`: The number of seconds to wait before waking up (must be between 60 and 3600).
- `reason`: A short explanation of why the agent is waiting (shown in the UI).
- `prompt`: An optional prompt to inject when the agent wakes up.
- `stop`: Set to true to stop an existing wakeup.

Repeating and cron usage (durable: survives restarts):
- `repeat_seconds`: Repeat the wakeup every N seconds (60 to 2592000). When `delay_seconds` is omitted the first fire is one interval out.
- `cron`: Standard five-field cron (minute hour day month weekday), or `@hourly`, `@daily`, `@weekly`, `@every <duration>` (e.g. `@every 30m`). Cannot be combined with `delay_seconds` or `daily_at`.
- `daily_at`: Local wall-clock time `HH:MM` to wake up every day (e.g. `08:00`). Cannot be combined with `delay_seconds` or `cron`.

Event-driven usage:
- `event`: A named event (e.g. `ci-failed`) that wakes the agent when it fires. With no timer fields the wakeup stays armed until something triggers the event (`star wakeup trigger --event <name>`). Combined with timer fields, whichever comes first wins.

Cloud usage:
- `cloud`: Set to true when the session runs in cloud mode, so the wakeup runs on the cloud machine instead of locally.

Long-horizon usage (hours to days out):
- Schedules are durable, so no process needs to stay alive until the fire time. A short-lived sweeper (`star wakeup run`, every few minutes via cron/systemd or the `wakeup-sweep` GitHub workflow) reaps whatever is due and exits in seconds when nothing is.
- A wakeup 24h out simply waits through the no-op sweeps and fires within one sweep interval of its due time. Individual sweeper runs never approach the GitHub Actions 6h job limit.
- Event schedules stay armed indefinitely until `star wakeup trigger --event <name>` fires them (e.g. from a CI workflow via `repository_dispatch` type `wakeup-event`).
