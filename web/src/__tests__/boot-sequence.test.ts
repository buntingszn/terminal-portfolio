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
 */
function installBootSequence(bootMessages: Array<{ type: string; text: string }>) {
  const screen = document.getElementById('boot-screen')
  const container = document.getElementById('boot-messages')

  // Reset visibility (remove hidden class to simulate first visit)
  screen?.classList.remove('hidden')

  if (sessionStorage.getItem('boot-done')) {
    screen?.classList.add('hidden')
    return { cleanup: () => {} }
  }

  let currentIndex = 0
  let timer: ReturnType<typeof setTimeout> | null = null
  const delay = 50 // faster for tests (original is 200)

  function addMessage() {
    if (!container) return
    if (currentIndex >= bootMessages.length) {
      finish()
      return
    }

    const msg = bootMessages[currentIndex]
    const line = document.createElement('div')
    line.className = 'boot-line ' + msg.type
    line.textContent = msg.text
    container.appendChild(line)
    currentIndex++
    timer = setTimeout(addMessage, delay)
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
      expect(container.children[0].textContent).toBe('BIOS v1.0')

      // Advance timer for second message
      vi.advanceTimersByTime(50)
      expect(container.children.length).toBe(2)
      expect(container.children[1].textContent).toBe('Memory check...')

      // Advance timer for third message
      vi.advanceTimersByTime(50)
      expect(container.children.length).toBe(3)
      expect(container.children[2].textContent).toBe('OK')

      cleanup()
    })

    it('assigns correct CSS classes to boot lines', () => {
      const { cleanup } = installBootSequence(sampleBootMessages)
      const container = document.getElementById('boot-messages')!

      expect(container.children[0].className).toBe('boot-line system')

      vi.advanceTimersByTime(50)
      expect(container.children[1].className).toBe('boot-line info')

      vi.advanceTimersByTime(50)
      expect(container.children[2].className).toBe('boot-line success')

      cleanup()
    })

    it('sets sessionStorage boot-done after all messages', () => {
      const { cleanup } = installBootSequence(sampleBootMessages)

      // Advance through all messages + final timer
      vi.advanceTimersByTime(50 * sampleBootMessages.length)
      expect(sessionStorage.getItem('boot-done')).toBe('1')

      cleanup()
    })

    it('hides boot screen after all messages complete', () => {
      const { cleanup } = installBootSequence(sampleBootMessages)
      const screen = document.getElementById('boot-screen')

      vi.advanceTimersByTime(50 * sampleBootMessages.length)
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
