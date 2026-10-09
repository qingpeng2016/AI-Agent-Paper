package paper

import "strings"

// resolveLiteraturePDFURL 返回可下载的 PDF 地址（无则空字符串）。
func resolveLiteraturePDFURL(sourceCode, externalKey, rawURL, doi string) string {
	ext := strings.TrimSpace(externalKey)
	switch strings.TrimSpace(sourceCode) {
	case sourceArxiv:
		id := strings.TrimPrefix(ext, "arxiv:")
		if id == "" {
			return ""
		}
		return "https://arxiv.org/pdf/" + id + ".pdf"
	case sourceOpenAlex:
		wid := strings.TrimPrefix(ext, "openalex:")
		if wid == "" {
			return ""
		}
		return "https://content.openalex.org/works/" + wid + ".pdf"
	case sourceSemanticScholar:
		u := strings.TrimSpace(rawURL)
		if u != "" && strings.Contains(strings.ToLower(u), ".pdf") {
			return u
		}
		return ""
	default:
		return ""
	}
}

func literatureLocalPDFFileName(externalKey string) string {
	safe := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			return r
		default:
			return '_'
		}
	}, externalKey)
	if safe == "" {
		safe = "unknown"
	}
	return safe + ".pdf"
}
