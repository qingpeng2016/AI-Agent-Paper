<script setup lang="ts">
import { computed, useId } from 'vue'
import { RouterLink } from 'vue-router'
import { SITE_NAME } from '@/constants/brand'

const props = withDefaults(
  defineProps<{
    variant?: 'header' | 'auth' | 'footer'
    link?: boolean
  }>(),
  {
    variant: 'header',
    link: true,
  },
)

const gradId = useId().replace(/:/g, '')

const iconMode = computed(() => (props.variant === 'auth' ? 'auth' : 'header'))
</script>

<template>
  <component
    :is="link ? RouterLink : 'div'"
    :to="link ? '/' : undefined"
    class="site-logo"
    :class="[`site-logo--${variant}`]"
    :aria-label="link ? SITE_NAME : undefined"
  >
    <span class="site-logo-icon" aria-hidden="true">
      <svg viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
        <defs>
          <linearGradient
            :id="gradId"
            x1="4"
            y1="4"
            x2="28"
            y2="28"
            gradientUnits="userSpaceOnUse"
          >
            <stop stop-color="#7c3aed" />
            <stop offset="1" stop-color="#6366f1" />
          </linearGradient>
        </defs>
        <rect
          width="32"
          height="32"
          rx="9"
          :fill="iconMode === 'auth' ? '#ffffff' : `url(#${gradId})`"
        />
        <rect
          x="8"
          y="19"
          width="16"
          height="2.5"
          rx="1.25"
          :fill="iconMode === 'auth' ? '#6d28d9' : '#ffffff'"
          :opacity="iconMode === 'auth' ? 0.45 : 0.95"
        />
        <rect
          x="8"
          y="14.25"
          width="12"
          height="2.5"
          rx="1.25"
          :fill="iconMode === 'auth' ? '#6d28d9' : '#ffffff'"
          :opacity="iconMode === 'auth' ? 0.75 : 1"
        />
        <rect
          x="8"
          y="9.5"
          width="8"
          height="2.5"
          rx="1.25"
          :fill="iconMode === 'auth' ? '#6d28d9' : '#ffffff'"
        />
      </svg>
    </span>
    <span class="site-logo-word">
      <span class="site-logo-ai">AI</span><span class="site-logo-plan">Plan</span>
    </span>
  </component>
</template>

<style scoped>
.site-logo {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  text-decoration: none;
  flex-shrink: 0;
}

.site-logo-icon {
  display: flex;
  width: 28px;
  height: 28px;
  flex-shrink: 0;
}

.site-logo-icon svg {
  width: 100%;
  height: 100%;
}

.site-logo--header .site-logo-icon {
  filter: drop-shadow(0 2px 8px rgba(124, 58, 237, 0.4));
}

.site-logo-word {
  display: inline-flex;
  align-items: baseline;
  letter-spacing: -0.03em;
  line-height: 1;
}

.site-logo-ai {
  font-size: 1.15em;
  font-weight: 800;
}

.site-logo-plan {
  font-size: 1em;
  font-weight: 700;
}

.site-logo--header .site-logo-word {
  font-size: 17px;
  color: var(--atm-text);
}

.site-logo--header .site-logo-ai {
  background: var(--atm-gradient);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.site-logo--auth .site-logo-icon {
  width: 32px;
  height: 32px;
  filter: drop-shadow(0 4px 12px rgba(0, 0, 0, 0.2));
}

.site-logo--auth .site-logo-word {
  font-size: 20px;
  color: #fff;
}

.site-logo--auth .site-logo-ai {
  color: #fff;
  background: none;
  -webkit-background-clip: unset;
}

.site-logo--auth .site-logo-plan {
  font-weight: 600;
  color: rgba(255, 255, 255, 0.92);
}

.site-logo--footer {
  gap: 0;
}

.site-logo--footer .site-logo-icon {
  display: none;
}

.site-logo--footer .site-logo-word {
  font-size: 16px;
  font-weight: 700;
  color: #fff;
}

.site-logo--footer .site-logo-ai {
  background: linear-gradient(135deg, #c4b5fd 0%, #a78bfa 100%);
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.site-logo--footer .site-logo-plan {
  color: #e2e8f0;
}
</style>
