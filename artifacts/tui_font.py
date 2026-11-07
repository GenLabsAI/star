"""Locate or download a monospace font for artifact rendering.

Screenshots and walkthrough videos must use the same font so pixels do not
shift between renders. Keep the downloaded file next to the scripts (it is
git-ignored) and let every renderer resolve it through this helper.
"""

import os
import sys
import urllib.request

FONT_NAME = "JetBrainsMono-Regular.ttf"
FONT_URL = (
    "https://raw.githubusercontent.com/JetBrains/JetBrainsMono/master/"
    "fonts/ttf/JetBrainsMono-Regular.ttf"
)


def resolve_font(explicit=None, script_dir=None):
    """Return a usable TTF path, downloading one on first use."""
    if explicit:
        if os.path.isfile(explicit):
            return explicit
        sys.exit(f"font not found: {explicit}")

    root = script_dir or os.path.dirname(os.path.abspath(__file__))
    for name in (FONT_NAME, "DejaVuSansMono.ttf"):
        path = os.path.join(root, name)
        if os.path.isfile(path):
            return path

    path = os.path.join(root, FONT_NAME)
    print(f"downloading {FONT_NAME} ...")
    try:
        urllib.request.urlretrieve(FONT_URL, path)
    except Exception as exc:  # noqa: BLE001 - surface the setup problem.
        if os.path.exists(path):
            os.remove(path)
        sys.exit(
            f"could not download {FONT_NAME}: {exc}\n"
            "Install a monospace TTF next to this script or pass --font."
        )
    return path
