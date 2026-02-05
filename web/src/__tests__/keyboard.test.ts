import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { setupBaseDom, pressKey, clearStorage } from './setup'

/**
 * Replicate the KeyboardNav.astro handler for testing.
 * This is an exact copy of the inline script logic.
 */
function installKeyboardHandler() {
  function handleKeyboard(e: KeyboardEvent) {
    const target = e.target as HTMLElement
    if (
      target.tagName === 'INPUT' ||
      target.tagName === 'TEXTAREA' ||
      target.isContentEditable
    ) {
      return
    }

    switch (e.key) {
      case 'j':
        window.scrollBy({ top: 100, behavior: 'smooth' })
        break
      case 'k':
        window.scrollBy({ top: -100, behavior: 'smooth' })
        break
      case '1':
        window.location.href = '/'
        break
      case '2':
        window.location.href = '/work'
        break
      case '3':
        window.location.href = '/cv'
        break
      case '4':
        window.location.href = '/links'
        break
      case 't': {
        const current = document.documentElement.getAttribute('data-theme')
        const next = current === 'dark' ? 'light' : 'dark'
        document.documentElement.setAttribute('data-theme', next)
        localStorage.setItem('theme', next)
        const icon = document.querySelector('#theme-toggle .theme-icon')
        if (icon) {
          icon.textContent = next === 'dark' ? '[ light ]' : '[ dark ]'
        }
        break
      }
      case '?':
        document.dispatchEvent(new CustomEvent('toggle-help'))
        break
      case ':':
        e.preventDefault()
        document.dispatchEvent(new CustomEvent('toggle-command'))
        break
    }
  }

  document.addEventListener('keydown', handleKeyboard)
  return () => document.removeEventListener('keydown', handleKeyboard)
}

describe('KeyboardNav', () => {
  let cleanup: () => void

  beforeEach(() => {
    setupBaseDom()
    clearStorage()
    cleanup = installKeyboardHandler()
  })

  afterEach(() => {
    cleanup()
  })

  describe('j/k scrolling', () => {
    it('scrolls down 100px when j is pressed', () => {
      const scrollSpy = vi.spyOn(window, 'scrollBy').mockImplementation(() => {})
      pressKey('j')
      expect(scrollSpy).toHaveBeenCalledWith({ top: 100, behavior: 'smooth' })
      scrollSpy.mockRestore()
    })

    it('scrolls up 100px when k is pressed', () => {
      const scrollSpy = vi.spyOn(window, 'scrollBy').mockImplementation(() => {})
      pressKey('k')
      expect(scrollSpy).toHaveBeenCalledWith({ top: -100, behavior: 'smooth' })
      scrollSpy.mockRestore()
    })
  })

  describe('1-4 page navigation', () => {
    it('navigates to / when 1 is pressed', () => {
      pressKey('1')
      expect(window.location.href).toContain('/')
    })

    it('navigates to /work when 2 is pressed', () => {
      pressKey('2')
      expect(window.location.href).toContain('/work')
    })

    it('navigates to /cv when 3 is pressed', () => {
      pressKey('3')
      expect(window.location.href).toContain('/cv')
    })

    it('navigates to /links when 4 is pressed', () => {
      pressKey('4')
      expect(window.location.href).toContain('/links')
    })
  })

  describe('theme toggle via t key', () => {
    it('toggles from dark to light', () => {
      document.documentElement.setAttribute('data-theme', 'dark')
      pressKey('t')
      expect(document.documentElement.getAttribute('data-theme')).toBe('light')
      expect(localStorage.getItem('theme')).toBe('light')
    })

    it('toggles from light to dark', () => {
      document.documentElement.setAttribute('data-theme', 'light')
      pressKey('t')
      expect(document.documentElement.getAttribute('data-theme')).toBe('dark')
      expect(localStorage.getItem('theme')).toBe('dark')
    })

    it('updates the theme icon text', () => {
      document.documentElement.setAttribute('data-theme', 'dark')
      pressKey('t')
      const icon = document.querySelector('#theme-toggle .theme-icon')
      expect(icon?.textContent).toBe('[ dark ]')
    })
  })

  describe('? dispatches toggle-help event', () => {
    it('fires toggle-help custom event', () => {
      const handler = vi.fn()
      document.addEventListener('toggle-help', handler)
      pressKey('?')
      expect(handler).toHaveBeenCalledOnce()
      document.removeEventListener('toggle-help', handler)
    })
  })

  describe(': dispatches toggle-command event', () => {
    it('fires toggle-command custom event', () => {
      const handler = vi.fn()
      document.addEventListener('toggle-command', handler)
      pressKey(':')
      expect(handler).toHaveBeenCalledOnce()
      document.removeEventListener('toggle-command', handler)
    })
  })

  describe('input field exclusion', () => {
    it('does not scroll when typing j in an input', () => {
      const scrollSpy = vi.spyOn(window, 'scrollBy').mockImplementation(() => {})
      const input = document.getElementById('command-input') as HTMLInputElement
      pressKey('j', input)
      expect(scrollSpy).not.toHaveBeenCalled()
      scrollSpy.mockRestore()
    })

    it('does not navigate when typing 1 in a textarea', () => {
      const textarea = document.createElement('textarea')
      document.body.appendChild(textarea)
      const originalHref = window.location.href
      pressKey('1', textarea)
      // href should not have changed to /
      // (in happy-dom the assignment may still work, so we check it was not triggered)
      expect(window.location.href).toBe(originalHref)
    })

    it('does not toggle theme when typing t in contenteditable', () => {
      const div = document.createElement('div')
      div.contentEditable = 'true'
      document.body.appendChild(div)
      document.documentElement.setAttribute('data-theme', 'dark')
      pressKey('t', div)
      expect(document.documentElement.getAttribute('data-theme')).toBe('dark')
    })
  })
})
