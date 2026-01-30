package development

import (
	"math/rand"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"gorm.io/gorm"
)

// SeedForums seeds test forum threads for development
func SeedForums(db *gorm.DB) error {
	// Get users for creating posts
	var users []models.User
	if err := db.Where("role = ?", models.RoleMember).Limit(10).Find(&users).Error; err != nil {
		return err
	}
	if len(users) == 0 {
		return nil // No users to create forums
	}

	// Get categories
	var categories []models.ForumCategory
	if err := db.Find(&categories).Error; err != nil {
		return err
	}

	forumData := []struct {
		Title    string
		Content  string
		Category string
	}{
		{
			Title:    "Bagaimana cara kalian mengatasi burnout?",
			Content:  "Belakangan ini saya merasa sangat lelah dengan pekerjaan. Setiap hari rasanya berat banget untuk bangun dan memulai aktivitas. Ada tips untuk mengatasi burnout tanpa harus resign? Terima kasih sebelumnya 🙏",
			Category: "Kesehatan Mental di Tempat Kerja",
		},
		{
			Title:    "Cerita sukses sembuh dari anxiety",
			Content:  "Saya ingin berbagi pengalaman saya sembuh dari anxiety disorder setelah 2 tahun berjuang. Dulu saya sering panic attack, tangan gemetar, dan susah tidur. Sekarang alhamdulillah sudah jauh lebih baik. Semoga bisa menginspirasi teman-teman semua. Feel free to ask if ada pertanyaan!",
			Category: "Kisah Inspiratif",
		},
		{
			Title:    "Butuh teman curhat",
			Content:  "Lagi merasa down banget hari ini. Banyak masalah yang menumpuk dan rasanya nggak ada yang ngerti. Ada yang bersedia mendengarkan? 😔",
			Category: "Curhat & Keluh Kesah",
		},
		{
			Title:    "Rekomendasi buku self-improvement",
			Content:  "Ada rekomendasi buku bagus untuk meningkatkan kepercayaan diri? Preferably yang ada versi Bahasa Indonesia-nya. Budget nggak masalah selama memang worth it. Thanks!",
			Category: "Diskusi Umum",
		},
		{
			Title:    "Tips menghadapi social anxiety di tempat kerja",
			Content:  "Baru mulai kerja di kantor baru dan social anxiety saya kumat. Susah banget buat ngobrol sama rekan kerja dan selalu overthinking kalau habis ngomong sesuatu. Ada yang punya tips?",
			Category: "Tips Mengelola Stres",
		},
		{
			Title:    "Apakah terapi worth it?",
			Content:  "Saya sedang mempertimbangkan untuk mencoba terapi/konseling. Buat yang sudah pernah coba, apakah worth it? Berapa lama biasanya sampai terasa manfaatnya?",
			Category: "Pertanyaan & Jawaban",
		},
		{
			Title:    "Meditasi pagi bikin produktif!",
			Content:  "Sudah 30 hari konsisten meditasi setiap pagi 10 menit. Hasilnya: lebih fokus, nggak gampang emosi, dan tidur lebih nyenyak. Highly recommend buat yang belum coba! 🧘‍♀️",
			Category: "Kisah Inspiratif",
		},
		{
			Title:    "Stress karena deadline skripsi",
			Content:  "Deadline skripsi tinggal sebulan tapi masih banyak yang belum selesai. Setiap mau ngerjain malah prokrastinasi. Gimana ya cara mengatasinya? 😭",
			Category: "Kesehatan Mental di Sekolah",
		},
	}

	replies := []string{
		"Semangat ya! Kamu nggak sendirian. Kalau butuh apa-apa, feel free to reach out 💙",
		"Terima kasih sudah berbagi. Ini sangat menginspirasi!",
		"Saya juga pernah mengalami hal serupa. Yang penting jangan menyerah dan terus mencoba.",
		"Coba teknik pomodoro untuk mengatasi prokrastinasi. Kerja 25 menit, istirahat 5 menit.",
		"Jangan lupa untuk istirahat juga ya. Kesehatan mental itu penting banget.",
		"I feel you. Semoga lekas membaik! 🤗",
		"Pengalaman yang sangat berharga. Terima kasih sudah mau berbagi.",
		"Semangat! Satu langkah kecil setiap hari itu sudah progress yang bagus.",
	}

	for _, fd := range forumData {
		var existing models.Forum
		if db.Where("title = ?", fd.Title).First(&existing).RowsAffected > 0 {
			continue
		}

		// Find category ID
		var catID *uint
		for _, cat := range categories {
			if cat.Name == fd.Category {
				catID = &cat.ID
				break
			}
		}

		// Random user as author
		author := users[rand.Intn(len(users))]

		forum := models.Forum{
			UserID:     author.ID,
			CategoryID: catID,
			Title:      fd.Title,
			Content:    fd.Content,
			CreatedAt:  time.Now().Add(-time.Duration(rand.Intn(168)) * time.Hour), // Random within last week
		}

		if err := db.Create(&forum).Error; err != nil {
			return err
		}

		// Create random replies
		numReplies := rand.Intn(5) + 1
		for i := 0; i < numReplies; i++ {
			replyUser := users[rand.Intn(len(users))]
			replyContent := replies[rand.Intn(len(replies))]

			post := models.ForumPost{
				ForumID:   forum.ID,
				UserID:    replyUser.ID,
				Content:   replyContent,
				CreatedAt: time.Now().Add(-time.Duration(rand.Intn(48)) * time.Hour),
			}
			db.Create(&post)
		}

		// Add random likes
		numLikes := rand.Intn(8)
		usedLikers := make(map[uint]bool)
		for i := 0; i < numLikes; i++ {
			likeUser := users[rand.Intn(len(users))]
			if usedLikers[likeUser.ID] || likeUser.ID == author.ID {
				continue
			}
			usedLikers[likeUser.ID] = true
			db.Create(&models.ForumLike{ForumID: forum.ID, UserID: likeUser.ID})
		}
	}

	return nil
}
