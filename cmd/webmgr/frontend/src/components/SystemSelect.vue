<template>
  <div class="system-select" ref="root">
    <button
      type="button"
      class="system-trigger"
      :class="{ open: isOpen }"
      :disabled="disabled"
      @click="toggle"
      @keydown.down.prevent="openAndHighlight(0)"
      @keydown.esc="close"
      aria-haspopup="listbox"
      :aria-expanded="isOpen"
    >
      <span class="system-trigger-label">{{ selectedName }}</span>
      <svg class="system-caret" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M6 9l6 6 6-6"/>
      </svg>
    </button>

    <Transition name="dropdown">
      <ul
        v-if="isOpen"
        class="system-menu"
        role="listbox"
        :style="menuStyle"
      >
        <li
          v-for="(sys, index) in systems"
          :key="sys.id"
          class="system-option"
          role="option"
          :class="{
            selected: sys.id === modelValue,
            disabled: sys.disabled,
            highlighted: index === highlightIndex,
          }"
          :aria-selected="sys.id === modelValue"
          @click="select(sys)"
          @mouseenter="highlightIndex = index"
        >
          <span class="system-option-name">{{ sys.name }}</span>
          <span v-if="sys.disabled" class="system-option-badge">未启用</span>
          <svg v-else-if="sys.id === modelValue" class="system-option-check" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
            <path d="M20 6L9 17l-5-5"/>
          </svg>
        </li>
      </ul>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch, onMounted, onBeforeUnmount } from 'vue'

export interface SystemOption {
  id: string
  name: string
  providers: string[]
  disabled: boolean
  disabled_hint?: string
}

const props = defineProps<{
  modelValue: string
  systems: SystemOption[]
  disabled?: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  select: [value: string]
}>()

const root = ref<HTMLElement | null>(null)
const isOpen = ref(false)
const highlightIndex = ref(0)
// 'down' opens below the trigger; 'up' opens above when there is not enough room.
const direction = ref<'down' | 'up'>('down')

const selectedName = computed(() => {
  const found = props.systems.find(s => s.id === props.modelValue)
  return found ? found.name : '选择搜索源'
})

const selectedIndex = computed(() =>
  props.systems.findIndex(s => s.id === props.modelValue)
)

function toggle() {
  if (props.disabled || props.systems.length === 0) return
  if (isOpen.value) {
    close()
  } else {
    open()
  }
}

function open() {
  if (props.systems.length === 0) return
  highlightIndex.value = selectedIndex.value >= 0 ? selectedIndex.value : 0
  direction.value = computeDirection()
  isOpen.value = true
}

function openAndHighlight(offset: number) {
  if (!isOpen.value) {
    open()
  }
  highlightIndex.value = clamp(
    (selectedIndex.value >= 0 ? selectedIndex.value : 0) + offset,
    0,
    props.systems.length - 1
  )
}

function close() {
  isOpen.value = false
}

function computeDirection(): 'down' | 'up' {
  const el = root.value
  if (!el) return 'down'
  const rect = el.getBoundingClientRect()
  const estimatedMenuHeight = props.systems.length * 40 + 12
  const spaceBelow = window.innerHeight - rect.bottom
  return spaceBelow < estimatedMenuHeight && rect.top > spaceBelow ? 'up' : 'down'
}

const menuStyle = computed(() => {
  if (direction.value === 'up') {
    return { bottom: 'calc(100% + 8px)', top: 'auto' }
  }
  return { top: 'calc(100% + 8px)', bottom: 'auto' }
})

function select(sys: SystemOption) {
  if (sys.disabled) return
  close()
  emit('update:modelValue', sys.id)
  emit('select', sys.id)
}

function onDocumentClick(e: MouseEvent) {
  if (!isOpen.value) return
  if (root.value && !root.value.contains(e.target as Node)) {
    close()
  }
}

function onWindowResize() {
  if (isOpen.value) {
    direction.value = computeDirection()
  }
}

function clamp(v: number, min: number, max: number) {
  return Math.min(Math.max(v, min), max)
}

function moveHighlight(delta: number) {
  if (props.systems.length === 0) return
  let next = highlightIndex.value + delta
  for (let i = 0; i < props.systems.length; i++) {
    if (next < 0) next = props.systems.length - 1
    if (next >= props.systems.length) next = 0
    if (!props.systems[next].disabled) break
    next += delta > 0 ? 1 : -1
    if (next === highlightIndex.value) break
  }
  highlightIndex.value = clamp(next, 0, props.systems.length - 1)
}

function onKeydown(e: KeyboardEvent) {
  if (!isOpen.value) return
  switch (e.key) {
    case 'ArrowDown':
      e.preventDefault()
      moveHighlight(1)
      break
    case 'ArrowUp':
      e.preventDefault()
      moveHighlight(-1)
      break
    case 'Enter':
    case ' ': {
      e.preventDefault()
      const sys = props.systems[highlightIndex.value]
      if (sys && !sys.disabled) select(sys)
      break
    }
    case 'Escape':
      e.preventDefault()
      close()
      break
    case 'Tab':
      close()
      break
  }
}

watch(
  () => props.systems,
  () => {
    if (isOpen.value && highlightIndex.value >= props.systems.length) {
      highlightIndex.value = 0
    }
  }
)

onMounted(() => {
  document.addEventListener('click', onDocumentClick, true)
  document.addEventListener('keydown', onKeydown)
  window.addEventListener('resize', onWindowResize)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', onDocumentClick, true)
  document.removeEventListener('keydown', onKeydown)
  window.removeEventListener('resize', onWindowResize)
})
</script>

<style scoped>
.system-select {
  position: relative;
  flex-shrink: 0;
}

.system-trigger {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 44px;
  padding: 0 12px;
  min-width: 96px;
  background: var(--bg-secondary);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.system-trigger:hover:not(:disabled) {
  background: var(--bg-tertiary);
  border-color: var(--primary);
}

.system-trigger:focus-visible {
  outline: 2px solid var(--primary);
  outline-offset: 2px;
}

.system-trigger:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.system-trigger.open {
  background: var(--bg-tertiary);
  border-color: var(--primary);
}

.system-trigger-label {
  flex: 1;
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  /* reserve room for the absolutely-positioned caret on both sides so the
     label stays optically centered within the button */
  padding: 0 18px;
}

.system-caret {
  position: absolute;
  right: 10px;
  width: 14px;
  height: 14px;
  color: var(--text-muted);
  pointer-events: none;
  transition: transform 0.2s ease;
}

.system-trigger.open .system-caret {
  transform: rotate(180deg);
}

.system-menu {
  position: absolute;
  left: 0;
  z-index: 100;
  min-width: 220px;
  margin: 0;
  padding: 6px;
  list-style: none;
  background: var(--bg-primary);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  max-height: 320px;
  overflow-y: auto;
}

.system-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 12px;
  border-radius: var(--radius-md);
  font-size: 14px;
  color: var(--text-primary);
  text-align: left;
  cursor: pointer;
  transition: background 0.15s ease;
}

.system-option.highlighted:not(.disabled) {
  background: var(--bg-tertiary);
}

.system-option.selected:not(.disabled) {
  color: var(--primary);
  font-weight: 600;
}

.system-option.disabled {
  color: var(--text-muted);
  cursor: not-allowed;
  opacity: 0.65;
}

.system-option-name {
  flex: 1;
  text-align: left;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.system-option-badge {
  flex-shrink: 0;
  padding: 2px 8px;
  background: var(--bg-tertiary);
  border-radius: 999px;
  font-size: 11px;
  font-weight: 400;
  color: var(--text-muted);
}

.system-option-check {
  width: 14px;
  height: 14px;
  flex-shrink: 0;
  color: var(--primary);
}

.dropdown-enter-active,
.dropdown-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.dropdown-enter-from,
.dropdown-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>