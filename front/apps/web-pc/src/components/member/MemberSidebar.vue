<script setup lang="ts">
import { RouterLink } from 'vue-router'
import type { MemberTab } from '@/mocks/member'
import { memberNav } from '@/mocks/member'
import { userAccountLabel } from '@/composables/useSessionUser'
import type { UserProfile } from '@ai-token-mall/shared'

defineProps<{
  active: MemberTab
  user: UserProfile
}>()

const emit = defineEmits<{
  pickPlan: []
}>()
</script>

<template>
  <aside class="member-sidebar">
    <div class="member-user">
      <span class="member-avatar" aria-hidden="true">{{
        userAccountLabel(user).charAt(0).toUpperCase()
      }}</span>
      <div class="member-user-text">
        <strong>{{ user.nickname || '会员' }}</strong>
        <span>{{ userAccountLabel(user) }}</span>
      </div>
    </div>

    <nav class="member-nav" aria-label="会员中心">
      <RouterLink
        v-for="item in memberNav"
        :key="item.id"
        :to="{ path: '/member', query: { tab: item.id } }"
        class="member-nav-item"
        :class="{ active: active === item.id }"
      >
        <span class="member-nav-label">{{ item.label }}</span>
      </RouterLink>
    </nav>

    <button type="button" class="member-buy-link" @click="emit('pickPlan')">
      选购套餐 →
    </button>
  </aside>
</template>

<style scoped>
.member-sidebar {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px;
  background: #fff;
  border: 1px solid #ede9fe;
  border-radius: 20px;
  box-shadow:
    0 4px 24px rgba(124, 58, 237, 0.08),
    0 16px 40px rgba(30, 27, 75, 0.08);
}

@media (min-width: 960px) {
  .member-sidebar {
    position: sticky;
    top: 72px;
    max-height: calc(100vh - 72px - 24px);
    overflow-y: auto;
  }
}

.member-user {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px 14px 18px;
  margin-bottom: 4px;
  background: linear-gradient(135deg, #f5f3ff 0%, #faf5ff 100%);
  border-radius: 14px;
}

.member-avatar {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 800;
  color: #fff;
  background: var(--atm-gradient);
  border-radius: 14px;
  box-shadow: 0 4px 14px rgba(124, 58, 237, 0.35);
}

.member-user-text {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.member-user-text strong {
  font-size: 16px;
  font-weight: 700;
  color: var(--atm-text);
}

.member-user-text span {
  font-size: 12px;
  color: var(--atm-text-muted);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.member-nav {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 4px 0;
}

.member-nav-item {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 44px;
  padding: 10px 14px 10px 11px;
  font-size: 14px;
  font-weight: 500;
  color: var(--atm-text-muted);
  text-decoration: none;
  border-left: 3px solid transparent;
  border-radius: 12px;
  box-sizing: border-box;
  transition:
    background 0.15s,
    color 0.15s,
    border-color 0.15s;
}

.member-nav-item:hover {
  color: var(--atm-primary);
  background: #f5f3ff;
}

.member-nav-item.active {
  font-weight: 600;
  color: var(--atm-primary);
  background: #ede9fe;
  border-left-color: var(--atm-primary);
}

.member-nav-label {
  font-size: 14px;
}

.member-buy-link {
  display: block;
  width: 100%;
  margin-top: 8px;
  padding: 14px;
  font-size: 14px;
  font-weight: 600;
  font-family: inherit;
  color: #fff;
  text-align: center;
  text-decoration: none;
  background: var(--atm-gradient);
  border: none;
  border-radius: 12px;
  box-shadow: 0 4px 16px rgba(124, 58, 237, 0.3);
  cursor: pointer;
}

.member-buy-link:hover {
  filter: brightness(1.05);
}
</style>
