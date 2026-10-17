package topicdiscovery

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	paperstep "github.com/qingpeng2016/ai-agent-paper/application/core-service/paper/step"
	"github.com/qingpeng2016/ai-agent-paper/domain/persistent/entity"
)

var runIDSlugSanitize = regexp.MustCompile(`[^a-z0-9-]+`)

// resolveIdeaDiscoveryRunID 原生：优先工作区 run_state 文件名，否则 meta / direction slug + 日期。
func resolveIdeaDiscoveryRunID(paperRoot string, step *entity.PaperOutputTopicStep, input topicRunInput) string {
	if ws := newestIdeaDiscoveryRunStateID(paperRoot); ws != "" {
		return ws
	}
	if step != nil {
		meta := paperstep.LoadStepExtra(step)
		if s := strings.TrimSpace(fmt.Sprint(meta["aris_run_id"])); s != "" && s != "<nil>" {
			return s
		}
	}
	dir := compactDirectionForPrompt(input)
	slug := strings.ToLower(strings.TrimSpace(dir))
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = runIDSlugSanitize.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "idea-discovery"
	}
	if len(slug) > 48 {
		slug = slug[:48]
	}
	date := time.Now().Format("20060102")
	return fmt.Sprintf("%s-%s", slug, date)
}
