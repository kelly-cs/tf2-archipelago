"""Turn a raw window screenshot into the picture the README shows.

The Windows launcher is a Win32 window, so nothing on a Linux machine can draw
it: the screenshots come from someone running it, and this is what happens to
them afterwards. Drop the raw capture in docs/images/raw/ and run
`make shadows`.

What it does is a drop shadow on transparent margins. The shadow is taken from
the capture's own alpha channel, so a window with rounded corners, which is
what the Windows 11 snipping tool hands back, casts a rounded shadow, and a
plain rectangle from Alt+PrintScreen casts a rectangular one.

Usage:
    shadow.py <input.png> <output.png>
"""

import sys
from pathlib import Path

from PIL import Image, ImageFilter

# The margin has to hold the blur and the offset, or the shadow meets the edge
# of the picture and stops looking like light.
MARGIN = 72
OFFSET_Y = 18
BLUR = 22
OPACITY = 140


def shadowed(source: Image.Image) -> Image.Image:
    """Return the capture on transparent margins, over its own drop shadow."""
    width, height = source.size
    canvas = Image.new("RGBA", (width + MARGIN * 2, height + MARGIN * 2), (0, 0, 0, 0))

    # The shape of the shadow is the shape of whatever is opaque in the capture.
    # The mask is what says where, and the colour is what says how dark: a mask
    # carrying the opacity as well multiplies the two and washes the shadow out.
    shape = source.getchannel("A").point(lambda value: 255 if value > 0 else 0)
    shadow = Image.new("RGBA", canvas.size, (0, 0, 0, 0))
    shadow.paste((0, 0, 0, OPACITY), (MARGIN, MARGIN + OFFSET_Y), shape)
    shadow = shadow.filter(ImageFilter.GaussianBlur(BLUR))

    canvas = Image.alpha_composite(canvas, shadow)
    canvas.paste(source, (MARGIN, MARGIN), source)
    return canvas


def main() -> int:
    if len(sys.argv) != 3:
        print(__doc__.strip(), file=sys.stderr)
        return 2

    source_path, target_path = Path(sys.argv[1]), Path(sys.argv[2])
    with Image.open(source_path) as opened:
        result = shadowed(opened.convert("RGBA"))

    target_path.parent.mkdir(parents=True, exist_ok=True)
    result.save(target_path, "PNG", optimize=True)
    print(f"wrote {target_path} ({result.width}x{result.height})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
