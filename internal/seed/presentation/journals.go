package presentation

import (
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// SeedJournals seeds sample journal entries for development
func SeedJournals(db *gorm.DB) error {
	var users []model.User
	if err := db.Where("role = ?", model.RoleUser).Order("id ASC").Find(&users).Error; err != nil {
		return err
	}
	if len(users) == 0 {
		return nil
	}

	type journalData struct {
		UserIdx     int
		Title       string
		Content     string
		Tags        []string
		DaysAgo     int
		IsPrivate   bool
		ShareWithAI bool
	}

	entries := []journalData{
		{
			UserIdx: 0,
			Title:   "Hari Pertama Mencoba Journaling",
			Content: "Hari ini saya memutuskan untuk mulai menulis jurnal. Sebenarnya saya sudah lama ingin mencoba, tapi selalu menunda. Hari ini rasanya cukup baik—tidak terlalu senang, tidak terlalu sedih. Mungkin itu yang namanya 'normal'? Saya berharap dengan menulis, saya bisa lebih memahami perasaan saya sendiri. Semoga ini bisa menjadi kebiasaan yang baik.",
			Tags:    []string{"journaling", "awal", "refleksi"},
			DaysAgo: 12,
			// Privat penuh, tidak dibagikan ke AI.
			IsPrivate:   true,
			ShareWithAI: false,
		},
		{
			UserIdx: 0,
			Title:   "Rasa Syukur di Pagi Hari",
			Content: "Pagi ini saya bangun lebih awal dan menikmati kopi sambil melihat matahari terbit. Kecil memang, tapi rasanya damai sekali. Saya bersyukur untuk hal-hal kecil: udara segar, secangkir kopi hangat, dan ketenangan pagi hari. Kadang kita terlalu sibuk mencari kebahagiaan besar sampai lupa bahwa kebahagiaan kecil ada di mana-mana.",
			Tags:    []string{"gratitude", "pagi", "ketenangan"},
			DaysAgo: 10,
			// Publik ke komunitas, juga dibagikan ke AI.
			IsPrivate:   false,
			ShareWithAI: true,
		},
		{
			UserIdx: 0,
			Title:   "Hari yang Menantang",
			Content: "Hari ini cukup berat. Ada banyak tekanan dari tugas kuliah dan saya merasa overwhelmed. Tapi saya ingat untuk bernapas dan mengambil jeda sebentar. Teknik pernapasan 4-7-8 yang saya pelajari ternyata membantu. Saya juga mencoba untuk memecah tugas besar menjadi langkah-langkah kecil. Besok saya akan mencoba lagi dengan lebih baik.",
			Tags:    []string{"tekanan", "pernapasan", "coping"},
			DaysAgo: 7,
			// Privat tapi dibagikan ke AI untuk dukungan personal.
			IsPrivate:   true,
			ShareWithAI: true,
		},
		{
			UserIdx: 0,
			Title:   "Refleksi Akhir Pekan",
			Content: "Akhir pekan ini saya menghabiskan waktu dengan teman-teman lama. Kami hanya duduk di kafe, ngobrol, dan tertawa. Sederhana tapi sangat menyenangkan. Saya sadar bahwa koneksi sosial itu penting untuk kesehatan mental saya. Minggu depan saya ingin lebih sering meluangkan waktu untuk orang-orang yang saya sayangi.",
			Tags:    []string{"teman", "sosial", "kebahagiaan"},
			DaysAgo: 5,
			// Publik ke komunitas, tidak dibagikan ke AI.
			IsPrivate:   false,
			ShareWithAI: false,
		},
		{
			UserIdx: 0,
			Title:   "Belajar Menerima Diri Sendiri",
			Content: "Hari ini saya menyadari bahwa saya sering terlalu keras pada diri sendiri. Ketika melakukan kesalahan kecil, saya langsung menyalahkan diri sendiri. Tapi hari ini saya mencoba untuk lebih berbelas kasih. Saya berkata pada diri sendiri: 'Tidak apa-apa, kamu sedang belajar.' Dan rasanya... lega. Self-compassion itu ternyata bukan kelemahan, tapi kekuatan.",
			Tags:    []string{"self-compassion", "penerimaan", "pertumbuhan"},
			DaysAgo: 3,
			IsPrivate:   true,
			ShareWithAI: false,
		},
		{
			UserIdx: 1,
			Title:   "Memulai Hari dengan Meditasi",
			Content: "Hari ini saya mencoba meditasi untuk pertama kalinya. Hanya 5 menit, tapi rasanya sangat berbeda. Pikiran saya biasanya sangat ramai di pagi hari, tapi setelah meditasi, terasa lebih tenang. Saya ingin mencoba untuk konsisten melakukannya setiap pagi.",
			Tags:    []string{"meditasi", "pagi", "mindfulness"},
			DaysAgo: 9,
			// Publik + AI: contoh jurnal komunitas yang ramah dibagikan.
			IsPrivate:   false,
			ShareWithAI: true,
		},
		{
			UserIdx: 1,
			Title:   "Progres Kecil Tetap Progres",
			Content: "Minggu ini saya berhasil tidur lebih awal 3 dari 7 hari. Mungkin terdengar sedikit, tapi bagi saya itu adalah kemajuan besar. Saya dulu selalu tidur jam 2-3 pagi. Sekarang saya mencoba untuk tidur sebelum tengah malam. Langkah kecil, tapi konsisten. Saya bangga dengan diri sendiri hari ini.",
			Tags:    []string{"tidur", "progres", "kebiasaan"},
			DaysAgo: 6,
			IsPrivate:   false,
			ShareWithAI: false,
		},
		{
			UserIdx: 1,
			Title:   "Menulis untuk Melegakan Hati",
			Content: "Ada perasaan yang sulit saya ungkapkan hari ini. Bukan sedih, bukan marah, tapi semacam... lelah secara emosional. Mungkin karena terlalu banyak berpikir. Tapi menulis ini membantu saya menuangkan beban itu. Kadang kita tidak perlu solusi, kita hanya perlu didengar—bahkan kalau yang mendengar adalah kertas kosong.",
			Tags:    []string{"emosi", "menulis", "katarsis"},
			DaysAgo: 2,
			IsPrivate:   true,
			ShareWithAI: true,
		},
	}

	for _, entry := range entries {
		author := users[entry.UserIdx%len(users)]

		var existing model.Journal
		if db.Where("title = ? AND user_id = ?", entry.Title, author.ID).First(&existing).RowsAffected > 0 {
			continue
		}

		wordCount := len(strings.Fields(entry.Content))
		createdAt := time.Now().AddDate(0, 0, -entry.DaysAgo)

		journal := model.Journal{
			UserID:      author.ID,
			Title:       entry.Title,
			Content:     entry.Content,
			Tags:        pq.StringArray(entry.Tags),
			IsPrivate:   entry.IsPrivate,
			ShareWithAI: entry.ShareWithAI,
			WordCount:   wordCount,
			CreatedAt:   createdAt,
			UpdatedAt:   createdAt,
		}

		if err := db.Create(&journal).Error; err != nil {
			return err
		}
	}

	return nil
}
