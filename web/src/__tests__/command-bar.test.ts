import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { setupBaseDom, pressKey, clearStorage } from './setup'

/**
 * Replicate CommandBar.astro inline script logic for testing.
 */
function installCommandBar() {
  const bar = document.getElementById('command-bar')
  const input = document.getElementById('command-input') as HTMLInputElement
  const errorEl = document.getElementById('command-error')
  let isOpen = false

  function openBar() {
    if (!bar || !input) return
    isOpen = true
    bar.classList.remove('hidden')
    input.value = ''
    if (errorEl) errorEl.textContent = ''
    input.focus()
  }

  function closeBar() {
    if (!bar || !input) return
    isOpen = false
    bar.classList.add('hidden')
    input.blur()
  }

  function executeCommand(cmd: string) {
    const trimmed = cmd.trim().toLowerCase()

    const routes: Record<string, string> = {
      home: '/',
      work: '/work',
      cv: '/cv',
      links: '/links',
    }

    if (routes[trimmed]) {
      closeBar()
      window.location.href = routes[trimmed]
      return
    }

    if (trimmed.startsWith('theme ')) {
      const theme = trimmed.split(' ')[1]
      if (theme === 'light' || theme === 'dark') {
        document.documentElement.setAttribute('data-theme', theme)
        localStorage.setItem('theme', theme)
        const icon = document.querySelector('#theme-toggle .theme-icon')
        if (icon) icon.textContent = theme === 'dark' ? '[ light ]' : '[ dark ]'
        closeBar()
        return
      }
    }

    if (trimmed === 'help') {
      closeBar()
      document.dispatchEvent(new CustomEvent('toggle-help'))
      return
    }

    if (errorEl) {
      errorEl.textContent = `Unknown: ${trimmed}`
    }
  }

  const toggleHandler = () => {
    if (isOpen) closeBar()
    else openBar()
  }
  document.addEventListener('toggle-command', toggleHandler)

  const inputHandler = (e: KeyboardEvent) => {
    if (e.key === 'Escape') {
      e.preventDefault()
      closeBar()
    } else if (e.key === 'Enter') {
      e.preventDefault()
      const val = input.value.trim()
      if (!val) {
        closeBar()
      } else {
        executeCommand(val)
      }
    }
  }
  input?.addEventListener('keydown', inputHandler)

  return {
    openBar,
    closeBar,
    executeCommand,
    cleanup() {
      document.removeEventListener('toggle-command', toggleHandler)
      input?.removeEventListener('keydown', inputHandler)
    },
  }
}

describe('CommandBar', () => {
  let commandBar: ReturnType<typeof installCommandBar>

  beforeEach(() => {
    setupBaseDom()
    clearStorage()
    commandBar = installCommandBar()
  })

  afterEach(() => {
    commandBar.cleanup()
  })

  describe('open/close', () => {
    it('opens when toggle-command event is dispatched', () => {
      document.dispatchEvent(new CustomEvent('toggle-command'))
      const bar = document.getElementById('command-bar')
      expect(bar?.classList.contains('hidden')).toBe(false)
    })

    it('closes when toggle-command fires while open', () => {
      commandBar.openBar()
      document.dispatchEvent(new CustomEvent('toggle-command'))
      const bar = document.getElementById('command-bar')
      expect(bar?.classList.contains('hidden')).toBe(true)
    })

    it('clears input value on open', () => {
      const input = document.getElementById('command-input') as HTMLInputElement
      input.value = 'leftover'
      commandBar.openBar()
      expect(input.value).toBe('')
    })

    it('clears error message on open', () => {
      const errorEl = document.getElementById('command-error')!
      errorEl.textContent = 'some error'
      commandBar.openBar()
      expect(errorEl.textContent).toBe('')
    })
  })

  describe('Escape key closes bar', () => {
    it('closes the command bar when Escape is pressed in input', () => {
      commandBar.openBar()
      const input = document.getElementById('command-input') as HTMLInputElement
      pressKey('Escape', input)
      const bar = document.getElementById('command-bar')
      expect(bar?.classList.contains('hidden')).toBe(true)
    })
  })

  describe('route commands', () => {
    it('navigates to / for :home', () => {
      commandBar.openBar()
      const input = document.getElementById('command-input') as HTMLInputElement
      input.value = 'home'
      pressKey('Enter', input)
      expect(window.location.href).toContain('/')
    })

    it('navigates to /work for :work', () => {
      commandBar.openBar()
      const input = document.getElementById('command-input') as HTMLInputElement
      input.value = 'work'
      pressKey('Enter', input)
      expect(window.location.href).toContain('/work')
    })

    it('navigates to /cv for :cv', () => {
      commandBar.openBar()
      const input = document.getElementById('command-input') as HTMLInputElement
      input.value = 'cv'
      pressKey('Enter', input)
      expect(window.location.href).toContain('/cv')
    })

    it('navigates to /links for :links', () => {
      commandBar.openBar()
      const input = document.getElementById('command-input') as HTMLInputElement
      input.value = 'links'
      pressKey('Enter', input)
      expect(window.location.href).toContain('/links')
    })

    it('is case-insensitive', () => {
      commandBar.openBar()
      const input = document.getElementById('command-input') as HTMLInputElement
      input.value = 'WORK'
      pressKey('Enter', input)
      expect(window.location.href).toContain('/work')
    })
  })

  describe('theme commands', () => {
    it('sets theme to light via :theme light', () => {
      document.documentElement.setAttribute('data-theme', 'dark')
      commandBar.executeCommand('theme light')
      expect(document.documentElement.getAttribute('data-theme')).toBe('light')
      expect(localStorage.getItem('theme')).toBe('light')
    })

    it('sets theme to dark via :theme dark', () => {
      document.documentElement.setAttribute('data-theme', 'light')
      commandBar.executeCommand('theme dark')
      expect(document.documentElement.getAttribute('data-theme')).toBe('dark')
      expect(localStorage.getItem('theme')).toBe('dark')
    })

    it('updates theme icon text', () => {
      commandBar.executeCommand('theme dark')
      const icon = document.querySelector('#theme-toggle .theme-icon')
      expect(icon?.textContent).toBe('[ light ]')
    })

    it('closes bar after theme command', () => {
      commandBar.openBar()
      commandBar.executeCommand('theme light')
      const bar = document.getElementById('command-bar')
      expect(bar?.classList.contains('hidden')).toBe(true)
    })
  })

  describe('help command', () => {
    it('dispatches toggle-help event for :help', () => {
      const handler = vi.fn()
      document.addEventListener('toggle-help', handler)
      commandBar.executeCommand('help')
      expect(handler).toHaveBeenCalledOnce()
      document.removeEventListener('toggle-help', handler)
    })

    it('closes bar after help command', () => {
      commandBar.openBar()
      commandBar.executeCommand('help')
      const bar = document.getElementById('command-bar')
      expect(bar?.classList.contains('hidden')).toBe(true)
    })
  })

  describe('unknown commands', () => {
    it('shows error for unknown command', () => {
      commandBar.executeCommand('foobar')
      const errorEl = document.getElementById('command-error')
      expect(errorEl?.textContent).toBe('Unknown: foobar')
    })

    it('does not close bar on unknown command', () => {
      commandBar.openBar()
      commandBar.executeCommand('foobar')
      const bar = document.getElementById('command-bar')
      expect(bar?.classList.contains('hidden')).toBe(false)
    })
  })

  describe('empty input', () => {
    it('closes bar on Enter with empty input', () => {
      commandBar.openBar()
      const input = document.getElementById('command-input') as HTMLInputElement
      input.value = ''
      pressKey('Enter', input)
      const bar = document.getElementById('command-bar')
      expect(bar?.classList.contains('hidden')).toBe(true)
    })
  })
})
