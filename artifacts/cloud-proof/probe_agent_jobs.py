"""Probe the Agent Jobs API the way Star and the harness would.

Reads everything from the environment so secrets never land in files:
  DECA_AGENT_API_BASE_URL  required, e.g. http://127.0.0.1:8931 or prod host
  DECA_API_KEY             user key (Bearer)
  DECA_AGENT_WORKER_TOKEN  worker token (X-Worker-Token), optional for user-only probe
  DECA_AGENT_JOB_ID        job id for worker calls, default job-proof-1

Run locally:  DECA_AGENT_API_BASE_URL=http://127.0.0.1:8931 DECA_API_KEY=user-secret \\
                DECA_AGENT_WORKER_TOKEN=worker-secret python3 probe_agent_jobs.py
"""

from __future__ import annotations

import json
import os
import sys
import urllib.request
import urllib.error


def req(method: str, url: str, headers: dict, payload=None):
    data = json.dumps(payload).encode() if payload is not None else None
    r = urllib.request.Request(url, data=data, method=method,
                               headers={"Content-Type": "application/json", **headers})
    try:
        with urllib.request.urlopen(r, timeout=20) as resp:
            return resp.status, resp.read().decode()[:500]
    except urllib.error.HTTPError as e:
        return e.code, e.read().decode()[:500]


def main() -> int:
    base = os.environ.get("DECA_AGENT_API_BASE_URL", "").rstrip("/")
    if not base:
        print("DECA_AGENT_API_BASE_URL is required", file=sys.stderr)
        return 2
    user_key = os.environ.get("DECA_API_KEY", "")
    worker = os.environ.get("DECA_AGENT_WORKER_TOKEN", "")
    job = os.environ.get("DECA_AGENT_JOB_ID", "job-proof-1")

    ok = True
    if user_key:
        for path in ("/deca-agents/v1/jobs", "/deca/v1/models"):
            code, body = req("GET", base + path, {"Authorization": f"Bearer {user_key}"})
            print(f"user GET {path} -> {code}: {body}")
            ok = ok and code < 500
    else:
        print("skip user calls (DECA_API_KEY not set)")

    if worker:
        code, body = req("GET", f"{base}/deca-agents/v1/jobs/{job}", {"X-Worker-Token": worker})
        print(f"worker GET job -> {code}: {body}")
        ok = ok and code < 500
        code, body = req("POST", f"{base}/deca-agents/v1/jobs/{job}/events",
                         {"X-Worker-Token": worker}, {"events": [{"type": "star-probe", "from": "star-tui"}]})
        print(f"worker POST events -> {code}: {body}")
        ok = ok and code < 500
    else:
        print("skip worker calls (DECA_AGENT_WORKER_TOKEN not set)")

    print("PROBE OK" if ok else "PROBE FAILED")
    return 0 if ok else 1


if __name__ == "__main__":
    sys.exit(main())
