"""Local verification for the relay _pump fix, using the REAL server code.

Runs the actual `flask_relay.py` from the genlabsapi checkout (patched
branch) under Flask/flask-sock, plus a stub model endpoint with an
artificial delay, then Star and the bridge talk through it.

Usage: local_relay_check.py <flask_relay_dir> <port>
Registers session "local-1" and serves forever.
"""

from __future__ import annotations

import sys
import time

from flask import Flask, jsonify, request

sys.path.insert(0, sys.argv[1])

from flask_relay import handle_browser_ws, handle_runner_ws, register_session  # noqa: E402
from flask_sock import Sock  # noqa: E402

app = Flask(__name__)
sock = Sock(app)


@sock.route("/terminal/browser/<session_id>")
def browser(ws, session_id):
    handle_browser_ws(ws, session_id)


@sock.route("/terminal/runner/<session_id>")
def runner(ws, session_id):
    handle_runner_ws(ws, session_id)


@app.route("/stub/v1/chat/completions", methods=["POST"])
def stub_completion():
    time.sleep(6)  # slower than the old 1s pump death: proves the fix
    body = request.get_json(force=True)
    prompt = (body.get("messages") or [{}])[-1].get("content", "")
    return jsonify({"choices": [{"message": {"content": f"LOCAL-RELAY-OK:{prompt[:24]}"},
                                 "finish_reason": "stop"}]})


register_session("local-1", "local@test", "tok")
app.run(host="127.0.0.1", port=int(sys.argv[2]), threaded=True)
