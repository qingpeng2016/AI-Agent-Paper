package topicdiscovery

// topicDiscoveryStages：原生 /idea-discovery（input 挂在 idea_discovery）；最后 confirm 入库。
var topicDiscoveryStages = []string{
	"idea_discovery",
	"topic_confirm",
}

const (
	BehaviorIdeaDiscoveryRun = "topic.discovery.run"
)
