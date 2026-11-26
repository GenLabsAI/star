"""Probe the Deca gateway: enroll a device key, then check tool calling.

The key is kept in memory only; it is never printed or written to disk.

Usage: python3 deca_model_probe.py [base_url]
"""

import json
import urllib.error
import urllib.request

BASE = "https://api.genlabs.dev"

SIGNALS = {
    "machine_id": "star-cloud-infer",
    "host_id": "star-cloud-infer",
    "hostname": "star-cloud-infer",
    "cpu": "star-cloud-infer",
    "os": "linux",
}


def post(path, payload, key=None):
    data = json.dumps(payload).encode()
    req = urllib.request.Request(BASE + path, data=data, method="POST")
    req.add_header("Content-Type", "application/json")
    req.add_header("User-Agent", "star-cloud-probe")
    if key:
        req.add_header("Authorization", "Bearer " + key)
    try:
        with urllib.request.urlopen(req, timeout=180) as resp:
            return resp.status, resp.read()
    except urllib.error.HTTPError as e:
        return e.code, e.read()


def main():
    status, body = post("/deca/v1/opencode/enroll", {"signals": SIGNALS})
    if status != 200:
        print("enroll", status, body[:300])
        return
    key = json.loads(body)["api_key"]
    print("enroll 200, key length", len(key))

    tools = [
        {
            "type": "function",
            "function": {
                "name": "bash",
                "description": "Run a shell command.",
                "parameters": {
                    "type": "object",
                    "properties": {"command": {"type": "string"}},
                    "required": ["command"],
                },
            },
        }
    ]
    for model in ("deca-2.6-pro", "deca-2.5-pro"):
        status, body = post(
            "/deca/v1/chat/completions",
            {
                "model": model,
                "messages": [
                    {"role": "user", "content": "Use the bash tool to run `echo hi`."}
                ],
                "tools": tools,
                "max_tokens": 300,
                "stream": False,
            },
            key,
        )
        if status != 200:
            print(model, status, body[:300])
            continue
        choice = json.loads(body)["choices"][0]
        msg = choice["message"]
        calls = [
            (t["function"]["name"], t["function"]["arguments"][:60])
            for t in msg.get("tool_calls") or []
        ]
        print(
            model,
            "finish", choice.get("finish_reason"),
            "content", repr((msg.get("content") or "")[:80]),
            "tool_calls", calls,
        )


if __name__ == "__main__":
    main()
