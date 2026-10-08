import type { PaperModuleId } from '@paper/types'
import { reloadInviteRebateModuleData } from '@/composables/useInviteRebateModuleData'
import { reloadTopicDiscoveryFormOptions } from '@/composables/useTopicDiscoveryFormOptions'

export type WorkbenchModuleLoadContext = {
  reloadManuscriptsFromStorage: () => void
}

/** 侧栏切换模块时：重新请求该模块右侧内容依赖的接口 */
export async function loadWorkbenchModuleContent(
  moduleId: PaperModuleId,
  ctx: WorkbenchModuleLoadContext,
): Promise<void> {
  switch (moduleId) {
    case 'topic-discovery':
    case 'personal-center':
      await reloadTopicDiscoveryFormOptions()
      return
    case 'invite-rebate':
      await reloadInviteRebateModuleData()
      return
    case 'my-manuscripts':
      ctx.reloadManuscriptsFromStorage()
      return
    default:
      return
  }
}
