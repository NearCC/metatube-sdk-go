<template>
  <div class="page-container">
    <div class="page-header">
      <button class="back-btn" onclick="window.history.back()">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M19 12H5M12 19l-7-7 7-7"/>
        </svg>
        返回
      </button>
      <h1>管理员</h1>
    </div>

    <div class="tabs">
      <button
        class="tab"
        :class="{ active: activeTab === 'users' }"
        @click="activeTab = 'users'"
      >
        用户管理
      </button>
      <button
        class="tab"
        :class="{ active: activeTab === 'logs' }"
        @click="activeTab = 'logs'; loadLogs()"
      >
        系统日志
      </button>
    </div>

    <div v-if="activeTab === 'logs'" class="tab-content">
      <div class="section">
        <div class="logs-header">
          <h2>系统日志</h2>
          <div class="logs-actions">
            <label class="logs-limit">
              显示行数
              <select v-model.number="logLimit" @change="loadLogs()">
                <option :value="100">100</option>
                <option :value="200">200</option>
                <option :value="500">500</option>
                <option :value="1000">1000</option>
              </select>
            </label>
            <button class="action-btn" @click="loadLogs()" :disabled="logsLoading">
              {{ logsLoading ? '加载中...' : '刷新' }}
            </button>
          </div>
        </div>
        <div v-if="logsTruncated" class="logs-hint">
          仅显示日志末尾的 {{ logLimit }} 行
        </div>
        <pre class="logs-content" v-if="logContent">{{ logContent }}</pre>
        <div v-else class="logs-empty">暂无日志</div>
      </div>
    </div>

    <div v-if="activeTab === 'users'" class="tab-content">
      <div class="section">
        <h2>创建用户</h2>
        <form class="create-form" @submit.prevent="handleCreate">
          <input type="text" v-model="newUsername" placeholder="用户名" required />
          <input type="password" v-model="newPassword" placeholder="密码" required />
          <select v-model="newRole">
            <option value="user">普通用户</option>
            <option value="admin">管理员</option>
          </select>
          <button type="submit" class="action-btn primary" :disabled="creating">
            {{ creating ? '创建中...' : '创建' }}
          </button>
        </form>
        <div class="error" v-if="createError">{{ createError }}</div>
        <div class="success" v-if="createSuccess">{{ createSuccess }}</div>
      </div>

      <div class="section">
        <h2>用户列表</h2>
        <div class="user-list">
          <div class="user-row header">
            <span>用户名</span>
            <span>角色</span>
            <span>状态</span>
            <span>创建时间</span>
            <span>操作</span>
          </div>
          <div class="user-row" v-for="user in users" :key="user.id">
            <span>{{ user.username }}</span>
            <span>
              <select
                v-model="user.role"
                @change="updateRole(user)"
                class="role-select"
                :disabled="user.id === currentUserId"
              >
                <option value="admin">管理员</option>
                <option value="user">普通用户</option>
              </select>
            </span>
            <span>
              <label class="switch" :class="{ disabled: user.id === currentUserId }">
                <input
                  type="checkbox"
                  v-model="user.enabled"
                  @change="updateEnabled(user)"
                  :disabled="user.id === currentUserId"
                />
                <span class="slider"></span>
              </label>
              {{ user.enabled ? '启用' : '停用' }}
              <span v-if="user.id === currentUserId" class="current-tag">（当前用户）</span>
            </span>
            <span>{{ formatDate(user.created_at) }}</span>
            <span>
              <button class="small-btn" @click="openResetPassword(user)">重置密码</button>
            </span>
          </div>
        </div>
      </div>
    </div>

    <div class="modal" v-if="showResetModal">
      <div class="modal-content">
        <h3>重置密码 - {{ selectedUser?.username }}</h3>
        <input type="password" v-model="resetPassword" placeholder="新密码" />
        <div class="modal-actions">
          <button class="action-btn" @click="showResetModal = false">取消</button>
          <button class="action-btn primary" @click="handleResetPassword" :disabled="resetting">
            {{ resetting ? '重置中...' : '确认' }}
          </button>
        </div>
        <div class="error" v-if="resetError">{{ resetError }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '@/api'

interface User {
  id: number
  username: string
  role: string
  enabled: boolean
  created_at: string
}

const activeTab = ref('users')
const users = ref<User[]>([])

const logContent = ref('')
const logsLoading = ref(false)
const logsTruncated = ref(false)
const logLimit = ref(200)

async function loadLogs() {
  logsLoading.value = true
  try {
    const res = await api.getLogs(logLimit.value)
    logContent.value = res.data?.content || ''
    logsTruncated.value = !!res.data?.truncated
  } catch (e: any) {
    logContent.value = ''
    logsTruncated.value = false
    console.error('Failed to load logs:', e)
  } finally {
    logsLoading.value = false
  }
}
const creating = ref(false)
const resetting = ref(false)
const createError = ref('')
const createSuccess = ref('')
const resetError = ref('')

const newUsername = ref('')
const newPassword = ref('')
const newRole = ref('user')

const showResetModal = ref(false)
const selectedUser = ref<User | null>(null)
const resetPassword = ref('')

const currentUserId = ref<number | null>(null)

onMounted(() => {
  const stored = localStorage.getItem('metatube_user')
  if (stored) {
    try {
      const u = JSON.parse(stored)
      currentUserId.value = u.id
    } catch (e) {}
  }
  loadUsers()
})

function loadUsers() {
  api.listUsers().then(res => {
    users.value = res.data
  }).catch(e => {
    console.error('Failed to load users:', e)
  })
}

async function handleCreate() {
  createError.value = ''
  createSuccess.value = ''
  creating.value = true
  try {
    await api.createUser(newUsername.value, newPassword.value, newRole.value)
    createSuccess.value = '用户创建成功'
    newUsername.value = ''
    newPassword.value = ''
    newRole.value = 'user'
    loadUsers()
  } catch (e: any) {
    createError.value = e.message || '创建失败'
  } finally {
    creating.value = false
  }
}

async function updateRole(user: User) {
  try {
    await api.setUserRole(user.id, user.role)
  } catch (e: any) {
    alert(e.message || '设置角色失败')
    loadUsers()
  }
}

async function updateEnabled(user: User) {
  try {
    await api.setUserEnabled(user.id, user.enabled)
  } catch (e: any) {
    alert(e.message || '设置状态失败')
    loadUsers()
  }
}

function openResetPassword(user: User) {
  selectedUser.value = user
  resetPassword.value = ''
  resetError.value = ''
  showResetModal.value = true
}

async function handleResetPassword() {
  if (!resetPassword.value) return
  resetting.value = true
  resetError.value = ''
  try {
    await api.resetUserPassword(selectedUser.value!.id, resetPassword.value)
    showResetModal.value = false
    resetPassword.value = ''
  } catch (e: any) {
    resetError.value = e.message || '重置密码失败'
  } finally {
    resetting.value = false
  }
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-'
  const d = new Date(dateStr)
  return d.toLocaleDateString('zh-CN')
}
</script>

<style scoped>
.page-container {
  padding: 24px;
}

.page-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 24px;
}

.page-header h1 {
  font-size: 24px;
  font-weight: 600;
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-primary);
  color: var(--text-primary);
  cursor: pointer;
  font-size: 14px;
}

.back-btn svg {
  width: 18px;
  height: 18px;
}

.tabs {
  display: flex;
  gap: 4px;
  border-bottom: 2px solid var(--border);
  margin-bottom: 24px;
}

.tab {
  padding: 12px 24px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  border-bottom: 2px solid transparent;
  margin-bottom: -2px;
  transition: all 0.2s;
}

.tab:hover {
  color: var(--text-primary);
}

.tab.active {
  color: var(--primary);
  border-bottom-color: var(--primary);
}

.tab-content {
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

.section {
  background: var(--bg-primary);
  border-radius: var(--radius-lg);
  padding: 24px;
  margin-bottom: 24px;
}

.section h2 {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 20px;
}

.create-form {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  align-items: center;
}

.create-form input,
.create-form select {
  padding: 8px 12px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 14px;
}

.user-list {
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.user-row {
  display: grid;
  grid-template-columns: 1fr 1fr 1.5fr 1fr 0.8fr;
  gap: 12px;
  padding: 12px 16px;
  align-items: center;
  border-bottom: 1px solid var(--border);
}

.user-row:last-child {
  border-bottom: none;
}

.user-row.header {
  background: var(--bg-tertiary);
  font-weight: 500;
  font-size: 13px;
}

.role-select {
  padding: 4px 8px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 13px;
}

.role-select:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.switch {
  position: relative;
  display: inline-block;
  width: 36px;
  height: 20px;
  margin-right: 6px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.switch.disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.slider {
  position: absolute;
  cursor: pointer;
  inset: 0;
  background: var(--border);
  border-radius: 20px;
  transition: 0.2s;
}

.slider:before {
  position: absolute;
  content: "";
  height: 14px;
  width: 14px;
  left: 3px;
  bottom: 3px;
  background: white;
  border-radius: 50%;
  transition: 0.2s;
}

input:checked + .slider {
  background: var(--primary);
}

input:checked + .slider:before {
  transform: translateX(16px);
}

.current-tag {
  font-size: 12px;
  color: var(--text-muted);
  margin-left: 4px;
}

.action-btn {
  padding: 8px 16px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-tertiary);
  color: var(--text-primary);
  cursor: pointer;
  font-size: 14px;
}

.action-btn.primary {
  background: var(--primary);
  border-color: var(--primary);
  color: white;
}

.action-btn:hover {
  opacity: 0.9;
}

.action-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.small-btn {
  padding: 4px 8px;
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  background: var(--bg-tertiary);
  color: var(--text-primary);
  cursor: pointer;
  font-size: 12px;
}

.error {
  color: #ef4444;
  font-size: 14px;
  margin-top: 12px;
}

.success {
  color: #22c55e;
  font-size: 14px;
  margin-top: 12px;
}

.modal {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: var(--bg-primary);
  border-radius: var(--radius-lg);
  padding: 24px;
  min-width: 320px;
}

.modal-content h3 {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 16px;
}

.modal-content input {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 14px;
  margin-bottom: 16px;
}

.modal-actions {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.logs-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}

.logs-header h2 {
  margin-bottom: 0;
}

.logs-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logs-limit {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: var(--text-secondary);
}

.logs-limit select {
  padding: 6px 10px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 14px;
}

.logs-hint {
  font-size: 13px;
  color: var(--text-muted);
  margin-bottom: 12px;
}

.logs-content {
  margin: 0;
  padding: 16px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 12px;
  line-height: 1.6;
  color: var(--text-primary);
  max-height: 60vh;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

.logs-empty {
  padding: 40px;
  text-align: center;
  color: var(--text-muted);
  font-size: 14px;
}
</style>