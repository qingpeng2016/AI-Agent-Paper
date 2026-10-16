package topicdiscovery

import "testing"

func TestTopicLiteratureSourceRunEnabled(t *testing.T) {
	if !topicLiteratureSourceRunEnabled("active", "arxiv", true) {
		t.Fatal("arxiv active native")
	}
	if topicLiteratureSourceRunEnabled("active", "openalex", true) {
		t.Fatal("openalex not allowed native")
	}
	if topicLiteratureSourceRunEnabled("disabled", "arxiv", true) {
		t.Fatal("disabled arxiv")
	}
	if !topicLiteratureSourceRunEnabled("active", "openalex", false) {
		t.Fatal("openalex allowed legacy retrieve")
	}
}
