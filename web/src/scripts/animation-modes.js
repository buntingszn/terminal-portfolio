/**
 * animation-modes.js
 *
 * Pluggable animation mode functions for the ASCII portrait.
 * Each mode is a pure function: (brightness, ramp, t, cols, rows, cfg) => string
 *
 * Modes:
 *   static     — No animation, raw character map (baseline)
 *   undulating — Horizontal sine wave + brightness ripple (original)
 *   breathe    — Global brightness pulse, no positional shift
 *   glitch     — Deterministic pseudo-random character noise bursts
 *   reveal     — Typewriter left-to-right, then settles into breathe
 *   flow       — Multi-frequency sine displacement field
 */

const TWO_PI = Math.PI * 2;

function clamp(val, min, max) {
  return val < min ? min : val > max ? max : val;
}

// --- Static ---

function renderStatic(brightness, ramp, _t, cols, rows, _cfg) {
  const rampLen = ramp.length;
  const lines = new Array(rows);
  for (let r = 0; r < rows; r++) {
    const chars = new Array(cols);
    for (let c = 0; c < cols; c++) {
      chars[c] = ramp[Math.round(brightness[r][c] * (rampLen - 1))];
    }
    lines[r] = chars.join('');
  }
  return lines.join('\n');
}

// --- Undulating (original) ---

function shiftRow(rowStr, offset, totalCols) {
  if (offset === 0) return rowStr;
  if (offset > 0) {
    const pad = ' '.repeat(Math.min(offset, totalCols));
    return (pad + rowStr).slice(0, totalCols);
  }
  const abs = Math.min(-offset, totalCols);
  return rowStr.slice(abs).padEnd(totalCols, ' ');
}

function renderUndulating(brightness, ramp, t, cols, rows, cfg) {
  const rampLen = ramp.length;
  const lines = new Array(rows);

  for (let r = 0; r < rows; r++) {
    const chars = new Array(cols);
    for (let c = 0; c < cols; c++) {
      const wavelengthB = rows * (cfg.wavelengthBFactor || 0.7);
      const delta = (cfg.amplitudeB || 0.15) * Math.sin(TWO_PI * r / wavelengthB + t * (cfg.speedB || 0.6));
      const adjusted = clamp(brightness[r][c] + delta, 0, 1);
      chars[c] = ramp[Math.round(adjusted * (rampLen - 1))];
    }

    const wavelengthX = rows * (cfg.wavelengthXFactor || 1.0);
    const offset = Math.round(
      (cfg.amplitudeX || 1.5) * Math.sin(TWO_PI * r / wavelengthX + t * (cfg.speedX || 0.4))
    );
    lines[r] = shiftRow(chars.join(''), offset, cols);
  }

  return lines.join('\n');
}

// --- Breathe ---

function renderBreathe(brightness, ramp, t, cols, rows, cfg) {
  const rampLen = ramp.length;
  const speed = cfg.speedB || 0.6;
  const amplitude = Math.max(cfg.amplitudeB || 0.15, 0.25);
  const pulse = amplitude * Math.sin(t * speed * 1.5);
  const lines = new Array(rows);

  for (let r = 0; r < rows; r++) {
    const chars = new Array(cols);
    for (let c = 0; c < cols; c++) {
      const adjusted = clamp(brightness[r][c] + pulse, 0, 1);
      chars[c] = ramp[Math.round(adjusted * (rampLen - 1))];
    }
    lines[r] = chars.join('');
  }
  return lines.join('\n');
}

// --- Glitch ---

function renderGlitch(brightness, ramp, t, cols, rows, _cfg) {
  const rampLen = ramp.length;
  // Burst timing: active ~30% of the time with irregular cadence
  const burstA = Math.sin(t * 3.7) > 0.65;
  const burstB = Math.sin(t * 7.3 + 1.2) > 0.75;
  const burstIntensity = burstA ? 0.35 : burstB ? 0.2 : 0.03;

  // Occasional horizontal row displacement during bursts
  const rowGlitch = burstA || burstB;
  const glitchSeed = Math.floor(t * 10);

  const lines = new Array(rows);
  for (let r = 0; r < rows; r++) {
    const chars = new Array(cols);
    for (let c = 0; c < cols; c++) {
      // Deterministic hash per cell per time step
      const hash = ((r * 7919 + c * 6271 + glitchSeed * 3571) % 997) / 997;
      if (hash < burstIntensity) {
        const base = Math.round(brightness[r][c] * (rampLen - 1));
        const offset = Math.floor(hash * 7) - 3;
        chars[c] = ramp[clamp(base + offset, 0, rampLen - 1)];
      } else {
        chars[c] = ramp[Math.round(brightness[r][c] * (rampLen - 1))];
      }
    }

    let line = chars.join('');
    // Horizontal displacement on some rows during bursts
    if (rowGlitch) {
      const rowHash = ((r * 4217 + glitchSeed * 2903) % 503) / 503;
      if (rowHash < 0.2) {
        const shift = Math.round((rowHash - 0.1) * 30);
        line = shiftRow(line, shift, cols);
      }
    }
    lines[r] = line;
  }
  return lines.join('\n');
}

// --- Glitch Subtle ---
// Rare, tiny brightness perturbations — a slight digital unease

function renderGlitchSubtle(brightness, ramp, t, cols, rows, _cfg) {
  const rampLen = ramp.length;
  const glitchSeed = Math.floor(t * 6);
  // Very occasional micro-bursts
  const burst = Math.sin(t * 5.1) > 0.9;
  const intensity = burst ? 0.08 : 0.015;

  const lines = new Array(rows);
  for (let r = 0; r < rows; r++) {
    const chars = new Array(cols);
    for (let c = 0; c < cols; c++) {
      const hash = ((r * 7919 + c * 6271 + glitchSeed * 3571) % 997) / 997;
      if (hash < intensity) {
        const base = Math.round(brightness[r][c] * (rampLen - 1));
        const offset = hash < intensity * 0.5 ? 1 : -1;
        chars[c] = ramp[clamp(base + offset, 0, rampLen - 1)];
      } else {
        chars[c] = ramp[Math.round(brightness[r][c] * (rampLen - 1))];
      }
    }
    lines[r] = chars.join('');
  }
  return lines.join('\n');
}

// --- Glitch Scan ---
// A horizontal scan line sweeps down the image, corrupting a narrow band

function renderGlitchScan(brightness, ramp, t, cols, rows, _cfg) {
  const rampLen = ramp.length;
  const glitchSeed = Math.floor(t * 8);
  // Scan line position cycles through rows
  const scanSpeed = 0.3;
  const scanPos = (t * scanSpeed * rows) % rows;
  const scanWidth = 3;

  const lines = new Array(rows);
  for (let r = 0; r < rows; r++) {
    const chars = new Array(cols);
    const distFromScan = Math.abs(r - scanPos);
    const inScanBand = distFromScan < scanWidth || distFromScan > rows - scanWidth;

    for (let c = 0; c < cols; c++) {
      if (inScanBand) {
        const hash = ((r * 7919 + c * 6271 + glitchSeed * 3571) % 997) / 997;
        if (hash < 0.4) {
          const base = Math.round(brightness[r][c] * (rampLen - 1));
          const offset = Math.floor(hash * 5) - 2;
          chars[c] = ramp[clamp(base + offset, 0, rampLen - 1)];
        } else {
          chars[c] = ramp[Math.round(brightness[r][c] * (rampLen - 1))];
        }
      } else {
        chars[c] = ramp[Math.round(brightness[r][c] * (rampLen - 1))];
      }
    }

    let line = chars.join('');
    // Shift the scan band rows slightly
    if (inScanBand) {
      const rowHash = ((r * 4217 + glitchSeed * 2903) % 503) / 503;
      if (rowHash < 0.5) {
        const shift = Math.round((rowHash - 0.25) * 6);
        line = shiftRow(line, shift, cols);
      }
    }
    lines[r] = line;
  }
  return lines.join('\n');
}

// --- Reveal ---

function renderReveal(brightness, ramp, t, cols, rows, cfg) {
  const revealDuration = 2.5; // seconds for full reveal
  const totalCells = cols * rows;
  const revealedCount = Math.min(totalCells, Math.floor((t / revealDuration) * totalCells));
  const fullyRevealed = t > revealDuration;

  const rampLen = ramp.length;
  // After reveal, settle into breathe
  const pulse = fullyRevealed
    ? Math.max(cfg.amplitudeB || 0.15, 0.25) * Math.sin((t - revealDuration) * (cfg.speedB || 0.6) * 1.5)
    : 0;

  const lines = new Array(rows);
  let cellIndex = 0;
  for (let r = 0; r < rows; r++) {
    const chars = new Array(cols);
    for (let c = 0; c < cols; c++) {
      if (cellIndex < revealedCount) {
        const adjusted = clamp(brightness[r][c] + pulse, 0, 1);
        chars[c] = ramp[Math.round(adjusted * (rampLen - 1))];
      } else {
        chars[c] = ' ';
      }
      cellIndex++;
    }
    lines[r] = chars.join('');
  }
  return lines.join('\n');
}

// --- Flow ---

function renderFlow(brightness, ramp, t, cols, rows, cfg) {
  const rampLen = ramp.length;
  const amp = Math.max(cfg.amplitudeX || 1.5, 2.5);
  const lines = new Array(rows);

  for (let r = 0; r < rows; r++) {
    const chars = new Array(cols);
    for (let c = 0; c < cols; c++) {
      // Multi-frequency sine displacement (Perlin-like without a full noise lib)
      const noiseX = Math.sin(r * 0.15 + t * 0.5) * Math.cos(c * 0.08 + t * 0.35)
                   + Math.sin(r * 0.07 + c * 0.11 + t * 0.25) * 0.5;
      const noiseY = Math.cos(r * 0.12 + t * 0.4) * Math.sin(c * 0.1 + t * 0.55)
                   + Math.cos(r * 0.09 + c * 0.13 + t * 0.3) * 0.5;

      const srcR = clamp(Math.round(r + noiseY * amp), 0, rows - 1);
      const srcC = clamp(Math.round(c + noiseX * amp), 0, cols - 1);

      chars[c] = ramp[Math.round(brightness[srcR][srcC] * (rampLen - 1))];
    }
    lines[r] = chars.join('');
  }
  return lines.join('\n');
}

// --- Export ---

export const MODES = {
  static: renderStatic,
  undulating: renderUndulating,
  breathe: renderBreathe,
  glitch: renderGlitch,
  'glitch-subtle': renderGlitchSubtle,
  'glitch-scan': renderGlitchScan,
  reveal: renderReveal,
  flow: renderFlow,
};

export const MODE_NAMES = Object.keys(MODES);
