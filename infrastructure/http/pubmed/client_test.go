package pubmed

import "testing"

func TestParseEsummaryDocs_uidsAndAuthorObjects(t *testing.T) {
	const sample = `{
		"header":{"type":"esummary","version":"0.3"},
		"result":{
			"uids":["28018183"],
			"28018183":{
				"uid":"28018183",
				"pubdate":"2016",
				"authors":[{"name":"Wu H","authtype":"Author"}],
				"title":"Changed Hub and Corresponding Functional Connectivity.",
				"articleids":[{"idtype":"doi","value":"10.3389/fnana.2016.00120"}]
			}
		}
	}`
	docs, err := parseEsummaryDocs([]byte(sample))
	if err != nil {
		t.Fatal(err)
	}
	doc, ok := docs["28018183"]
	if !ok {
		t.Fatal("missing doc 28018183")
	}
	if doc.Title == "" {
		t.Fatal("empty title")
	}
	names := doc.authorNames()
	if len(names) != 1 || names[0] != "Wu H" {
		t.Fatalf("authors: %v", names)
	}
}
