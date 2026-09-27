#!/usr/bin/env python3
"""Frozen golangci-lint baseline.

Current findings must be covered by controller/.golangci-baseline.txt
(one path<TAB>code per allowed occurrence). A higher count fails.
A finding that is gone may leave a stale line. This script never
rewrites the file.
"""

from __future__ import annotations

import argparse
import json
import re
import subprocess
import sys
import tempfile
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DEFAULT_MODULE = ROOT / "controller"
DEFAULT_CONFIG = ROOT / "controller" / ".golangci.yml"
DEFAULT_BASELINE = ROOT / "controller" / ".golangci-baseline.txt"
REQUIRED_VERSION = "2.13.2"
CODE_RE = re.compile(r"^([A-Z]+[0-9]+):")


def refuse_rewrite(argv: list[str]) -> None:
    blocked = {"--write-baseline", "--create-baseline", "--update-baseline", "-cb"}
    if blocked.intersection(argv):
        print("refusing to rewrite the golangci baseline", file=sys.stderr)
        raise SystemExit(1)


def rel_path(filename: str, module: Path) -> str:
    path = Path(filename)
    if path.is_absolute():
        try:
            path = path.relative_to(module)
        except ValueError:
            return path.as_posix()
    return path.as_posix()


def finding_code(issue: dict) -> str:
    text = issue.get("Text") or ""
    first = text.splitlines()[0].strip() if text else ""
    match = CODE_RE.match(first)
    if match:
        return match.group(1)
    return str(issue.get("FromLinter") or "")


def require_golangci() -> None:
    try:
        probe = subprocess.run(
            ["golangci-lint", "version"],
            capture_output=True,
            text=True,
            check=False,
        )
    except OSError as exc:
        print(f"golangci-lint is not installed: {exc}", file=sys.stderr)
        raise SystemExit(2) from exc
    text = f"{probe.stdout}\n{probe.stderr}"
    if probe.returncode != 0 or REQUIRED_VERSION not in text:
        print(
            f"golangci-lint {REQUIRED_VERSION} is required so the baseline stays stable",
            file=sys.stderr,
        )
        raise SystemExit(2)


def run_golangci(module: Path, config: Path) -> list[dict]:
    require_golangci()
    handle = tempfile.NamedTemporaryFile(prefix="golangci-", suffix=".json", delete=False)
    json_path = handle.name
    handle.close()
    cmd = [
        "golangci-lint",
        "run",
        "--config",
        str(config),
        "--timeout",
        "10m",
        "--max-issues-per-linter",
        "0",
        "--max-same-issues",
        "0",
        "--output.json.path",
        json_path,
        "--output.text.path",
        "/dev/null",
        "--issues-exit-code",
        "0",
        "./...",
    ]
    try:
        result = subprocess.run(cmd, cwd=module, text=True, capture_output=True, check=False)
        if result.returncode != 0:
            sys.stderr.write(result.stderr or result.stdout)
            raise SystemExit(result.returncode or 2)
        try:
            payload = json.loads(Path(json_path).read_text(encoding="utf-8") or "{}")
        except (OSError, json.JSONDecodeError):
            sys.stderr.write(result.stdout)
            sys.stderr.write(result.stderr)
            print("golangci-lint did not return findings JSON", file=sys.stderr)
            raise SystemExit(2) from None
    finally:
        Path(json_path).unlink(missing_ok=True)
    if isinstance(payload, dict):
        issues = payload.get("Issues") or []
    elif isinstance(payload, list):
        issues = payload
    else:
        print("golangci-lint did not return a finding list", file=sys.stderr)
        raise SystemExit(2)
    if not isinstance(issues, list):
        print("golangci-lint did not return a finding list", file=sys.stderr)
        raise SystemExit(2)
    return issues


def current_counts(issues: list[dict], module: Path) -> Counter[tuple[str, str]]:
    counts: Counter[tuple[str, str]] = Counter()
    for issue in issues:
        code = finding_code(issue)
        pos = issue.get("Pos") or {}
        filename = rel_path(str(pos.get("Filename") or ""), module)
        if code and filename:
            counts[(filename, code)] += 1
    return counts


def load_baseline(path: Path) -> Counter[tuple[str, str]]:
    if not path.is_file():
        print(f"missing golangci baseline: {path}", file=sys.stderr)
        raise SystemExit(2)
    counts: Counter[tuple[str, str]] = Counter()
    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        filename, code = line.split("\t", 1)
        counts[(filename, code)] += 1
    return counts


def new_findings(
    current: Counter[tuple[str, str]], baseline: Counter[tuple[str, str]]
) -> list[str]:
    failures = []
    for key, count in sorted(current.items()):
        allowed = baseline.get(key, 0)
        if count > allowed:
            failures.append(f"NEW {key[0]} {key[1]} count {allowed} -> {count}")
    return failures


def main() -> int:
    refuse_rewrite(sys.argv[1:])
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=ROOT)
    parser.add_argument("--module", type=Path, default=None)
    parser.add_argument("--config", type=Path, default=None)
    parser.add_argument("--baseline", type=Path, default=None)
    args = parser.parse_args()
    root = args.root.resolve()
    module = (args.module if args.module is not None else DEFAULT_MODULE)
    config = args.config if args.config is not None else DEFAULT_CONFIG
    baseline_path = args.baseline if args.baseline is not None else DEFAULT_BASELINE
    if not module.is_absolute():
        module = (root / module).resolve()
    if not config.is_absolute():
        config = (root / config).resolve()
    if not baseline_path.is_absolute():
        baseline_path = (root / baseline_path).resolve()
    if not config.is_file():
        print(f"missing golangci config: {config}", file=sys.stderr)
        return 2
    baseline = load_baseline(baseline_path)
    before = baseline_path.read_bytes()
    current = current_counts(run_golangci(module, config), module)
    if baseline_path.read_bytes() != before:
        print("golangci baseline was rewritten", file=sys.stderr)
        return 1
    failures = new_findings(current, baseline)
    if failures:
        print("check-static-baseline: FAIL", file=sys.stderr)
        for line in failures:
            print(line, file=sys.stderr)
        return 1
    print(f"check-static-baseline: PASS ({sum(current.values())} finding(s))")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
