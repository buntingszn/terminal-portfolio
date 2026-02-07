/**
 * Shared test setup helpers for terminal-portfolio integration tests.
 *
 * Creates the minimal DOM structure needed for testing the boot sequence
 * in a headless happy-dom environment.
 */

/**
 * Creates the minimal DOM structure used by boot sequence tests.
 */
export function setupBaseDom() {
  document.documentElement.setAttribute('data-theme', 'dark')

  document.body.innerHTML = `
    <div id="boot-screen" class="boot-screen hidden" aria-hidden="true">
      <div id="boot-messages" class="boot-messages"></div>
    </div>
    <main id="main-content"></main>
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
