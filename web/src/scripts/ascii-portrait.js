/**
 * ascii-portrait.js
 *
 * Runtime animation engine for the undulating ASCII portrait.
 *
 * Supports pluggable animation modes via animation-modes.js:
 *   static, undulating, breathe, glitch, reveal, flow
 *
 * Renders to a single <pre> element via textContent updates at 30fps.
 * Pauses when off-screen or tab is hidden.
 */

import { MODES } from './animation-modes.js';

// --- Default animation parameters ---

const DEFAULTS = {
  // Animation mode
  mode: 'undulating',

  // Horizontal wave
  amplitudeX: 1.5,         // max character displacement
  wavelengthXFactor: 1.0,  // × rows = vertical wavelength
  speedX: 0.4,             // radians per second

  // Brightness ripple
  amplitudeB: 0.15,        // brightness offset (0–1 scale)
  wavelengthBFactor: 0.7,  // × rows = vertical wavelength
  speedB: 0.6,             // radians per second

  // Rendering
  targetFps: 30,

  // Global speed multiplier
  speedMultiplier: 1,
};

// --- Animation controller ---

/**
 * Initialize the undulating ASCII portrait animation.
 *
 * @param {HTMLPreElement} preEl - The <pre> element to render into
 * @param {import('../types/portrait').PortraitData} data - Portrait data from JSON
 * @param {Partial<typeof DEFAULTS>} [overrides] - Optional parameter overrides
 * @returns {{ destroy: () => void }} - Cleanup handle
 */
export function initAnimation(preEl, data, overrides) {
  const cfg = { ...DEFAULTS, ...overrides };
  const { brightness, ramp, cols, rows } = data;
  const frameDuration = 1000 / cfg.targetFps;
  const renderFn = MODES[cfg.mode] || MODES.undulating;

  let running = false;
  let rafId = 0;
  let lastFrameTime = 0;
  let pausedAt = 0;
  let timeOffset = 0;

  function animate(timestamp) {
    if (!running) return;
    rafId = requestAnimationFrame(animate);

    // Throttle to target FPS
    if (timestamp - lastFrameTime < frameDuration) return;
    lastFrameTime = timestamp;

    const t = (timestamp - timeOffset) * 0.001 * (cfg.speedMultiplier || 1);
    preEl.textContent = renderFn(brightness, ramp, t, cols, rows, cfg);
  }

  function pause(now) {
    if (!running) return;
    running = false;
    pausedAt = now;
    cancelAnimationFrame(rafId);
  }

  function resume(now) {
    if (running) return;
    timeOffset += now - pausedAt;
    running = true;
    rafId = requestAnimationFrame(animate);
  }

  function start() {
    running = true;
    timeOffset = 0;
    pausedAt = 0;
    lastFrameTime = 0;
    rafId = requestAnimationFrame(animate);
  }

  // --- Visibility management ---

  const observer = new IntersectionObserver(
    ([entry]) => {
      if (entry.isIntersecting) {
        resume(performance.now());
      } else {
        pause(performance.now());
      }
    },
    { threshold: 0.1 }
  );
  observer.observe(preEl);

  function onVisibilityChange() {
    if (document.hidden) {
      pause(performance.now());
    } else {
      resume(performance.now());
    }
  }
  document.addEventListener('visibilitychange', onVisibilityChange);

  // --- Lifecycle ---

  start();

  return {
    destroy() {
      pause(performance.now());
      observer.disconnect();
      document.removeEventListener('visibilitychange', onVisibilityChange);
    },
  };
}
