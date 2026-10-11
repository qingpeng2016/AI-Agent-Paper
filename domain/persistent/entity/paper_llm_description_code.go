package entity

// paper_llm_description.description_code：首位数字=大类，每类 10000 个码（10000–19999、20000–29999…）
const (
	LLMDescriptionCodeRangeSize = 10000

	LLMDescriptionCodeGateBase  uint64 = 10000 // constraint_gate：1xxxx
	LLMDescriptionCodeVenueBase uint64 = 20000 // constraint_venue：2xxxx
	LLMDescriptionCodeTopicBase uint64 = 30000 // template_topic：3xxxx
)
