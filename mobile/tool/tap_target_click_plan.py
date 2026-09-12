#!/usr/bin/env python3
"""Drives the tap-target demo's clicks. Called by record_tap_target_demo.sh.

Reads the rects the demo app published and clicks a fixed sweep across each
field, first in the BEFORE panel and then the same relative points in AFTER,
so the two panels are compared at identical positions within the field.
"""
import json
import subprocess
import sys
import time

# Fractions across the field's width. The Categories row sweeps the empty
# field; the Tags row is aimed at the two chips, the gap between them, and the
# space past them -- the places the old tree could not hit-test.
CATEGORY_XS = [0.06, 0.25, 0.45, 0.62, 0.80, 0.95]
TAG_XS = [0.14, 0.263, 0.31, 0.63, 0.95]


def xdo(*args):
    subprocess.run(["xdotool", *args], check=True)


def glide(x, y, steps=10, seconds=0.20):
    """Moves the pointer in visible steps -- a teleporting cursor reads as a
    cut in the recording."""
    out = subprocess.run(
        ["xdotool", "getmouselocation", "--shell"],
        check=True, capture_output=True, text=True).stdout
    env = dict(line.split("=", 1) for line in out.strip().splitlines())
    x0, y0 = int(env["X"]), int(env["Y"])
    for i in range(1, steps + 1):
        xdo("mousemove", str(round(x0 + (x - x0) * i / steps)),
            str(round(y0 + (y - y0) * i / steps)))
        time.sleep(seconds / steps)


def sweep(rect, fractions):
    y = round((rect["top"] + rect["bottom"]) / 2)
    width = rect["right"] - rect["left"]
    for fraction in fractions:
        glide(round(rect["left"] + width * fraction), y)
        time.sleep(0.10)
        xdo("click", "1")
        time.sleep(0.34)


def main():
    rects = json.load(open(sys.argv[1]))
    for panel in ("before", "after"):
        sweep(rects[f"{panel}Categories"], CATEGORY_XS)
        time.sleep(0.3)
        sweep(rects[f"{panel}Tags"], TAG_XS)
        # Step away so the cursor is not parked on the panel being judged.
        glide(640, 560)
        time.sleep(1.0)


if __name__ == "__main__":
    main()
