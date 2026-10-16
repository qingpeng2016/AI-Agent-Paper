package topicdiscovery

import "testing"

func TestRetrieveSearchPending(t *testing.T) {
	if !RetrieveSearchPending(nil) {
		t.Fatal("empty result should be pending")
	}
	if !RetrieveSearchPending([]byte(`{"retrieve_search_pending":true}`)) {
		t.Fatal("explicit pending")
	}
	if RetrieveSearchPending([]byte(`{"hit_count":1}`)) {
		t.Fatal("completed search should not be pending")
	}
}
