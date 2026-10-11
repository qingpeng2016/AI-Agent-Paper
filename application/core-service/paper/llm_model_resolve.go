package paper

import "github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"

// modelConfigIDFromTopicPrompts：环节优先，其次 default；0 表示走平台默认 active 模型。
func modelConfigIDFromTopicPrompts(stagePrompt, defPrompt *entity.PaperLLMPromptTemplate) uint {
	if stagePrompt != nil && stagePrompt.Description != nil && stagePrompt.Description.ModelConfigID > 0 {
		return uint(stagePrompt.Description.ModelConfigID)
	}
	if defPrompt != nil && defPrompt.Description != nil && defPrompt.Description.ModelConfigID > 0 {
		return uint(defPrompt.Description.ModelConfigID)
	}
	return 0
}
