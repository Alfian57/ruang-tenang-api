package development

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"gorm.io/gorm"
)

// SeedChatSessions seeds test chat sessions for development
func SeedChatSessions(db *gorm.DB) error {
	// Get a test user
	var user models.User
	if err := db.Where("email = ?", "john@example.com").First(&user).Error; err != nil {
		return nil // Skip if user doesn't exist
	}

	chatData := []struct {
		Title    string
		Messages []struct {
			Role    models.ChatRole
			Content string
		}
	}{
		{
			Title: "Cara menenangkan diri",
			Messages: []struct {
				Role    models.ChatRole
				Content string
			}{
				{Role: models.ChatRoleUser, Content: "Saya merasa cemas hari ini."},
				{Role: models.ChatRoleAI, Content: "Saya mengerti perasaan Anda. Rasa cemas adalah hal yang wajar dialami. Cobalah teknik pernapasan 4-7-8 untuk menenangkan diri sejenak. Tarik napas selama 4 detik, tahan 7 detik, lalu hembuskan selama 8 detik."},
				{Role: models.ChatRoleUser, Content: "Bagaimana kalau rasa cemasnya tidak hilang-hilang?"},
				{Role: models.ChatRoleAI, Content: "Jika kecemasan berlangsung terus-menerus dan mengganggu aktivitas sehari-hari, ada beberapa hal yang bisa Anda coba:\n\n1. Journaling - tulis perasaan Anda\n2. Olahraga ringan seperti jalan kaki\n3. Batasi kafein dan alkohol\n4. Bicara dengan orang yang Anda percaya\n\nJika masih berlanjut, pertimbangkan untuk berkonsultasi dengan profesional kesehatan mental. Apakah ada hal spesifik yang membuat Anda cemas?"},
			},
		},
		{
			Title: "Curhat tentang pekerjaan",
			Messages: []struct {
				Role    models.ChatRole
				Content string
			}{
				{Role: models.ChatRoleUser, Content: "Pekerjaan saya sangat menumpuk dan saya merasa overwhelmed."},
				{Role: models.ChatRoleAI, Content: "Terdengar melelahkan. Beban kerja yang berlebihan bisa sangat menguras energi mental dan fisik. Apakah Anda sudah mencoba memprioritaskan tugas-tugas tersebut? Kadang memecah pekerjaan besar menjadi langkah-langkah kecil bisa membantu."},
				{Role: models.ChatRoleUser, Content: "Sudah coba tapi tetap saja rasanya nggak ada habisnya."},
				{Role: models.ChatRoleAI, Content: "Saya mengerti frustasi Anda. Beberapa tips yang mungkin membantu:\n\n1. Gunakan teknik Pomodoro (kerja 25 menit, istirahat 5 menit)\n2. Delegasikan tugas jika memungkinkan\n3. Komunikasikan beban kerja Anda ke atasan\n4. Tetapkan batasan yang jelas - kerja di jam kerja saja\n\nYang terpenting, jangan lupa untuk beristirahat. Burnout adalah masalah serius. Apakah ada waktu akhir-akhir ini Anda mengambil cuti?"},
			},
		},
		{
			Title: "Tips untuk tidur lebih baik",
			Messages: []struct {
				Role    models.ChatRole
				Content string
			}{
				{Role: models.ChatRoleUser, Content: "Saya susah tidur akhir-akhir ini. Ada tips?"},
				{Role: models.ChatRoleAI, Content: "Insomnia atau kesulitan tidur memang sangat mengganggu. Berikut beberapa tips yang bisa membantu:\n\n1. **Jadwal tidur konsisten** - bangun dan tidur di jam yang sama setiap hari\n2. **Hindari layar** 1 jam sebelum tidur\n3. **Ciptakan lingkungan tidur yang nyaman** - gelap, sejuk, dan tenang\n4. **Hindari kafein** setelah jam 2 siang\n5. **Teknik relaksasi** - coba meditasi atau pernapasan 4-7-8 sebelum tidur\n\nApakah ada faktor tertentu yang menurut Anda menyebabkan sulit tidur?"},
			},
		},
	}

	for _, cd := range chatData {
		var existing models.ChatSession
		if db.Where("title = ? AND user_id = ?", cd.Title, user.ID).First(&existing).RowsAffected > 0 {
			continue
		}

		session := models.ChatSession{
			UserID: user.ID,
			Title:  cd.Title,
		}

		if err := db.Create(&session).Error; err != nil {
			return err
		}

		// Create messages
		for i, msg := range cd.Messages {
			message := models.ChatMessage{
				ChatSessionID: session.ID,
				Role:          msg.Role,
				Content:       msg.Content,
				CreatedAt:     time.Now().Add(time.Duration(i) * time.Minute),
			}
			if err := db.Create(&message).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
