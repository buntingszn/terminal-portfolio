import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { setupBaseDom, pressKey, clearStorage } from './setup'

const sampleBootMessages = [
  { type: 'system', text: 'BIOS v1.0' },
  { type: 'info', text: 'Memory check...' },
  { type: 'success', text: 'OK' },
]

/**
 * Replicate BootSequence.astro inline script logic for testing.
 * The original uses `is:inline` with `define:vars` to inject bootMessages.
 * Updated to match TUI-aligned behavior: 80ms delay, no fade, blinking cursor.
 */
function installBootSequence(
  bootMessages: Array<{ type: string; text: string }>,
) {
  const screen = document.getElementById('boot-screen')
  const container = document.getElementById('boot-messages')

  // Boot screen starts hidden (matches production HTML).
  // On return visits, it stays hidden — nothing to do.
  if (sessionStorage.getItem('boot-done')) {
    return { cleanup: () => {} }
  }

  // First visit: show the boot screen
  screen?.classList.remove('hidden')

  let currentIndex = 0
  let timer: ReturnType<typeof setTimeout> | null = null
  const delay = 40 // faster for tests (production is 80ms)
  const finalLineDelay = 75 // faster for tests (production is 150ms)
  let cursor: HTMLSpanElement | null = null

  function createCursor() {
    cursor = document.createElement('span')
    cursor.className = 'boot-cursor'
    cursor.textContent = '\u2588'
    cursor.setAttribute('aria-hidden', 'true')
    return cursor
  }

  function addMessage() {
    if (!container) return
    if (currentIndex >= bootMessages.length) {
      // Start blinking cursor after all lines
      if (cursor) cursor.classList.add('blinking')
      // Pause with blinking cursor before dismissal
      timer = setTimeout(finish, 250)
      return
    }

    const msg = bootMessages[currentIndex]
    const line = document.createElement('div')
    line.className = 'boot-line ' + msg.type
    line.textContent = msg.text

    // Move cursor to end of latest line
    if (cursor && cursor.parentNode) {
      cursor.parentNode.removeChild(cursor)
    }
    if (!cursor) cursor = createCursor()
    line.appendChild(cursor)
    cursor.classList.remove('blinking')

    container.appendChild(line)
    currentIndex++

    // Use longer delay before the final line
    const nextDelay =
      currentIndex === bootMessages.length - 1 ? finalLineDelay : delay
    timer = setTimeout(addMessage, nextDelay)
  }

  function finish() {
    sessionStorage.setItem('boot-done', '1')
    document.removeEventListener('keydown', skipBoot)
    document.removeEventListener('click', skipBoot)
    // In tests, we skip the fade-out animation and just hide directly
    if (screen) {
      screen.classList.add('hidden')
    }
  }

  function skipBoot() {
    if (timer) clearTimeout(timer)
    sessionStorage.setItem('boot-done', '1')
    if (screen) {
      screen.classList.add('hidden')
    }
    document.removeEventListener('keydown', skipBoot)
    document.removeEventListener('click', skipBoot)
  }

  document.addEventListener('keydown', skipBoot)
  document.addEventListener('click', skipBoot)

  addMessage()

  return {
    cleanup() {
      if (timer) clearTimeout(timer)
      document.removeEventListener('keydown', skipBoot)
      document.removeEventListener('click', skipBoot)
    },
  }
}

describe('BootSequence', () => {
  beforeEach(() => {
    setupBaseDom()
    clearStorage()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  describe('first visit', () => {
    it('shows boot screen when no sessionStorage flag', () => {
      const { cleanup } = installBootSequence(sampleBootMessages)
      const screen = document.getElementById('boot-screen')
      expect(screen?.classList.contains('hidden')).toBe(false)
      cleanup()
    })

    it('renders boot messages progressively', () => {
      const { cleanup } = installBootSequence(sampleBootMessages)
      const container = document.getElementById('boot-messages')!

      // First message renders immediately
      expect(container.children.length).toBe(1)
      expect(container.children[0].textContent).toContain('BIOS v1.0')

      // Advance timer for second message
      vi.advanceTimersByTime(40)
      expect(container.children.length).toBe(2)
      expect(container.children[1].textContent).toContain('Memory check...')

      // Advance timer for third message (uses finalLineDelay = 75)
      vi.advanceTimersByTime(75)
      expect(container.children.length).toBe(3)
      expect(container.children[2].textContent).toContain('OK')

      cleanup()
    })

    it('assigns correct CSS classes to boot lines', () => {
      const { cleanup } = installBootSequence(sampleBootMessages)
      const container = document.getElementById('boot-messages')!

      expect(container.children[0].className).toContain('boot-line system')

      vi.advanceTimersByTime(40)
      expect(container.children[1].className).toContain('boot-line info')

      vi.advanceTimersByTime(75)
      expect(container.children[2].className).toContain('boot-line success')

      cleanup()
    })

    it('creates a cursor element on the latest line', () => {
      const { cleanup } = installBootSequence(sampleBootMessages)
      const container = document.getElementById('boot-messages')!

      // Cursor should be on the first line
      const cursor = container.querySelector('.boot-cursor')
      expect(cursor).not.toBeNull()
      expect(cursor?.textContent).toBe('\u2588')
      expect(cursor?.parentElement).toBe(container.children[0])

      // After advancing, cursor moves to next line
      vi.advanceTimersByTime(40)
      const cursor2 = container.querySelector('.boot-cursor')
      expect(cursor2?.parentElement).toBe(container.children[1])

      cleanup()
    })

    it('cursor starts blinking after all messages complete', () => {
      const { cleanup } = installBootSequence(sampleBootMessages)
      const container = document.getElementById('boot-messages')!

      // Advance through all messages
      vi.advanceTimersByTime(40 + 75)
      // Now the last addMessage call fires which triggers the "all done" branch
      vi.advanceTimersByTime(40)

      const cursor = container.querySelector('.boot-cursor')
      expect(cursor?.classList.contains('blinking')).toBe(true)

      cleanup()
    })

    it('sets sessionStorage boot-done after all messages and pause', () => {
      const { cleanup } = installBootSequence(sampleBootMessages)

      // Advance through all messages + blinking pause
      vi.advanceTimersByTime(40 + 75 + 40 + 250)
      expect(sessionStorage.getItem('boot-done')).toBe('1')

      cleanup()
    })

    it('hides boot screen after all messages and pause complete', () => {
      const { cleanup } = installBootSequence(sampleBootMessages)
      const screen = document.getElementById('boot-screen')

      vi.advanceTimersByTime(40 + 75 + 40 + 250)
      expect(screen?.classList.contains('hidden')).toBe(true)

      cleanup()
    })
  })

  describe('skip boot', () => {
    it('skips boot on keydown', () => {
      const { cleanup } = installBootSequence(sampleBootMessages)
      const screen = document.getElementById('boot-screen')

      pressKey('Escape')
      expect(screen?.classList.contains('hidden')).toBe(true)
      expect(sessionStorage.getItem('boot-done')).toBe('1')

      cleanup()
    })

    it('skips boot on click', () => {
      const { cleanup } = installBootSequence(sampleBootMessages)
      const screen = document.getElementById('boot-screen')

      document.dispatchEvent(new MouseEvent('click', { bubbles: true }))
      expect(screen?.classList.contains('hidden')).toBe(true)
      expect(sessionStorage.getItem('boot-done')).toBe('1')

      cleanup()
    })

    it('stops adding messages after skip', () => {
      const { cleanup } = installBootSequence(sampleBootMessages)
      const container = document.getElementById('boot-messages')!

      // Skip immediately (only first message rendered)
      pressKey('Escape')
      const countAtSkip = container.children.length

      // Advance timers -- no more messages should be added
      vi.advanceTimersByTime(500)
      expect(container.children.length).toBe(countAtSkip)

      cleanup()
    })

    it('skips on any key press, not just Escape', () => {
      const { cleanup } = installBootSequence(sampleBootMessages)
      const screen = document.getElementById('boot-screen')

      pressKey('a')
      expect(screen?.classList.contains('hidden')).toBe(true)

      cleanup()
    })
  })

  describe('return visit (boot-done already set)', () => {
    it('hides boot screen immediately when sessionStorage has boot-done', () => {
      sessionStorage.setItem('boot-done', '1')
      const { cleanup } = installBootSequence(sampleBootMessages)
      const screen = document.getElementById('boot-screen')
      expect(screen?.classList.contains('hidden')).toBe(true)

      cleanup()
    })

    it('does not render any boot messages on return visit', () => {
      sessionStorage.setItem('boot-done', '1')
      const { cleanup } = installBootSequence(sampleBootMessages)
      const container = document.getElementById('boot-messages')!
      expect(container.children.length).toBe(0)

      cleanup()
    })
  })
})
