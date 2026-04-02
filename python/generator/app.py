#!/usr/bin/env python3
import sys
import json
from model_stub import generate_batch

def process_line(line: str):
    try:
        payload = json.loads(line)
    except json.JSONDecodeError:
        return
    if payload.get("op") != "generate":
        return
    batch = payload.get("batch", [])
    prompts = [f"Enhance image {i}" for i in range(len(batch))]
    outputs = generate_batch(prompts, batch)
    out = {"status": "ok", "outputs": outputs}
    print(json.dumps(out), flush=True)

def main():
    for line in sys.stdin:
        if not line.strip():
            continue
        process_line(line)

if __name__ == "__main__":
    main()
