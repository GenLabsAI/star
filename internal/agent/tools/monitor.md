# monitor

Runs a background command and streams its output back to the agent in real time.

Use this tool when you need to continuously monitor a long-running process, tail logs, or listen for events. Instead of polling with `job_output`, the monitor tool captures each new line of output and automatically injects it into the conversation as a new message, waking you up so you can react immediately.

## Usage Notes

- The tool returns immediately with a Monitor ID.
- The background command continues to run and stream output until terminated or the session ends.
- Output lines are batched slightly to avoid flooding the conversation.
- You can use the `job_kill` tool to terminate the monitor using its Monitor ID when you are done.
- This is best for open-ended, continuous processes (like `tail -f`, a server with request logs, or a watcher).
