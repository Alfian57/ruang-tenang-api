package application

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/jung-kurt/gofpdf"
)

// ExportJournals exports journals in the specified format
func (s *JournalService) ExportJournals(ctx context.Context, userID uint, req dto.JournalExportRequest) (*dto.JournalExportResponse, error) {
	var startDate *time.Time
	var endDate *time.Time

	if req.StartDate != "" {
		if t, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			startDate = &t
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			t = t.Add(24*time.Hour - time.Second)
			endDate = &t
		}
	}

	journals, _, err := s.journalRepo.FindByUserID(ctx, userID, 1, 1000, req.Tags, nil, startDate, endDate)
	if err != nil {
		return nil, err
	}

	content := ""
	filename := ""

	switch req.Format {
	case "txt":
		content = s.exportToTXT(ctx, journals)
		filename = fmt.Sprintf("journal_export_%s.txt", time.Now().Format("2006-01-02"))
	case "pdf":
		pdfContent, pdfErr := s.exportToPDF(ctx, journals)
		if pdfErr != nil {
			return nil, pdfErr
		}
		content = base64.StdEncoding.EncodeToString(pdfContent)
		filename = fmt.Sprintf("journal_export_%s.pdf", time.Now().Format("2006-01-02"))
	default:
		return nil, fmt.Errorf("unsupported format: %s", req.Format)
	}

	return &dto.JournalExportResponse{Format: req.Format, Content: content, Filename: filename}, nil
}

func (s *JournalService) exportToTXT(_ context.Context, journals []model.Journal) string {
	var builder strings.Builder
	builder.WriteString("=== RUANG TENANG - JOURNAL EXPORT ===\n")
	builder.WriteString(fmt.Sprintf("Exported: %s\n", time.Now().Format("2006-01-02 15:04")))
	builder.WriteString(fmt.Sprintf("Total Entries: %d\n", len(journals)))
	builder.WriteString("=====================================\n\n")

	for _, journal := range journals {
		builder.WriteString(fmt.Sprintf("--- %s ---\n", journal.CreatedAt.Format("Monday, 2 January 2006")))
		if journal.Title != "" {
			builder.WriteString(fmt.Sprintf("Title: %s\n", journal.Title))
		}
		if journal.Mood != nil {
			builder.WriteString(fmt.Sprintf("Mood: %s %s\n", journal.Mood.GetMoodEmoji(), journal.Mood.Mood))
		}
		if len(journal.Tags) > 0 {
			builder.WriteString(fmt.Sprintf("Tags: %s\n", strings.Join(journal.Tags, ", ")))
		}
		builder.WriteString("\n")
		builder.WriteString(s.stripHTML(journal.Content))
		builder.WriteString("\n\n")
	}

	return builder.String()
}

func (s *JournalService) exportToHTML(_ context.Context, journals []model.Journal) string {
	var builder strings.Builder
	builder.WriteString(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>Journal Export - Ruang Tenang</title>
<style>
body { font-family: 'Segoe UI', sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
h1 { color: #E11D48; text-align: center; }
.entry { border: 1px solid #eee; padding: 20px; margin: 20px 0; border-radius: 8px; }
.entry-date { color: #666; font-size: 0.9em; }
.entry-title { font-size: 1.2em; font-weight: bold; margin: 10px 0; }
.entry-mood { display: inline-block; padding: 4px 12px; background: #f0f0f0; border-radius: 20px; margin: 5px 0; }
.entry-tags { color: #E11D48; font-size: 0.9em; }
.entry-content { margin-top: 15px; line-height: 1.6; white-space: pre-wrap; }
</style>
</head>
<body>
<h1>🧘 Ruang Tenang - Journal Export</h1>
<p style="text-align: center; color: #666;">`)
	builder.WriteString(fmt.Sprintf("Exported: %s | Total Entries: %d", time.Now().Format("2 January 2006"), len(journals)))
	builder.WriteString(`</p>`)

	for _, journal := range journals {
		builder.WriteString(`<div class="entry">`)
		builder.WriteString(fmt.Sprintf(`<div class="entry-date">%s</div>`, journal.CreatedAt.Format("Monday, 2 January 2006")))
		if journal.Title != "" {
			builder.WriteString(fmt.Sprintf(`<div class="entry-title">%s</div>`, journal.Title))
		}
		if journal.Mood != nil {
			builder.WriteString(fmt.Sprintf(`<span class="entry-mood">%s %s</span>`, journal.Mood.GetMoodEmoji(), journal.Mood.Mood))
		}
		if len(journal.Tags) > 0 {
			builder.WriteString(fmt.Sprintf(`<div class="entry-tags">Tags: %s</div>`, strings.Join(journal.Tags, ", ")))
		}
		builder.WriteString(fmt.Sprintf(`<div class="entry-content">%s</div>`, journal.Content))
		builder.WriteString(`</div>`)
	}

	builder.WriteString(`</body></html>`)
	return builder.String()
}

func (s *JournalService) exportToPDF(_ context.Context, journals []model.Journal) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Ruang Tenang - Journal Export")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 10, fmt.Sprintf("Exported: %s | Total Entries: %d", time.Now().Format("2 Jan 2006"), len(journals)))
	pdf.Ln(20)

	for _, journal := range journals {
		pdf.SetFont("Arial", "B", 12)
		pdf.SetFillColor(240, 240, 240)
		header := fmt.Sprintf("%s", journal.CreatedAt.Format("Monday, 2 January 2006"))
		pdf.CellFormat(0, 10, header, "0", 1, "L", true, 0, "")
		pdf.Ln(10)

		if journal.Title != "" {
			pdf.SetFont("Arial", "B", 11)
			pdf.Cell(0, 8, journal.Title)
			pdf.Ln(8)
		}

		pdf.SetFont("Arial", "I", 10)
		meta := ""
		if journal.Mood != nil {
			meta += fmt.Sprintf("Mood: %s | ", journal.Mood.Mood)
		}
		if len(journal.Tags) > 0 {
			meta += fmt.Sprintf("Tags: %s", strings.Join(journal.Tags, ", "))
		}
		if meta != "" {
			pdf.Cell(0, 8, meta)
			pdf.Ln(8)
		}

		pdf.SetFont("Arial", "", 11)
		pdf.MultiCell(0, 6, s.stripHTML(journal.Content), "", "L", false)
		pdf.Ln(10)

		pdf.SetDrawColor(200, 200, 200)
		pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
		pdf.Ln(10)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// stripHTML removes HTML tags and converts common block elements to newlines
func (s *JournalService) stripHTML(content string) string {
	content = strings.ReplaceAll(content, "</p>", "\n\n")
	content = strings.ReplaceAll(content, "</div>", "\n")
	content = strings.ReplaceAll(content, "<br>", "\n")
	content = strings.ReplaceAll(content, "<br/>", "\n")
	content = strings.ReplaceAll(content, "<br />", "\n")

	re := regexp.MustCompile(`<[^>]*>`)
	content = re.ReplaceAllString(content, "")
	content = html.UnescapeString(content)

	return strings.TrimSpace(content)
}
