<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { userApi } from '@/api'
import {
  beginRouteNavigationLoading,
  endRouteNavigationLoading,
} from '@/composables/useRouteNavigationLoading'

const open = defineModel<boolean>('open', { default: false })

const saving = ref(false)
const form = reactive({
  old_password: '',
  new_password: '',
  confirm_password: '',
})

function resetForm() {
  form.old_password = ''
  form.new_password = ''
  form.confirm_password = ''
}

function close() {
  open.value = false
}

async function submit() {
  if (!form.old_password.trim()) {
    ElMessage.warning('请输入原密码')
    return
  }
  if (form.new_password.length < 6) {
    ElMessage.warning('新密码至少 6 位')
    return
  }
  if (form.new_password !== form.confirm_password) {
    ElMessage.warning('两次输入的新密码不一致')
    return
  }
  saving.value = true
  beginRouteNavigationLoading()
  try {
    await userApi.changePassword({
      old_password: form.old_password,
      new_password: form.new_password,
      confirm_password: form.confirm_password,
    })
    ElMessage.success('密码已修改')
    resetForm()
    close()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '修改失败')
  } finally {
    saving.value = false
    endRouteNavigationLoading()
  }
}

watch(open, (v) => {
  if (v) resetForm()
})
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="cpm-backdrop" @click.self="close">
      <div class="cpm-panel" role="dialog" aria-labelledby="change-password-title">
        <header class="cpm-head">
          <h3 id="change-password-title">修改密码</h3>
          <button type="button" class="cpm-close" aria-label="关闭" @click="close">×</button>
        </header>
        <form class="cpm-form" @submit.prevent="submit">
          <label class="cpm-field">
            <span class="cpm-label">原密码</span>
            <input
              v-model="form.old_password"
              class="cpm-input"
              type="password"
              autocomplete="current-password"
            />
          </label>
          <label class="cpm-field">
            <span class="cpm-label">新密码</span>
            <input
              v-model="form.new_password"
              class="cpm-input"
              type="password"
              autocomplete="new-password"
              minlength="6"
            />
          </label>
          <label class="cpm-field">
            <span class="cpm-label">确认新密码</span>
            <input
              v-model="form.confirm_password"
              class="cpm-input"
              type="password"
              autocomplete="new-password"
              minlength="6"
            />
          </label>
          <div class="cpm-actions">
            <button type="button" class="atm-btn-ghost btn-xs" @click="close">取消</button>
            <button type="submit" class="atm-btn-primary btn-xs" :disabled="saving">
              {{ saving ? '提交中…' : '确认修改' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.cpm-backdrop {
  position: fixed;
  inset: 0;
  z-index: 2050;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px 16px;
  background: rgba(15, 10, 40, 0.52);
  backdrop-filter: blur(4px);
}

.cpm-panel {
  width: 100%;
  max-width: min(420px, 96vw);
  padding: 22px 24px 24px;
  background: #fff;
  border-radius: 18px;
  box-shadow: 0 24px 64px rgba(30, 27, 75, 0.22);
}

.cpm-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}

.cpm-head h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 800;
  color: var(--atm-text);
}

.cpm-close {
  width: 36px;
  height: 36px;
  font-size: 22px;
  line-height: 1;
  color: var(--atm-text-muted);
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  cursor: pointer;
}

.cpm-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.cpm-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.cpm-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--atm-text-muted);
}

.cpm-input {
  width: 100%;
  box-sizing: border-box;
  padding: 10px 12px;
  font-size: 13px;
  color: var(--atm-text);
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.cpm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 8px;
}
</style>
