import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { setupBaseDom, clearStorage } from './setup'

/**
 * Replicate ThemeToggle.astro inline script logic for testing.
 */
function installThemeToggle() {
  const toggle = document.getElementById('theme-toggle')
  const icon = toggle?.querySelector('.theme-icon')

  function getTheme(): string {
    const stored = localStorage.getItem('theme')
    if (stored) return stored
    return window.matchMedia('(prefers-color-scheme: light)').matches
      ? 'light'
      : 'dark'
  }

  function setTheme(theme: string) {
    document.documentElement.setAttribute('data-theme', theme)
    localStorage.setItem('theme', theme)
    updateIcon(theme)
  }

  function updateIcon(theme: string) {
    if (icon) {
      icon.textContent = theme === 'dark' ? '[ light ]' : '[ dark ]'
    }
  }

  // Initialize
  const currentTheme = getTheme()
  setTheme(currentTheme)

  const clickHandler = () => {
    const current = document.documentElement.getAttribute('data-theme')
    setTheme(current === 'dark' ? 'light' : 'dark')
  }
  toggle?.addEventListener('click', clickHandler)

  return {
    getTheme,
    setTheme,
    cleanup() {
      toggle?.removeEventListener('click', clickHandler)
    },
  }
}

describe('ThemeToggle', () => {
  let themeToggle: ReturnType<typeof installThemeToggle>

  beforeEach(() => {
    setupBaseDom()
    clearStorage()
  })

  afterEach(() => {
    themeToggle?.cleanup()
    vi.restoreAllMocks()
  })

  describe('initialization', () => {
    it('defaults to dark when no stored preference and system prefers dark', () => {
      vi.spyOn(window, 'matchMedia').mockImplementation(
        (query: string) =>
          ({
            matches: query === '(prefers-color-scheme: light)' ? false : false,
            media: query,
            addEventListener: vi.fn(),
            removeEventListener: vi.fn(),
            dispatchEvent: vi.fn(),
            onchange: null,
            addListener: vi.fn(),
            removeListener: vi.fn(),
          }) as MediaQueryList,
      )
      themeToggle = installThemeToggle()
      expect(document.documentElement.getAttribute('data-theme')).toBe('dark')
    })

    it('defaults to light when system prefers light', () => {
      vi.spyOn(window, 'matchMedia').mockImplementation(
        (query: string) =>
          ({
            matches: query === '(prefers-color-scheme: light)',
            media: query,
            addEventListener: vi.fn(),
            removeEventListener: vi.fn(),
            dispatchEvent: vi.fn(),
            onchange: null,
            addListener: vi.fn(),
            removeListener: vi.fn(),
          }) as MediaQueryList,
      )
      themeToggle = installThemeToggle()
      expect(document.documentElement.getAttribute('data-theme')).toBe('light')
    })

    it('uses stored preference from localStorage', () => {
      localStorage.setItem('theme', 'light')
      themeToggle = installThemeToggle()
      expect(document.documentElement.getAttribute('data-theme')).toBe('light')
    })
  })

  describe('localStorage persistence', () => {
    it('stores theme in localStorage on toggle', () => {
      themeToggle = installThemeToggle()
      themeToggle.setTheme('light')
      expect(localStorage.getItem('theme')).toBe('light')
    })

    it('persists across setTheme calls', () => {
      themeToggle = installThemeToggle()
      themeToggle.setTheme('dark')
      expect(localStorage.getItem('theme')).toBe('dark')
      themeToggle.setTheme('light')
      expect(localStorage.getItem('theme')).toBe('light')
    })
  })

  describe('data-theme attribute toggling', () => {
    it('sets data-theme on documentElement', () => {
      themeToggle = installThemeToggle()
      themeToggle.setTheme('light')
      expect(document.documentElement.getAttribute('data-theme')).toBe('light')
    })

    it('toggles on button click from dark to light', () => {
      document.documentElement.setAttribute('data-theme', 'dark')
      localStorage.setItem('theme', 'dark')
      themeToggle = installThemeToggle()

      const toggle = document.getElementById('theme-toggle')!
      toggle.click()
      expect(document.documentElement.getAttribute('data-theme')).toBe('light')
    })

    it('toggles on button click from light to dark', () => {
      localStorage.setItem('theme', 'light')
      themeToggle = installThemeToggle()

      const toggle = document.getElementById('theme-toggle')!
      toggle.click()
      expect(document.documentElement.getAttribute('data-theme')).toBe('dark')
    })
  })

  describe('icon text', () => {
    it('shows "[ dark ]" when theme is light (offers opposite)', () => {
      themeToggle = installThemeToggle()
      themeToggle.setTheme('light')
      const icon = document.querySelector('#theme-toggle .theme-icon')
      expect(icon?.textContent).toBe('[ dark ]')
    })

    it('shows "[ light ]" when theme is dark (offers opposite)', () => {
      themeToggle = installThemeToggle()
      themeToggle.setTheme('dark')
      const icon = document.querySelector('#theme-toggle .theme-icon')
      expect(icon?.textContent).toBe('[ light ]')
    })
  })
})
