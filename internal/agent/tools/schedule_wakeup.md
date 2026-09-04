Create a timer that can be used by an agent to schedule an automatic wake up to continue an operation after a fixed delay.

Usage:
- `delay_seconds`: The number of seconds to wait before waking up (must be between 60 and 3600).
- `reason`: A short explanation of why the agent is waiting (shown in the UI).
- `prompt`: An optional prompt to inject when the agent wakes up.
- `stop`: Set to true to stop an existing wakeup.
