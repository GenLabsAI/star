"""Mock Agent Jobs API implementing the harness runner.py contract.

Implements just enough of the genlabsapi surface for the proof:
  POST /deca-agents/v1/jobs/<id>/events   (X-Worker-Token)
  PATCH /deca-agents/v1/jobs/<id>         (X-Worker-Token)
  GET /deca-agents/v1/jobs                (Bearer user key)
  GET /deca-agents/v1/jobs/<id>           (either auth)

Run: python3 mock_agent_jobs.py [port]
Env: MOCK_WORKER_TOKEN, MOCK_USER_KEY (defaults: worker-secret, user-secret)
"""

from __future__ import annotations

import json
import os
import sys
from http.server import BaseHTTPRequestHandler, HTTPServer

WORKER_TOKEN = os.environ.get("MOCK_WORKER_TOKEN", "worker-secret")
USER_KEY = os.environ.get("MOCK_USER_KEY", "user-secret")
JOBS: dict = {"job-proof-1": {"id": "job-proof-1", "status": "running"}}


class Handler(BaseHTTPRequestHandler):
    def log_message(self, *args):
        pass

    def _send(self, code: int, payload: dict):
        body = json.dumps(payload).encode()
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def _auth(self):
        if self.headers.get("X-Worker-Token") == WORKER_TOKEN:
            return "worker"
        auth = self.headers.get("Authorization", "")
        if auth == f"Bearer {USER_KEY}":
            return "user"
        return None

    def do_GET(self):
        role = self._auth()
        if role is None:
            return self._send(401, {"error": "unauthorized"})
        if self.path == "/deca-agents/v1/jobs":
            return self._send(200, {"jobs": list(JOBS.values())})
        if self.path.startswith("/deca-agents/v1/jobs/"):
            job_id = self.path.rsplit("/", 1)[-1]
            job = JOBS.get(job_id)
            if job is None:
                return self._send(404, {"error": "unknown job"})
            return self._send(200, {"job": job})
        if self.path == "/deca/v1/models":
            return self._send(200, {"models": ["deca-test"]})
        return self._send(404, {"error": "not found"})

    def do_PATCH(self):
        role = self._auth()
        if role != "worker":
            return self._send(401 if role is None else 403, {"error": "worker token required"})
        if self.path.startswith("/deca-agents/v1/jobs/"):
            length = int(self.headers.get("Content-Length", 0))
            payload = json.loads(self.rfile.read(length) or b"{}")
            job_id = self.path.rsplit("/", 1)[-1]
            job = JOBS.setdefault(job_id, {"id": job_id})
            job.update(payload)
            return self._send(200, {"job": job})
        return self._send(404, {"error": "not found"})

    def do_POST(self):
        role = self._auth()
        if role != "worker":
            return self._send(401 if role is None else 403, {"error": "worker token required"})
        if self.path.endswith("/events"):
            length = int(self.headers.get("Content-Length", 0))
            payload = json.loads(self.rfile.read(length) or b"{}")
            events = payload.get("events", [])
            return self._send(200, {"received": len(events)})
        return self._send(404, {"error": "not found"})


if __name__ == "__main__":
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8931
    print(f"mock agent jobs on http://127.0.0.1:{port}", flush=True)
    HTTPServer(("127.0.0.1", port), Handler).serve_forever()
