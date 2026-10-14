<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'

const reportTitle = ref('')
const courseName = ref('')
const reportNote = ref('')

function onCreateDraft() {
  if (!reportTitle.value.trim()) {
    ElMessage.warning('请填写汇报标题')
    return
  }
  ElMessage.info('课程汇报创建功能即将接入')
}
</script>

<template>
  <section class="cr-panel">
    <div class="cr-card">
      <h2 class="cr-title">新建课程汇报</h2>
      <p class="cr-lead">整理课程大纲、汇报幻灯与讲稿（与论文工作流独立）。</p>
      <label class="cr-field">
        <span class="cr-label">汇报标题</span>
        <input v-model="reportTitle" type="text" class="cr-input" placeholder="例：第 8 周文献研读汇报" />
      </label>
      <label class="cr-field">
        <span class="cr-label">课程名称（可选）</span>
        <input v-model="courseName" type="text" class="cr-input" placeholder="例：机器学习导论" />
      </label>
      <label class="cr-field">
        <span class="cr-label">备注（可选）</span>
        <textarea v-model="reportNote" class="cr-textarea" rows="3" placeholder="截止时间、评分 rubric 等" />
      </label>
      <button type="button" class="paper-btn-primary cr-submit" @click="onCreateDraft">创建草稿</button>
    </div>
    <p class="cr-meta">暂无历史汇报；接入 API 后将在此列出。</p>
  </section>
</template>

<style scoped>
.cr-panel {
  max-width: 640px;
}

.cr-card {
  padding: 24px 26px 28px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}

.cr-title {
  margin: 0 0 8px;
  font-size: 18px;
  font-weight: 700;
  color: #1e1b4b;
}

.cr-lead {
  margin: 0 0 20px;
  font-size: 14px;
  line-height: 1.55;
  color: #64748b;
}

.cr-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 16px;
}

.cr-label {
  font-size: 13px;
  font-weight: 600;
  color: #475569;
}

.cr-input,
.cr-textarea {
  box-sizing: border-box;
  width: 100%;
  padding: 10px 12px;
  font-size: 14px;
  color: #1e293b;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.cr-input:focus,
.cr-textarea:focus {
  border-color: #94a3b8;
  outline: none;
  box-shadow: 0 0 0 3px rgba(148, 163, 184, 0.22);
}

.cr-textarea {
  resize: vertical;
  min-height: 72px;
}

.cr-submit {
  margin-top: 4px;
}

.cr-meta {
  margin: 16px 4px 0;
  font-size: 13px;
  color: #94a3b8;
}
</style>
