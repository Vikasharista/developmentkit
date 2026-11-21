#!/usr/bin/env python3
"""Walk a Jira work item through its workflow one hop at a time, by status name.

Transition IDs collide across issue types, so every hop is resolved by the
*name of the target status* from the live transition set, fired by the ID that
set reports, and read back before the next hop.

Two twg backends, picked at runtime:
    api     twg builds that still have the `twg api jira:<path>` REST escape
            hatch; preferred when present, since it is the path proven there
    native  twg without `api` (1.3.6): `jira workitem get --fields status`,
            `jira workitem transitions query`, `jira workitem transition`.
            Reads verified live on 1.3.6; the fire step not yet.
Set JIRA_WALK_BACKEND=native|api to force one.

Usage:
    jira-walk.py status      <KEY>
    jira-walk.py transitions <KEY>
    jira-walk.py walk        <KEY> "<status>" ["<status>" ...] [--dry-run]

Exit codes:
    0  every requested hop landed (or the item already sat on the target)
    1  a hop was not offered from the current status, or a fired hop read back
       a different status; nothing after it fired
    2  twg call failed (auth, network, bad key, non-JSON output)
    3  usage error

Stdlib only. Requires `twg` on PATH with a working auth profile.
"""

from __future__ import annotations

import json
import os
import subprocess
import sys

TWG = "twg"
_backend: str | None = None


class TwgError(RuntimeError):
    """A twg call failed or returned something that is not JSON."""


def _run(cmd: list[str], stdin: str | None = None, parse: bool = True) -> dict:
    try:
        proc = subprocess.run(cmd, input=stdin, capture_output=True, text=True)
    except OSError as exc:
        raise TwgError(f"cannot run {cmd[0]}: {exc}") from exc
    if proc.returncode != 0:
        detail = proc.stderr.strip() or proc.stdout.strip()
        raise TwgError(f"{' '.join(cmd)} exited {proc.returncode}: {detail}")
    out = proc.stdout.strip()
    if not out or not parse:
        return {}
    try:
        return json.loads(out)
    except json.JSONDecodeError as exc:
        raise TwgError(f"{' '.join(cmd)} returned non-JSON: {out[:200]}") from exc


def _has_command(*path: str) -> bool:
    try:
        proc = subprocess.run([TWG, *path, "--help"], capture_output=True, text=True)
    except OSError as exc:
        raise TwgError(f"cannot run {TWG}: {exc}") from exc
    return proc.returncode == 0 and "unknown command" not in proc.stderr


def backend() -> str:
    """'api' when twg still has `api`, else 'native' when it has `transitions query`."""
    global _backend
    if _backend is None:
        forced = os.environ.get("JIRA_WALK_BACKEND")
        if forced in ("native", "api"):
            _backend = forced
        elif _has_command("api"):
            _backend = "api"
        elif _has_command("jira", "workitem", "transitions", "query"):
            _backend = "native"
        else:
            raise TwgError(
                "installed twg has neither `jira workitem transitions query` nor `api`"
            )
    return _backend


def _native(*args: str, parse: bool = True) -> dict:
    return _run([TWG, *args, "-o", "json", "--output-summary", "none"], parse=parse)


def _api(path: str, method: str = "GET", body: dict | None = None) -> dict:
    cmd = [TWG, "api", f"jira:{path}"]
    if method != "GET":
        cmd += ["--method", method]
    if body is not None:
        cmd += ["--input", "-"]
    return _run(cmd, json.dumps(body) if body is not None else None)


def status(key: str) -> str:
    """Current status name of the work item."""
    if backend() == "native":
        data = _native("jira", "workitem", "get", key, "--fields", "status")
        items = data.get("data") or []
        if isinstance(items, dict):
            items = [items]
        if not items:
            raise TwgError(f"twg jira workitem get {key} returned no item")
        item = items[0]
        # twg flattens REST `fields` onto the item; accept both shapes.
        return (item.get("fields") or item)["status"]["name"]
    data = _api(f"/rest/api/3/issue/{key}?fields=status")
    return data["fields"]["status"]["name"]


def transitions(key: str) -> list[dict]:
    """Transitions offered from the current status: [{id, name, to}]."""
    if backend() == "native":
        data = _native("jira", "workitem", "transitions", "query", "--id", key)
        return [
            {"id": str(t["id"]), "name": t["name"], "to": t["toName"]}
            for t in (data.get("data") or {}).get("transitions", [])
        ]
    data = _api(f"/rest/api/3/issue/{key}/transitions")
    return [
        {"id": t["id"], "name": t["name"], "to": t["to"]["name"]}
        for t in data.get("transitions", [])
    ]


def fire(key: str, transition_id: str) -> None:
    """Fire one transition by the ID the live transition set reported."""
    if backend() == "native":
        # Exit code is the signal; the read-back in walk() verifies the landing.
        _native(
            "jira", "workitem", "transition", "--id", key, "--transition-id", transition_id,
            parse=False,
        )
        return
    _api(
        f"/rest/api/3/issue/{key}/transitions",
        method="POST",
        body={"transition": {"id": transition_id}},
    )


def walk(key: str, *targets: str, dry_run: bool = False) -> int:
    """Walk KEY through each target status in order. Returns an exit code.

    A target equal to the current status is a no-op, not an error. A target
    not reachable in one hop stops the walk with exit 1 and fires nothing
    further. Each fired hop is read back; a mismatch also stops with exit 1.
    With dry_run, only the first pending hop is checked against the live
    transition set (Jira offers transitions from the live status only); the
    remaining targets are listed as unverified and nothing is fired.
    """
    current = status(key)
    print(f"{key}: at {current!r} (twg backend: {backend()})", flush=True)
    for target in targets:
        if current == target:
            print(f"{key}: already at {target!r}, skipping")
            continue
        offered = transitions(key)
        matches = [t for t in offered if t["to"] == target]
        match = matches[0] if matches else None
        if len(matches) > 1:
            print(
                f"{key}: {len(matches)} transitions lead to {target!r}; using the first: "
                + ", ".join(f"{t['id']} {t['name']!r}" for t in matches),
                file=sys.stderr,
            )
        if match is None:
            names = ", ".join(repr(t["to"]) for t in offered) or "(none)"
            print(
                f"{key}: no transition from {current!r} to {target!r}; offered: {names}",
                file=sys.stderr,
            )
            return 1
        if dry_run:
            print(f"{key}: would fire {match['id']} {match['name']!r} -> {target!r}")
            remaining = targets[targets.index(target) + 1 :]
            if remaining:
                # Jira only reports transitions out of the *live* status, so
                # later hops cannot be checked without firing the first one.
                print(
                    f"{key}: dry run stops here; unverified later hops: "
                    + ", ".join(repr(t) for t in remaining)
                )
            return 0
        fire(key, match["id"])
        print(f"{key}: fired {match['id']} {match['name']!r} toward {target!r}", flush=True)
        landed = status(key)
        if landed != target:
            print(
                f"{key}: fired {match['id']} {match['name']!r} but read back {landed!r}, expected {target!r}",
                file=sys.stderr,
            )
            return 1
        print(f"{key}: {current!r} -> {landed!r} via {match['id']} {match['name']!r}")
        current = landed
    return 0


def main(argv: list[str]) -> int:
    dry_run = "--dry-run" in argv
    argv = [a for a in argv if a != "--dry-run"]
    if len(argv) < 2:
        print(__doc__, file=sys.stderr)
        return 3
    verb, key, rest = argv[0], argv[1], argv[2:]
    try:
        if verb == "status" and not rest:
            print(status(key))
            return 0
        if verb == "transitions" and not rest:
            for t in transitions(key):
                print(f"{t['id']}\t{t['name']}\t-> {t['to']}")
            return 0
        if verb == "walk" and rest:
            return walk(key, *rest, dry_run=dry_run)
    except TwgError as exc:
        print(f"twg error: {exc}", file=sys.stderr)
        return 2
    except (KeyError, TypeError, AttributeError, IndexError) as exc:
        print(f"twg error: unexpected response shape ({exc!r})", file=sys.stderr)
        return 2
    print(__doc__, file=sys.stderr)
    return 3


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
