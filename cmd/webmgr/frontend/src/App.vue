<template>
  <div class="app" :key="key">
    <header class="header">
      <div class="header-content">
        <router-link to="/" class="logo">
          <svg class="logo-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M4 4h16a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2z"/>
            <path d="M8 12h8M8 8h8M8 16h4"/>
          </svg>
          <span class="logo-text">MetaTube</span>
        </router-link>

        <nav class="nav">
          <router-link to="/" class="nav-item">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>
              <polyline points="9 22 9 12 15 12 15 22"/>
            </svg>
            <span>首页</span>
          </router-link>
          <router-link to="/favorites" class="nav-item">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"/>
            </svg>
            <span>收藏</span>
          </router-link>
          <router-link to="/history" class="nav-item">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <polyline points="12 6 12 12 16 14"/>
            </svg>
            <span>历史</span>
          </router-link>
          <router-link to="/follows" class="nav-item">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
              <circle cx="9" cy="7" r="4"/>
              <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
              <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
            </svg>
            <span>关注</span>
</router-link>
        </nav>

        <button class="theme-toggle" onclick="window.themeToggle()"></button>

        <div class="user-menu" style="display: none">
          <button class="user-menu-btn" onclick="window.toggleUserDropdown()">
            <span class="user-name"></span>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="6 9 12 15 18 9"/>
            </svg>
          </button>
          <div class="user-dropdown" id="user-dropdown" style="display: none">
            <div class="dropdown-item" id="menu-change-password" onclick="window.openChangePassword()">
              修改密码
            </div>
            <div class="dropdown-item" id="menu-admin" onclick="window.openAdmin()">
              管理员
            </div>
            <div class="dropdown-item" onclick="window.logoutUser()">
              退出登录
            </div>
          </div>
        </div>
      </div>
    </header>

    <main class="main">
      <router-view />
    </main>
  </div>
</template>

<script>
import './theme.js'
import { api } from '@/api'
import { currentUser, setCurrentUser } from '@/globalState'

window.currentUserData = null

window.updateUserMenu = function(user) {
  window.currentUserData = user
  var menu = document.querySelector('.user-menu')
  var nameSpan = document.querySelector('.user-name')
  var adminMenu = document.getElementById('menu-admin')
  if (menu && nameSpan && user) {
    nameSpan.textContent = user.username
    menu.style.display = 'flex'
    if (adminMenu) {
      adminMenu.style.display = user.role === 'admin' ? 'block' : 'none'
    }
  }
}

window.hideUserMenu = function() {
  var menu = document.querySelector('.user-menu')
  if (menu) menu.style.display = 'none'
}

window.toggleUserDropdown = function() {
  var dropdown = document.getElementById('user-dropdown')
  if (dropdown) {
    dropdown.style.display = dropdown.style.display === 'none' ? 'block' : 'none'
  }
}

window.openChangePassword = function() {
  window.toggleUserDropdown()
  window.location.href = '/change-password'
}

window.openAdmin = function() {
  window.toggleUserDropdown()
  window.location.href = '/admin'
}

window.logoutUser = async function() {
  try {
    await api.logout()
  } catch (e) {}
  setCurrentUser(null)
  window.location.href = '/login'
}

window.addEventListener('user-login-success', function() {
  api.getMe().then(function(res) {
    setCurrentUser(res.user)
    window.updateUserMenu(res.user)
  }).catch(function() {
    setCurrentUser(null)
    window.hideUserMenu()
  })
})

window.initApp = function() {
  api.getMe().then(function(res) {
    setCurrentUser(res.user)
    window.updateUserMenu(res.user)
  }).catch(function() {
    setCurrentUser(null)
    window.hideUserMenu()
  })
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', window.initApp)
} else {
  window.initApp()
}
</script>

<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap');

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

/* must stay :root — modals are teleported to <body>, outside .app, and would
   lose these variables if they were scoped to .app */
:root {
  --primary: #0d9488;
  --primary-light: #14b8a6;
  --primary-dark: #0f766e;
  --bg-primary: #ffffff;
  --bg-secondary: #f8fafc;
  --bg-tertiary: #f1f5f9;
  --text-primary: #0f172a;
  --text-secondary: #475569;
  --text-muted: #94a3b8;
  --border: #e2e8f0;
  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.05);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  --radius-sm: 8px;
  --radius-md: 12px;
  --radius-lg: 16px;
  --radius-xl: 24px;
}

/* html.dark also covers modals teleported to <body>, which sit outside .app */
html.dark,
.app.dark {
  --primary: #eab308;
  --primary-light: #facc15;
  --primary-dark: #ca8a04;
  --bg-primary: #0a0a0a;
  --bg-secondary: #111111;
  --bg-tertiary: #1a1a1a;
  --text-primary: #fafafa;
  --text-secondary: #a1a1aa;
  --text-muted: #71717a;
  --border: #27272a;
  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.4);
  --shadow-md: 0 4px 6px -1px rgba(0, 0, 0, 0.5);
  --shadow-lg: 0 10px 15px -3px rgba(0, 0, 0, 0.6);
}

.app.dark body,
.app.dark {
  background: var(--bg-secondary);
  color: var(--text-primary);
}

.app.dark .header {
  background: var(--bg-primary);
  border-color: var(--border);
}

.app.dark .nav-item {
  color: var(--text-secondary);
}

.app.dark .nav-item:hover {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.app.dark .nav-item.router-link-active {
  background: linear-gradient(135deg, var(--primary) 0%, var(--primary-light) 100%);
  color: black;
  box-shadow: 0 2px 8px rgba(234, 179, 8, 0.4);
}

.app.dark .theme-toggle {
  background: var(--bg-tertiary);
  border-color: var(--border);
  color: var(--text-secondary);
}

.app.dark .theme-toggle:hover {
  background: var(--primary);
  border-color: var(--primary);
  color: black;
}

body {
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  background: var(--bg-secondary);
  color: var(--text-primary);
  line-height: 1.6;
}

.app {
  min-height: 100vh;
}

.header {
  background: var(--bg-primary);
  border-bottom: 1px solid var(--border);
  position: sticky;
  top: 0;
  z-index: 100;
  backdrop-filter: blur(10px);
  background: var(--bg-primary);
}

.header-content {
  max-width: 1400px;
  margin: 0 auto;
  padding: 16px 32px;
  display: flex;
  align-items: center;
  gap: 32px;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
  color: var(--text-primary);
}

.logo-icon {
  width: 28px;
  height: 28px;
  color: var(--primary);
}

.logo-text {
  font-size: 22px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--primary) 0%, var(--primary-light) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.nav {
  display: flex;
  gap: 8px;
  margin-left: auto;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  color: var(--text-secondary);
  text-decoration: none;
  border-radius: var(--radius-md);
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.nav-item svg {
  width: 18px;
  height: 18px;
}

.nav-item:hover {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.nav-item.router-link-active {
  background: linear-gradient(135deg, var(--primary) 0%, var(--primary-light) 100%);
  color: white;
  box-shadow: 0 2px 8px rgba(13, 148, 136, 0.3);
}

.nav-item.router-link-active svg {
  stroke: white;
}

.app.dark .nav-item.router-link-active {
  box-shadow: 0 2px 8px rgba(234, 179, 8, 0.3);
}

.theme-toggle {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
}

.theme-toggle:hover {
  background: var(--primary);
  border-color: var(--primary);
  color: white;
}

.app.dark .theme-toggle:hover {
  color: black;
}

.theme-toggle svg {
  width: 20px;
  height: 20px;
}

.user-menu {
  position: relative;
}

.user-menu-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-tertiary);
  color: var(--text-primary);
  cursor: pointer;
}

.user-menu-btn:hover {
  background: var(--bg-primary);
}

.user-menu-btn svg {
  width: 16px;
  height: 16px;
}

.user-name {
  font-size: 14px;
  font-weight: 500;
}

.user-dropdown {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  min-width: 120px;
  background: var(--bg-primary);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-lg);
  z-index: 200;
}

.dropdown-item {
  padding: 10px 16px;
  font-size: 14px;
  color: var(--text-primary);
  cursor: pointer;
}

.dropdown-item:hover {
  background: var(--bg-tertiary);
}

.main {
  max-width: 1400px;
  margin: 0 auto;
  padding: 32px;
  background: var(--bg-secondary);
}
</style>