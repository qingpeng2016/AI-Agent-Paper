<script setup lang="ts">
import { reactive, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import AuthShell from '@/components/auth/AuthShell.vue'
import { userApi } from '@/api'
import { setSessionUser } from '@/composables/useSessionUser'

const router = useRouter()
const route = useRoute()
const loading = ref(false)
const form = reactive({
  email: '',
  phone: '',
  password: '',
  mode: 'email' as 'email' | 'phone',
})

async function onSubmit() {
  if (form.mode === 'email' && !form.email.trim()) {
    ElMessage.warning('请输入邮箱')
    return
  }
  if (form.mode === 'phone' && !form.phone.trim()) {
    ElMessage.warning('请输入手机号')
    return
  }
  if (!form.password.trim()) {
    ElMessage.warning('请输入密码')
    return
  }

  loading.value = true
  try {
    const body =
      form.mode === 'email'
        ? { email: form.email.trim(), password: form.password }
        : { phone: form.phone.trim(), password: form.password }
    const { user } = await userApi.login(body)
    setSessionUser({
      id: user.id,
      email: user.email ?? null,
      phone: user.phone ?? null,
      nickname: user.nickname ?? null,
    })
    ElMessage.success('登录成功')
    const redirect =
      typeof route.query.redirect === 'string' && route.query.redirect.startsWith('/')
        ? route.query.redirect
        : '/'
    await router.push(redirect)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AuthShell title="欢迎回来">
    <div class="auth-mode-pills">
      <button
        type="button"
        class="auth-mode-pill"
        :class="{ active: form.mode === 'email' }"
        @click="form.mode = 'email'"
      >
        邮箱登录
      </button>
      <button
        type="button"
        class="auth-mode-pill"
        :class="{ active: form.mode === 'phone' }"
        @click="form.mode = 'phone'"
      >
        手机登录
      </button>
    </div>

    <form class="auth-form" @submit.prevent="onSubmit">
      <label v-if="form.mode === 'email'" class="auth-field">
        <span class="auth-field-label">邮箱</span>
        <el-input
          v-model="form.email"
          type="email"
          autocomplete="username"
          placeholder="you@example.com"
          size="large"
        />
      </label>
      <label v-else class="auth-field">
        <span class="auth-field-label">手机号</span>
        <el-input
          v-model="form.phone"
          autocomplete="tel"
          placeholder="11 位手机号"
          size="large"
        />
      </label>

      <label class="auth-field">
        <span class="auth-field-label">密码</span>
        <el-input
          v-model="form.password"
          type="password"
          autocomplete="current-password"
          placeholder="请输入密码"
          show-password
          size="large"
        />
      </label>

      <button type="submit" class="auth-submit" :disabled="loading">
        {{ loading ? '登录中…' : '登录' }}
      </button>
    </form>

    <p class="auth-foot">
      还没有账号？
      <RouterLink to="/register">免费注册</RouterLink>
    </p>
  </AuthShell>
</template>
