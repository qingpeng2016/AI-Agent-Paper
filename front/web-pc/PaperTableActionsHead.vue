<script setup lang="ts">
import { ElTooltip } from 'element-plus'

withDefaults(
  defineProps<{
    tip?: string
  }>(),
  {
    tip: '本列可对每条记录执行相应操作',
  },
)
</script>

<template>
  <ElTooltip :content="tip" placement="top" :show-after="280">
    <span class="ul-actions-head" tabindex="0">操作</span>
  </ElTooltip>
</template>
