package paper

import "strings"

// resolveLiteratureHitURL returns a browser-openable link for a literature hit.
func resolveLiteratureHitURL(sourceCode, externalKey, rawURL, doi string) string {
	u := strings.TrimSpace(rawURL)
	if u != "" && strings.HasPrefix(u, "http") {
		return u
	}
	if d := strings.TrimSpace(doi); d != "" {
		if strings.HasPrefix(d, "http") {
			return d
		}
		d = strings.TrimPrefix(strings.TrimPrefix(d, "doi:"), "DOI:")
		return "https://doi.org/" + d
	}
	ext := strings.TrimSpace(externalKey)
	switch strings.TrimSpace(sourceCode) {
	case sourceArxiv:
		id := strings.TrimPrefix(ext, "arxiv:")
		if id == "" {
			id = ext
		}
		if id != "" {
			return "https://arxiv.org/abs/" + id
		}
	case sourceOpenAlex:
		if strings.HasPrefix(ext, "http") {
			return ext
		}
		if strings.HasPrefix(ext, "openalex:") {
			ext = strings.TrimPrefix(ext, "openalex:")
		}
		if ext != "" {
			return "https://openalex.org/" + ext
		}
	case sourceSemanticScholar:
		if strings.HasPrefix(ext, "http") {
			return ext
		}
		id := strings.TrimPrefix(ext, "s2:")
		if id != "" {
			return "https://www.semanticscholar.org/paper/" + id
		}
	}
	return u
}
