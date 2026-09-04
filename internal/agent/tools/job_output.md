Get stdout/stderr from a background shell by ID; set wait=true to block until completion.

<usage>
- Returns current output from a background shell started by the bash tool
- set wait=false to get current output without blocking
- set wait=true with a timeout (1-600 seconds) to block until the shell completes or the timeout elapses
- optionally set keyword together with wait=true to return as soon as that substring appears in the output, instead of waiting for the whole command to finish (useful for waiting on a server "listening on port" style log line)
</usage>

<tips>
- Use wait=false to poll for progress without risking an indefinite block
- Use wait=true with a reasonable timeout (e.g. 30s for fast commands, 120s for builds)
- Use keyword to detect readiness signals (e.g. "Listening on", "Compiled successfully") without waiting for a long-running process to exit
- If the shell is still running after the timeout, call job_output again or use job_kill to terminate it
</tips>
