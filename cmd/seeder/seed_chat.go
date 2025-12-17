package main

import (
	"log"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"gorm.io/gorm"
)

func seedChats(db *gorm.DB) {
	log.Println("📝 Seeding chat sessions...")

	// Get a user to bind chats to (e.g., Member)
	var member models.User
	if err := db.Where("email = ?", "john@example.com").First(&member).Error; err != nil {
		log.Println("  ⚠️ User john@example.com not found, skipping chat seeding")
		return
	}

	chats := []struct {
		Title    string
		Messages []models.ChatMessage
	}{
		{
			Title: "Cara menenangkan diri",
			Messages: []models.ChatMessage{
				{Role: models.ChatRoleUser, Content: "Saya merasa cemas hari ini."},
				{Role: models.ChatRoleAI, Content: "Saya mengerti perasaan Anda. Cobalah teknik pernapasan 4-7-8 untuk menenangkan diri sejenak."},
				{Role: models.ChatRoleUser, Content: "Bagaimana caranya?"},
				{Role: models.ChatRoleAI, Content: "Tarik napas selama 4 detik, tahan selama 7 detik, dan hembuskan perlahan selama 8 detik. Ulangi beberapa kali."},
			},
		},
		{
			Title: "Curhat tentang pekerjaan",
			Messages: []models.ChatMessage{
				{Role: models.ChatRoleUser, Content: "Pekerjaan saya sangat menumpuk."},
				{Role: models.ChatRoleAI, Content: "Terdengar melelahkan. Apakah Anda sudah mencoba memprioritaskan tugas-tugas tersebut?"},
			},
		},
	}

	for _, c := range chats {
		var existing models.ChatSession
		if db.Where("title = ? AND user_id = ?", c.Title, member.ID).First(&existing).RowsAffected == 0 {
			session := models.ChatSession{
				UserID: member.ID,
				Title:  c.Title,
			}
			if err := db.Create(&session).Error; err != nil {
				log.Printf("  ❌ Failed to create chat session: %v", err)
				continue
			}
			log.Printf("  ✓ Created chat session: %s", c.Title)

			// Add messages
			for _, m := range c.Messages {
				m.ChatSessionID = session.ID
				m.CreatedAt = time.Now()
				db.Create(&m)
			}
			log.Printf("    - Added %d messages", len(c.Messages))
		}
	}
}
