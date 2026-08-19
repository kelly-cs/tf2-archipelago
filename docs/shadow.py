"""Turn a raw window screenshot into the picture the README shows.

Two steps. The first repairs what the capture got wrong, the second is the
drop shadow on transparent margins that makes a window look like a window.

The repair is for captures taken through Wine, which is how docs/window-shot.sh
photographs a Win32 window on a machine that is not running Windows. The strip
beside the last tab is left unpainted there, and what shows through is whatever
the frame buffer held, which is black. Windows paints the button face colour.
Any run of pure black wider than a glyph gets that colour, so text is left
alone and the bands are not.

The shadow is taken from the capture's own alpha channel, so a window with
rounded corners, which is what the Windows 11 snipping tool hands back, casts a
rounded shadow, and a plain rectangle casts a rectangular one.

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

# What Windows paints where a control does not, and how long a run of black has
# to be before it is one of those rather than a letter. Ten times the width of
# a glyph at the size the launcher draws them.
FACE = (240, 240, 240)
UNPAINTED_RUN_MIN = 120


def repaint_unpainted(source: Image.Image) -> Image.Image:
    """Fill the black bands a Wine capture leaves with the button face colour."""
    pixels = source.load()
    width, height = source.size

    for y in range(height):
        run = 0
        for x in range(width + 1):
            black = x < width and pixels[x, y][:3] == (0, 0, 0)
            if black:
                run += 1
                continue
            if run >= UNPAINTED_RUN_MIN:
                for back in range(x - run, x):
                    pixels[back, y] = (*FACE, 255)
            run = 0

    return source


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
        result = shadowed(repaint_unpainted(opened.convert("RGBA")))

    target_path.parent.mkdir(parents=True, exist_ok=True)
    result.save(target_path, "PNG", optimize=True)
    print(f"wrote {target_path} ({result.width}x{result.height})")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
