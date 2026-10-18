;(function () {
  function initTheme() {
    var stored = localStorage.getItem('theme')

    if (stored) {
      window.__isDark = stored === 'dark'
    } else {
      window.__isDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    }
    applyTheme()
  }

  function getThemeIcon() {
    return window.__isDark
      ? '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>'
      : '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>'
  }

  function applyTheme() {
    var app = document.querySelector('.app')

    if (!app) {
      requestAnimationFrame(applyTheme)
      return
    }

    // also flag <html> so teleported nodes (modals under <body>) inherit the theme
    if (window.__isDark) {
      app.classList.add('dark')
      document.documentElement.classList.add('dark')
    } else {
      app.classList.remove('dark')
      document.documentElement.classList.remove('dark')
    }

    var btn = document.querySelector('.theme-toggle')
    if (btn) {
      btn.innerHTML = getThemeIcon()
      btn.setAttribute('title', window.__isDark ? '切换到浅色模式' : '切换到深色模式')
    }
  }

  function toggleTheme() {
    window.__isDark = !window.__isDark
    localStorage.setItem('theme', window.__isDark ? 'dark' : 'light')
    applyTheme()
  }

  window.themeToggle = toggleTheme
  window.themeApply = applyTheme

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initTheme)
  } else {
    initTheme()
  }

  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function (e) {
    if (!localStorage.getItem('theme')) {
      window.__isDark = e.matches
      applyTheme()
    }
  })
})()