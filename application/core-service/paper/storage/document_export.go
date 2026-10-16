package storage

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jung-kurt/gofpdf/v2"
)

// TopicDocumentExportBasename 文件名：user_{id}_{YYYYMMDD_HHMMSS}
func TopicDocumentExportBasename(userID uint, at time.Time) string {
	if at.IsZero() {
		at = time.Now()
	}
	return fmt.Sprintf("user_%d_%s", userID, at.Format("20060102_150405"))
}

func topicDocumentExportRelPaths(subdir string, userID uint, at time.Time) (pdfRel, docxRel string) {
	base := TopicDocumentExportBasename(userID, at)
	pdfRel = filepath.ToSlash(filepath.Join(subdir, base+".pdf"))
	docxRel = filepath.ToSlash(filepath.Join(subdir, base+".docx"))
	return pdfRel, docxRel
}

func storageSubdirRoot(subdir, envKey string) string {
	var root string
	if v := strings.TrimSpace(os.Getenv(envKey)); v != "" {
		root = v
	} else {
		root = filepath.Join(StorageRoot(), subdir)
	}
	ensureDirAbs(root)
	abs, err := filepath.Abs(root)
	if err == nil && abs != "" {
		return abs
	}
	return root
}

// WriteTopicDocumentExportFiles 在 storage/{subdir}/ 下写入 PDF 与 Word。
func WriteTopicDocumentExportFiles(subdir, envKey string, userID uint, title, body string, at time.Time) (pdfRel, docxRel string, err error) {
	if at.IsZero() {
		at = time.Now()
	}
	pdfRel, docxRel = topicDocumentExportRelPaths(subdir, userID, at)
	root := storageSubdirRoot(subdir, envKey)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", "", err
	}
	pdfAbs := StorageLocalAbsPath(pdfRel)
	docxAbs := StorageLocalAbsPath(docxRel)
	if err := writeExportPDF(pdfAbs, title, body); err != nil {
		return "", "", fmt.Errorf("write pdf: %w", err)
	}
	if err := writeExportDocx(docxAbs, title, body); err != nil {
		_ = os.Remove(pdfAbs)
		return "", "", fmt.Errorf("write docx: %w", err)
	}
	return pdfRel, docxRel, nil
}

func writeExportPDF(absPath, title, body string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 16)
	for _, line := range splitExportLines(title) {
		pdf.MultiCell(0, 8, sanitizePDFText(line), "", "L", false)
	}
	pdf.Ln(4)
	pdf.SetFont("Helvetica", "", 11)
	for _, line := range splitExportLines(body) {
		if strings.TrimSpace(line) == "" {
			pdf.Ln(3)
			continue
		}
		pdf.MultiCell(0, 5.5, sanitizePDFText(line), "", "L", false)
	}
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return pdf.OutputFileAndClose(absPath)
}

func sanitizePDFText(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == utf8.RuneError {
			continue
		}
		if r < 32 && r != '\n' && r != '\t' {
			continue
		}
		if r > 0xFF {
			b.WriteRune('?')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func splitExportLines(text string) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return strings.Split(text, "\n")
}

func writeExportDocx(absPath, title, body string) error {
	var doc bytes.Buffer
	zw := zip.NewWriter(&doc)

	writeZip := func(name, content string) error {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(content))
		return err
	}

	if err := writeZip("[Content_Types].xml", docxContentTypesXML); err != nil {
		return err
	}
	if err := writeZip("_rels/.rels", docxRelsXML); err != nil {
		return err
	}
	if err := writeZip("word/_rels/document.xml.rels", docxDocumentRelsXML); err != nil {
		return err
	}
	documentXML := buildDocxDocumentXML(title, body)
	if err := writeZip("word/document.xml", documentXML); err != nil {
		return err
	}
	if err := zw.Close(); err != nil {
		return err
	}

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(absPath, doc.Bytes(), 0o644)
}

const docxContentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`

const docxRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

const docxDocumentRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"></Relationships>`

func buildDocxDocumentXML(title, body string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	b.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	if t := strings.TrimSpace(title); t != "" {
		b.WriteString(`<w:p><w:pPr><w:pStyle w:val="Title"/></w:pPr><w:r><w:t xml:space="preserve">`)
		b.WriteString(escapeDocxText(t))
		b.WriteString(`</w:t></w:r></w:p>`)
	}
	full := strings.TrimSpace(body)
	if full != "" {
		for _, line := range splitExportLines(full) {
			line = strings.TrimRight(line, " \t")
			if line == "" {
				b.WriteString(`<w:p/>`)
				continue
			}
			b.WriteString(`<w:p><w:r><w:t xml:space="preserve">`)
			b.WriteString(escapeDocxText(line))
			b.WriteString(`</w:t></w:r></w:p>`)
		}
	}
	b.WriteString(`</w:body></w:document>`)
	return b.String()
}

func escapeDocxText(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	return s
}
