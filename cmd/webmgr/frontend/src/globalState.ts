import { ref } from 'vue'

const USER_KEY = 'metatube_user'

export const currentUser = ref<any>(null)

export function setCurrentUser(user: any) {
  if (user) {
    localStorage.setItem(USER_KEY, JSON.stringify(user))
  } else {
    localStorage.removeItem(USER_KEY)
  }
  currentUser.value = user
}

export function getCurrentUser(): any {
  const data = localStorage.getItem(USER_KEY)
  if (!data) return null
  try {
    return JSON.parse(data)
  } catch {
    return null
  }
}

export function initCurrentUser() {
  currentUser.value = getCurrentUser()
}