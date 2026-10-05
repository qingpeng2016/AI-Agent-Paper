<script setup lang="ts">
import { ElMessage } from 'element-plus'

function openSupport() {
  ElMessage.info('在线客服 UI 待接入（企微 / 工单 / 第三方客服）')
}
</script>

<template>
  <Teleport to="body">
    <button
      type="button"
      class="floating-cs"
      aria-label="在线客服"
      @click="openSupport"
    >
      <span class="floating-cs-icon" aria-hidden="true">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
          <path d="M4 11v3a2 2 0 002 2h1" />
          <path d="M20 11v3a2 2 0 01-2 2h-1" />
          <path d="M4 11a8 8 0 0116 0v-1a3 3 0 00-3-3h-1l-1.5-2h-5L8 7H7a3 3 0 00-3 3v1z" />
        </svg>
      </span>
      <span class="floating-cs-label">客服</span>
    </button>
  </Teleport>
</template>

<style scoped>
.floating-cs {
  position: fixed;
  right: max(16px, env(safe-area-inset-right, 0px));
  bottom: max(20px, env(safe-area-inset-bottom, 0px));
  z-index: 1200;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 12px 18px 12px 14px;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  cursor: pointer;
  background: var(--atm-gradient);
  border: none;
  border-radius: 999px;
  box-shadow:
    0 4px 20px rgba(124, 58, 237, 0.45),
    0 2px 8px rgba(30, 27, 75, 0.12);
  transition: transform 0.15s, box-shadow 0.15s;
}

.floating-cs:hover {
  transform: translateY(-2px);
  box-shadow:
    0 8px 28px rgba(124, 58, 237, 0.5),
    0 4px 12px rgba(30, 27, 75, 0.14);
}

.floating-cs:active {
  transform: translateY(0);
}

.floating-cs-icon {
  display: flex;
  width: 22px;
  height: 22px;
}

.floating-cs-icon svg {
  width: 100%;
  height: 100%;
}

.floating-cs-label {
  line-height: 1;
  letter-spacing: 0.02em;
}

@media (max-width: 480px) {
  .floating-cs {
    padding: 14px;
    border-radius: 50%;
  }
  .floating-cs-label {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    white-space: nowrap;
    border: 0;
  }
}
</style>
