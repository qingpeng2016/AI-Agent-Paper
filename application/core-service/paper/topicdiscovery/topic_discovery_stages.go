package topicdiscovery

// topicDiscoveryStages：平台 retrieve 备语料；中间一步原生 /idea-discovery 全流程；最后 confirm 入库。
var topicDiscoveryStages = []string{
	"retrieve",
	"idea_discovery",
	"topic_confirm",
}

const (
	BehaviorIdeaDiscoveryRun = "topic.discovery.run"
)
