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
  confirmPassword: '',
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
  if (form.password.length < 6) {
    ElMessage.warning('密码至少 6 位')
    return
  }
  if (form.password !== form.confirmPassword) {
    ElMessage.warning('两次输入的密码不一致')
    return
  }

  loading.value = true
  try {
    const registrationHost =
      typeof window !== 'undefined' ? window.location.hostname : ''
    const body =
      form.mode === 'email'
        ? {
            email: form.email.trim(),
            password: form.password,
            confirm_password: form.confirmPassword,
            registration_host: registrationHost,
          }
        : {
            phone: form.phone.trim(),
            password: form.password,
            confirm_password: form.confirmPassword,
            registration_host: registrationHost,
          }
    const res = await userApi.register(body)
    setSessionUser({
      id: res.user.id,
      email: res.user.email ?? null,
      phone: res.user.phone ?? null,
      nickname: res.user.nickname ?? null,
    })
    if (res.register_coupons_granted && res.register_coupons_granted > 0) {
      ElMessage.success(`注册成功，已发放 ${res.register_coupons_granted} 张优惠券`)
    } else {
      ElMessage.success('注册成功')
    }
    const redirect =
      typeof route.query.redirect === 'string' && route.query.redirect.startsWith('/')
        ? route.query.redirect
        : '/'
    await router.push(redirect)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '注册失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AuthShell title="创建账号" subtitle="一分钟完成注册，即可选购套餐。">
    <div class="auth-mode-pills">
      <button
        type="button"
        class="auth-mode-pill"
        :class="{ active: form.mode === 'email' }"
        @click="form.mode = 'email'"
      >
        邮箱注册
      </button>
      <button
        type="button"
        class="auth-mode-pill"
        :class="{ active: form.mode === 'phone' }"
        @click="form.mode = 'phone'"
      >
        手机注册
      </button>
    </div>

    <form class="auth-form" @submit.prevent="onSubmit">
      <label v-if="form.mode === 'email'" class="auth-field">
        <span class="auth-field-label">邮箱</span>
        <el-input
          v-model="form.email"
          type="email"
          autocomplete="email"
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
          autocomplete="new-password"
          placeholder="至少 6 位"
          show-password
          size="large"
        />
      </label>

      <label class="auth-field">
        <span class="auth-field-label">确认密码</span>
        <el-input
          v-model="form.confirmPassword"
          type="password"
          autocomplete="new-password"
          placeholder="再次输入密码"
          show-password
          size="large"
        />
      </label>

      <button type="submit" class="auth-submit" :disabled="loading">
        {{ loading ? '提交中…' : '免费注册' }}
      </button>
    </form>

    <p class="auth-foot">
      已有账号？
      <RouterLink to="/login">直接登录</RouterLink>
    </p>
  </AuthShell>
</template>
