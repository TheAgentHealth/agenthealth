#!/usr/bin/env python3
"""Checks that relative markdown links (and their #anchors) resolve to real
files and headings in this repository.

This is what would have caught the broken relative GOVERNANCE.md/CONTRIBUTING.md
links in .github/PULL_REQUEST_TEMPLATE.md found during the Phase 1 doc review.

Limitations (acceptable for a repo this size): only scans *.md files (not the
.github/ISSUE_TEMPLATE/*.yml bodies), does not de-duplicate repeated headings
the way GitHub's "-1"/"-2" suffixing does, and does not parse links inside
fenced code blocks out (none of our docs currently use real markdown links
inside code fences).
"""
import pathlib
import re
import sys

ROOT = pathlib.Path(__file__).resolve().parents[1]

LINK_RE = re.compile(r"\[[^\]]*\]\(([^)]+)\)")
HEADING_RE = re.compile(r"^(#{1,6})\s+(.*)$", re.MULTILINE)


def slugify(heading: str) -> str:
    text = heading.strip().lower()
    text = re.sub(r"[`*_]", "", text)
    text = re.sub(r"[^\w\s-]", "", text)
    text = text.replace(" ", "-")
    return text


def headings_in(path: pathlib.Path) -> set:
    return {slugify(m.group(2)) for m in HEADING_RE.finditer(path.read_text(encoding="utf-8"))}


def check_file(path: pathlib.Path, errors: list) -> None:
    text = path.read_text(encoding="utf-8")
    own_headings = headings_in(path)
    for match in LINK_RE.finditer(text):
        target = match.group(1).strip()
        if not target or target.startswith(("http://", "https://", "mailto:")):
            continue
        file_part, _, anchor = target.partition("#")
        if not file_part:
            if anchor and anchor not in own_headings:
                errors.append(f"{path}: broken same-file anchor '#{anchor}'")
            continue
        resolved = (path.parent / file_part).resolve()
        if not resolved.exists():
            errors.append(f"{path}: link target does not exist: {file_part}")
            continue
        if anchor and resolved.is_file() and resolved.suffix == ".md":
            if anchor not in headings_in(resolved):
                errors.append(f"{path}: broken anchor '#{anchor}' in {file_part}")


def main() -> int:
    errors: list = []
    for md_file in sorted(ROOT.rglob("*.md")):
        if ".git" in md_file.parts:
            continue
        check_file(md_file, errors)

    if errors:
        print(f"Found {len(errors)} broken markdown link(s):")
        for err in errors:
            print(f"  - {err}")
        return 1

    print("All markdown links OK.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
