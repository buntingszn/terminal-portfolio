/**
 * Shared test setup helpers for terminal-portfolio integration tests.
 *
 * Since the interactive logic lives inline in Astro components, tests
 * recreate the relevant DOM structure and re-implement the event handlers
 * identically to the source components. This validates the behavior of
 * keyboard navigation, command bar, theme toggle, help overlay, and boot
 * sequence in a headless happy-dom environment.
 */

/**
 * Creates the minimal DOM structure used by keyboard nav, command bar,
 * help overlay, and theme toggle components.
 */
export function setupBaseDom() {
  document.documentElement.setAttribute('data-theme', 'dark')

  document.body.innerHTML = `
    <div id="boot-screen" class="boot-screen hidden" aria-hidden="true">
      <div id="boot-messages" class="boot-messages"></div>
    </div>
    <main id="main-content">
      <div id="theme-toggle" type="button" aria-label="Toggle theme">
        <span class="theme-icon">[ light ]</span>
      </div>
    </main>
    <div id="command-bar" class="command-bar hidden" role="dialog" aria-label="Command bar">
      <span class="prompt">:</span>
      <input id="command-input" type="text" autocomplete="off" spellcheck="false" aria-label="Enter command" />
      <span id="command-error" class="error"></span>
    </div>
    <div id="help-overlay" class="help-overlay hidden" role="dialog" aria-label="Keyboard shortcuts" tabindex="-1">
      <div class="help-backdrop"></div>
      <div class="help-content">
        <div class="help-body">
          <div class="help-row"><span class="key">j / k</span><span class="desc">Scroll down / up</span></div>
        </div>
      </div>
    </div>
  `
}

/**
 * Dispatches a keyboard event on a given target.
 */
export function pressKey(
  key: string,
  target: EventTarget = document,
  options: Partial<KeyboardEventInit> = {},
) {
  const event = new KeyboardEvent('keydown', {
    key,
    bubbles: true,
    cancelable: true,
    ...options,
  })
  target.dispatchEvent(event)
  return event
}

/**
 * Clears localStorage and sessionStorage.
 */
export function clearStorage() {
  localStorage.clear()
  sessionStorage.clear()
}
