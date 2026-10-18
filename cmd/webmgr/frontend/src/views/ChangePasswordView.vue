<template>
  <div class="page-container">
    <div class="page-header">
      <button class="back-btn" onclick="window.history.back()">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M19 12H5M12 19l-7-7 7-7"/>
        </svg>
        返回
      </button>
      <h1>修改密码</h1>
    </div>

    <div class="form-card">
      <form @submit.prevent="handleSubmit">
        <div class="form-group">
          <label>当前密码</label>
          <input type="password" v-model="oldPassword" placeholder="请输入当前密码" required />
        </div>
        <div class="form-group">
          <label>新密码</label>
          <input type="password" v-model="newPassword" placeholder="请输入新密码" required />
        </div>
        <div class="form-group">
          <label>确认新密码</label>
          <input type="password" v-model="confirmPassword" placeholder="请再次输入新密码" required />
        </div>
        <div class="error" v-if="error">{{ error }}</div>
        <div class="success" v-if="success">{{ success }}</div>
        <button type="submit" class="submit-btn" :disabled="loading">
          {{ loading ? '提交中...' : '确认修改' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { api } from '@/api'

const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const error = ref('')
const success = ref('')

async function handleSubmit() {
  error.value = ''
  success.value = ''

  if (newPassword.value !== confirmPassword.value) {
    error.value = '两次输入的新密码不一致'
    return
  }

  if (newPassword.value.length < 6) {
    error.value = '新密码长度不能少于6位'
    return
  }

  loading.value = true
  try {
    await api.changePassword(oldPassword.value, newPassword.value)
    await api.logout()
    localStorage.removeItem('metatube_user')
    success.value = '密码修改成功，正在跳转...'
    setTimeout(() => {
      window.location.href = '/login'
    }, 1000)
  } catch (e: any) {
    error.value = e.message || '修改失败'
  } finally {
    loading.value = false
  }
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
  margin-bottom: 32px;
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

.form-card {
  max-width: 500px;
  background: var(--bg-primary);
  border-radius: var(--radius-lg);
  padding: 32px;
  box-shadow: var(--shadow-md);
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 8px;
  color: var(--text-primary);
}

.form-group input {
  width: 100%;
  padding: 10px 14px;
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  background: var(--bg-secondary);
  color: var(--text-primary);
  font-size: 14px;
}

.form-group input:focus {
  outline: none;
  border-color: var(--primary);
}

.error {
  color: #ef4444;
  font-size: 14px;
  margin-bottom: 16px;
}

.success {
  color: #22c55e;
  font-size: 14px;
  margin-bottom: 16px;
}

.submit-btn {
  width: 100%;
  padding: 12px;
  border: none;
  border-radius: var(--radius-md);
  background: var(--primary);
  color: white;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}

.submit-btn:hover {
  background: var(--primary-dark);
}

.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
