package application

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/dto"
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/jung-kurt/gofpdf"
)

func (s *ChatService) ExportChat(ctx context.Context, sessionID, userID uint, req *dto.ExportChatRequest) (*dto.ExportChatResponse, error) {
	session, err := s.sessionRepo.FindByIDWithMessages(ctx, sessionID)
	if err != nil {
		return nil, errors.New("session not found")
	}

	if session.UserID != userID {
		return nil, errors.New("unauthorized")
	}

	messages := session.Messages
	if req.IncludePinned {
		var pinned []model.ChatMessage
		for _, msg := range messages {
			if msg.IsPinned {
				pinned = append(pinned, msg)
			}
		}
		messages = pinned
	}

	switch req.Format {
	case dto.ExportFormatTXT:
		return s.exportAsTXT(ctx, session, messages, req.IncludeMetadata)
	case dto.ExportFormatPDF:
		return s.exportAsPDF(ctx, session, messages, req.IncludeMetadata)
	default:
		return nil, errors.New("unsupported export format")
	}
}

func (s *ChatService) exportAsTXT(ctx context.Context, session *model.ChatSession, messages []model.ChatMessage, includeMetadata bool) (*dto.ExportChatResponse, error) {
	var content strings.Builder

	content.WriteString("═══════════════════════════════════════════\n")
	content.WriteString("  Ruang Tenang - Chat Export\n")
	content.WriteString("═══════════════════════════════════════════\n\n")

	if includeMetadata {
		content.WriteString(fmt.Sprintf("Judul: %s\n", session.Title))
		content.WriteString(fmt.Sprintf("Tanggal Dibuat: %s\n", session.CreatedAt.Format("02 January 2006, 15:04")))
		content.WriteString(fmt.Sprintf("Total Pesan: %d\n", len(messages)))
		if session.Summary != nil && *session.Summary != "" {
			content.WriteString(fmt.Sprintf("\n📝 Ringkasan:\n%s\n", *session.Summary))
		}
		content.WriteString("\n───────────────────────────────────────────\n\n")
	}

	for _, msg := range messages {
		role := "Anda"
		if msg.Role == model.ChatRoleAI {
			role = "AI"
		}

		if includeMetadata {
			content.WriteString(fmt.Sprintf("[%s] %s:\n", msg.CreatedAt.Format("15:04"), role))
		} else {
			content.WriteString(fmt.Sprintf("%s:\n", role))
		}

		content.WriteString(fmt.Sprintf("%s\n", msg.Content))

		if msg.IsPinned {
			content.WriteString("📌 (Pinned)\n")
		}
		content.WriteString("\n")
	}

	content.WriteString("\n───────────────────────────────────────────\n")
	content.WriteString(fmt.Sprintf("Diekspor dari Ruang Tenang pada %s\n", time.Now().Format("02 January 2006, 15:04")))

	filename := fmt.Sprintf("ruang-tenang-chat-%s-%s.txt",
		sanitizeFilename(session.Title),
		time.Now().Format("20060102-150405"))

	return &dto.ExportChatResponse{
		Filename:    filename,
		ContentType: "text/plain; charset=utf-8",
		Content:     content.String(),
		Size:        int64(len(content.String())),
	}, nil
}

func (s *ChatService) exportAsPDF(ctx context.Context, session *model.ChatSession, messages []model.ChatMessage, includeMetadata bool) (*dto.ExportChatResponse, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	pdf.Cell(0, 10, "Ruang Tenang - Chat Export")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 10)

	if includeMetadata {
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(0, 5, fmt.Sprintf("Judul: %s", session.Title))
		pdf.Ln(5)
		pdf.SetFont("Arial", "", 10)
		pdf.Cell(0, 5, fmt.Sprintf("Tanggal: %s", session.CreatedAt.Format("02 Jan 2006, 15:04")))
		pdf.Ln(5)
		pdf.Cell(0, 5, fmt.Sprintf("Total Pesan: %d", len(messages)))
		pdf.Ln(8)

		if session.Summary != nil && *session.Summary != "" {
			pdf.SetFont("Arial", "I", 10)
			pdf.MultiCell(0, 5, fmt.Sprintf("Ringkasan: %s", *session.Summary), "", "", false)
			pdf.Ln(8)
		}

		pdf.Line(10, pdf.GetY(), 200, pdf.GetY())
		pdf.Ln(5)
	}

	for _, msg := range messages {
		pdf.SetTextColor(128, 128, 128)
		pdf.SetFont("Arial", "", 8)
		pdf.Cell(0, 4, msg.CreatedAt.Format("15:04"))
		pdf.Ln(4)

		if msg.Role == model.ChatRoleAI {
			pdf.SetTextColor(0, 0, 128)
			pdf.SetFont("Arial", "B", 10)
			pdf.Cell(0, 5, "AI:")
			pdf.Ln(5)

			pdf.SetTextColor(0, 0, 0)
			pdf.SetFont("Arial", "", 10)
			pdf.MultiCell(0, 5, msg.Content, "", "L", false)
		} else {
			pdf.SetTextColor(0, 100, 0)
			pdf.SetFont("Arial", "B", 10)
			pdf.Cell(0, 5, "Anda:")
			pdf.Ln(5)

			pdf.SetTextColor(0, 0, 0)
			pdf.SetFont("Arial", "", 10)
			pdf.MultiCell(0, 5, msg.Content, "", "L", false)
		}

		if msg.IsPinned {
			pdf.SetTextColor(255, 165, 0)
			pdf.SetFont("Arial", "I", 8)
			pdf.Cell(0, 4, "(Pinned)")
			pdf.Ln(4)
		}

		pdf.Ln(4)
	}

	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(128, 128, 128)
	pdf.Cell(0, 10, fmt.Sprintf("Diekspor dari Ruang Tenang pada %s", time.Now().Format("02/01/2006 15:04")))

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	filename := fmt.Sprintf("ruang-tenang-chat-%s-%s.pdf",
		sanitizeFilename(session.Title),
		time.Now().Format("20060102-150405"))

	return &dto.ExportChatResponse{
		Filename:    filename,
		ContentType: "application/pdf",
		Content:     encoded,
		Size:        int64(buf.Len()),
	}, nil
}

func sanitizeFilename(name string) string {
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "-")
	}
	if len(result) > 50 {
		result = result[:50]
	}
	return strings.ToLower(strings.ReplaceAll(result, " ", "-"))
}
