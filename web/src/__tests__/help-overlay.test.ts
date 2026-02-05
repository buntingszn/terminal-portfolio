import { describe, it, expect, beforeEach, afterEach } from 'vitest'
import { setupBaseDom, pressKey, clearStorage } from './setup'

/**
 * Replicate HelpOverlay.astro inline script logic for testing.
 */
function installHelpOverlay() {
  const overlay = document.getElementById('help-overlay')
  let isOpen = false

  function openHelp() {
    if (!overlay) return
    isOpen = true
    overlay.classList.remove('hidden')
    overlay.focus()
  }

  function closeHelp() {
    if (!overlay) return
    isOpen = false
    overlay.classList.add('hidden')
  }

  const toggleHandler = () => {
    if (isOpen) closeHelp()
    else openHelp()
  }
  document.addEventListener('toggle-help', toggleHandler)

  const keydownHandler = (e: KeyboardEvent) => {
    if (e.key === 'Escape' || e.key === '?') {
      e.preventDefault()
      closeHelp()
    }
  }
  overlay?.addEventListener('keydown', keydownHandler)

  overlay?.setAttribute('tabindex', '-1')

  return {
    openHelp,
    closeHelp,
    get isOpen() {
      return isOpen
    },
    cleanup() {
      document.removeEventListener('toggle-help', toggleHandler)
      overlay?.removeEventListener('keydown', keydownHandler)
    },
  }
}

describe('HelpOverlay', () => {
  let helpOverlay: ReturnType<typeof installHelpOverlay>

  beforeEach(() => {
    setupBaseDom()
    clearStorage()
    helpOverlay = installHelpOverlay()
  })

  afterEach(() => {
    helpOverlay.cleanup()
  })

  describe('toggle-help event', () => {
    it('opens help overlay when toggle-help is dispatched', () => {
      document.dispatchEvent(new CustomEvent('toggle-help'))
      const overlay = document.getElementById('help-overlay')
      expect(overlay?.classList.contains('hidden')).toBe(false)
    })

    it('closes help overlay when toggle-help is dispatched while open', () => {
      helpOverlay.openHelp()
      document.dispatchEvent(new CustomEvent('toggle-help'))
      const overlay = document.getElementById('help-overlay')
      expect(overlay?.classList.contains('hidden')).toBe(true)
    })

    it('toggles open/close with repeated dispatches', () => {
      const overlay = document.getElementById('help-overlay')

      document.dispatchEvent(new CustomEvent('toggle-help'))
      expect(overlay?.classList.contains('hidden')).toBe(false)

      document.dispatchEvent(new CustomEvent('toggle-help'))
      expect(overlay?.classList.contains('hidden')).toBe(true)

      document.dispatchEvent(new CustomEvent('toggle-help'))
      expect(overlay?.classList.contains('hidden')).toBe(false)
    })
  })

  describe('Escape key closes', () => {
    it('closes overlay when Escape is pressed on overlay', () => {
      helpOverlay.openHelp()
      const overlay = document.getElementById('help-overlay')!
      pressKey('Escape', overlay)
      expect(overlay.classList.contains('hidden')).toBe(true)
    })
  })

  describe('? key closes', () => {
    it('closes overlay when ? is pressed on overlay', () => {
      helpOverlay.openHelp()
      const overlay = document.getElementById('help-overlay')!
      pressKey('?', overlay)
      expect(overlay.classList.contains('hidden')).toBe(true)
    })
  })

  describe('starts hidden', () => {
    it('overlay is hidden by default', () => {
      const overlay = document.getElementById('help-overlay')
      expect(overlay?.classList.contains('hidden')).toBe(true)
    })
  })

  describe('tabindex', () => {
    it('has tabindex=-1 for programmatic focus', () => {
      const overlay = document.getElementById('help-overlay')
      expect(overlay?.getAttribute('tabindex')).toBe('-1')
    })
  })

  describe('aria attributes', () => {
    it('has role=dialog', () => {
      const overlay = document.getElementById('help-overlay')
      expect(overlay?.getAttribute('role')).toBe('dialog')
    })

    it('has aria-label', () => {
      const overlay = document.getElementById('help-overlay')
      expect(overlay?.getAttribute('aria-label')).toBe('Keyboard shortcuts')
    })
  })
})
