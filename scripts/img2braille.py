#!/usr/bin/env python3
"""Convert an image to Braille Unicode halftone art.

Each Braille character (U+2800) represents a 2-wide x 4-tall dot grid.
The script loads an image, enhances it for detail preservation, dithers
to binary, and maps each 2x4 block to the corresponding Braille codepoint.

Preprocessing pipeline:
  1. Convert to grayscale
  2. CLAHE (local contrast) to bring out facial features
  3. Unsharp mask to sharpen edges
  4. Optional gamma correction
  5. Dithering (Atkinson or Floyd-Steinberg)

Usage:
    python3 scripts/img2braille.py <image_path> [options]
"""

import argparse
import math
import sys

from PIL import Image, ImageFilter, ImageOps

# Braille dot positions: each dot in a 2x4 grid maps to a specific bit.
# Column 0 (left):  rows 0-2 -> bits 0,1,2; row 3 -> bit 6
# Column 1 (right): rows 0-2 -> bits 3,4,5; row 3 -> bit 7
BRAILLE_MAP = [
    [0x01, 0x08],  # row 0
    [0x02, 0x10],  # row 1
    [0x04, 0x20],  # row 2
    [0x40, 0x80],  # row 3
]


def clahe(img, clip_limit=2.0, grid_size=8):
    """Contrast Limited Adaptive Histogram Equalization (pure Pillow).

    Divides the image into grid_size x grid_size tiles, equalizes each
    tile's histogram with a clipped redistribution, then bilinearly
    interpolates between tiles for smooth transitions.
    """
    width, height = img.size
    pixels = list(img.getdata())
    src = [pixels[y * width : (y + 1) * width] for y in range(height)]

    tile_w = max(width // grid_size, 1)
    tile_h = max(height // grid_size, 1)

    # Build per-tile LUTs.
    luts = {}
    for ty in range(grid_size):
        for tx in range(grid_size):
            x0 = tx * tile_w
            y0 = ty * tile_h
            x1 = min(x0 + tile_w, width)
            y1 = min(y0 + tile_h, height)

            hist = [0] * 256
            for y in range(y0, y1):
                for x in range(x0, x1):
                    hist[src[y][x]] += 1

            n = (x1 - x0) * (y1 - y0)
            limit = max(int(clip_limit * n / 256), 1)

            excess = 0
            for i in range(256):
                if hist[i] > limit:
                    excess += hist[i] - limit
                    hist[i] = limit

            bonus = excess // 256
            remainder = excess % 256
            for i in range(256):
                hist[i] += bonus
                if i < remainder:
                    hist[i] += 1

            cdf = [0] * 256
            cdf[0] = hist[0]
            for i in range(1, 256):
                cdf[i] = cdf[i - 1] + hist[i]

            total = cdf[255] if cdf[255] > 0 else 1
            lut = [min(255, int(255.0 * cdf[i] / total)) for i in range(256)]
            luts[(tx, ty)] = lut

    # Bilinear interpolation between tiles.
    out = Image.new("L", (width, height))
    out_pixels = out.load()

    for y in range(height):
        fy = (y - tile_h / 2.0) / tile_h
        ty1 = max(0, min(int(math.floor(fy)), grid_size - 1))
        ty2 = min(ty1 + 1, grid_size - 1)
        wy = fy - ty1
        wy = max(0.0, min(1.0, wy))

        for x in range(width):
            fx = (x - tile_w / 2.0) / tile_w
            tx1 = max(0, min(int(math.floor(fx)), grid_size - 1))
            tx2 = min(tx1 + 1, grid_size - 1)
            wx = fx - tx1
            wx = max(0.0, min(1.0, wx))

            val = src[y][x]
            v00 = luts[(tx1, ty1)][val]
            v10 = luts[(tx2, ty1)][val]
            v01 = luts[(tx1, ty2)][val]
            v11 = luts[(tx2, ty2)][val]

            top = v00 * (1 - wx) + v10 * wx
            bot = v01 * (1 - wx) + v11 * wx
            result = top * (1 - wy) + bot * wy
            out_pixels[x, y] = min(255, max(0, int(result)))

    return out


def preprocess(img, contrast=1.5, sharpen=1.5, gamma=1.0):
    """Enhance image for better Braille rendering.

    Steps:
      1. CLAHE for local contrast (brings out eyes, nose, mouth)
      2. Unsharp mask for edge sharpening
      3. Gamma correction to control midtone density
    """
    # CLAHE: adaptive local contrast.
    img = clahe(img, clip_limit=contrast, grid_size=8)

    # Unsharp mask: sharpen edges to preserve facial features at low res.
    if sharpen > 0:
        radius = 1.5
        img = img.filter(
            ImageFilter.UnsharpMask(
                radius=radius, percent=int(sharpen * 100), threshold=2
            )
        )

    # Gamma correction: <1.0 brightens midtones, >1.0 darkens them.
    if gamma != 1.0:
        lut = [min(255, int(255.0 * (i / 255.0) ** gamma)) for i in range(256)]
        img = img.point(lut)

    return img


def atkinson_dither(pixels, width, height):
    """Atkinson dithering: diffuses only 6/8 of error for crisper detail.

    Bill Atkinson's algorithm (used in the original Mac) preserves more
    contrast than Floyd-Steinberg by intentionally losing 1/4 of the error.
    This creates a more "contrasty" look that's great for portraits.
    """
    for y in range(height):
        for x in range(width):
            old = pixels[y][x]
            new = 255.0 if old > 127.5 else 0.0
            pixels[y][x] = new
            err = (old - new) / 8.0

            # Atkinson distributes error to 6 neighbors (each gets err/8):
            #        X  1  1
            #     1  1  1
            #        1
            neighbors = [
                (x + 1, y),
                (x + 2, y),
                (x - 1, y + 1),
                (x, y + 1),
                (x + 1, y + 1),
                (x, y + 2),
            ]
            for nx, ny in neighbors:
                if 0 <= nx < width and 0 <= ny < height:
                    pixels[ny][nx] += err


def floyd_steinberg_dither(pixels, width, height):
    """Classic Floyd-Steinberg error diffusion."""
    for y in range(height):
        for x in range(width):
            old = pixels[y][x]
            new = 255.0 if old > 127.5 else 0.0
            pixels[y][x] = new
            err = old - new
            if x + 1 < width:
                pixels[y][x + 1] += err * 7.0 / 16.0
            if y + 1 < height:
                if x - 1 >= 0:
                    pixels[y + 1][x - 1] += err * 3.0 / 16.0
                pixels[y + 1][x] += err * 5.0 / 16.0
                if x + 1 < width:
                    pixels[y + 1][x + 1] += err * 1.0 / 16.0


def image_to_braille(
    image_path,
    char_width=25,
    invert=False,
    dither="atkinson",
    contrast=1.5,
    sharpen=1.5,
    gamma=1.0,
):
    """Convert image to Braille art string."""
    img = Image.open(image_path).convert("L")

    # Auto-crop to face region: trim uniform borders.
    img = ImageOps.autocontrast(img, cutoff=0.5)

    # Pixel dimensions: each Braille char covers 2 pixel cols x 4 pixel rows.
    px_width = char_width * 2
    px_height = int(img.height * px_width / img.width)
    # Round up to multiple of 4 for clean row mapping.
    px_height = ((px_height + 3) // 4) * 4

    img = img.resize((px_width, px_height), Image.LANCZOS)

    # Preprocessing: enhance for Braille rendering.
    img = preprocess(img, contrast=contrast, sharpen=sharpen, gamma=gamma)

    # Build float pixel grid for dithering.
    pixels = []
    for y in range(px_height):
        row = []
        for x in range(px_width):
            row.append(float(img.getpixel((x, y))))
        pixels.append(row)

    if dither == "atkinson":
        atkinson_dither(pixels, px_width, px_height)
    else:
        floyd_steinberg_dither(pixels, px_width, px_height)

    # Map 2x4 blocks to Braille characters.
    lines = []
    for block_y in range(0, px_height, 4):
        line = []
        for block_x in range(0, px_width, 2):
            codepoint = 0
            for row in range(4):
                for col in range(2):
                    py = block_y + row
                    px = block_x + col
                    if py < px_height and px < px_width:
                        is_dark = pixels[py][px] < 127.5
                        if invert:
                            is_dark = not is_dark
                        if is_dark:
                            codepoint |= BRAILLE_MAP[row][col]
            line.append(chr(0x2800 + codepoint))
        lines.append("".join(line))

    # Strip trailing empty Braille lines (all U+2800).
    while lines and all(c == "\u2800" for c in lines[-1]):
        lines.pop()

    return "\n".join(lines)


def main():
    parser = argparse.ArgumentParser(
        description="Convert image to Braille halftone art"
    )
    parser.add_argument("image", help="Path to input image")
    parser.add_argument(
        "--width",
        type=int,
        default=25,
        help="Output width in characters (default: 25)",
    )
    parser.add_argument(
        "--invert",
        action="store_true",
        help="Invert brightness (dark bg = filled dots)",
    )
    parser.add_argument(
        "--dither",
        choices=["atkinson", "floyd-steinberg"],
        default="atkinson",
        help="Dithering algorithm (default: atkinson)",
    )
    parser.add_argument(
        "--contrast",
        type=float,
        default=1.5,
        help="CLAHE clip limit for local contrast (default: 1.5)",
    )
    parser.add_argument(
        "--sharpen",
        type=float,
        default=1.5,
        help="Unsharp mask strength (default: 1.5)",
    )
    parser.add_argument(
        "--gamma",
        type=float,
        default=1.0,
        help="Gamma correction: <1 brightens, >1 darkens (default: 1.0)",
    )
    args = parser.parse_args()

    try:
        result = image_to_braille(
            args.image,
            char_width=args.width,
            invert=args.invert,
            dither=args.dither,
            contrast=args.contrast,
            sharpen=args.sharpen,
            gamma=args.gamma,
        )
    except FileNotFoundError:
        print(f"Error: file not found: {args.image}", file=sys.stderr)
        sys.exit(1)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)

    print(result)


if __name__ == "__main__":
    main()
