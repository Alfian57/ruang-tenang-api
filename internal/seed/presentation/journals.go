package presentation

import (
	"errors"
	"strings"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// SeedJournals adds realistic presentation entries only to fixed demo users.
// Public entries are deliberately low-sensitivity, clearly labelled examples,
// and none is shared with AI by default.
func SeedJournals(db *gorm.DB) error {
	allUsers, err := presentationUsers(db)
	if err != nil {
		return err
	}
	var users []model.User
	for _, user := range allUsers {
		if user.Role == model.RoleUser {
			users = append(users, user)
		}
	}
	if len(users) == 0 {
		return nil
	}
	userIDs := make([]uint, len(users))
	for i, user := range users {
		userIDs[i] = user.ID
	}
	legacyTitles := []string{
		"Hari Pertama Mencoba Journaling", "Rasa Syukur di Pagi Hari", "Hari yang Menantang",
		"Refleksi Akhir Pekan", "Belajar Menerima Diri Sendiri", "Memulai Hari dengan Meditasi",
		"Progres Kecil Tetap Progres", "Menulis untuk Melegakan Hati",
	}
	if len(userIDs) > 0 {
		oldJournals := db.Model(&model.Journal{}).Select("id").Where("user_id IN ? AND title IN ?", userIDs, legacyTitles)
		if err := db.Where("journal_id IN (?)", oldJournals).Delete(&model.JournalAIAccessLog{}).Error; err != nil {
			return err
		}
		if err := db.Where("user_id IN ? AND title IN ?", userIDs, legacyTitles).Delete(&model.Journal{}).Error; err != nil {
			return err
		}
	}
	// Presentation journals default to no AI access, including entries left by
	// older seed versions. The one explicit opt-in below is applied afterward.
	if err := db.Model(&model.Journal{}).Where("user_id IN ?", userIDs).Update("share_with_ai", false).Error; err != nil {
		return err
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
			UserIdx: 0, Title: "Catatan malam: hal yang belum selesai",
			Content: "Hari ini saya kembali merasa daftar tugas lebih panjang daripada waktu yang tersedia. Saya menulis mana yang benar-benar perlu ditangani besok dan mana yang bisa menunggu. Tidak semua hal selesai, tetapi setidaknya sekarang semuanya tidak berputar di kepala pada saat yang sama. Besok saya akan mencoba mulai dari satu langkah kecil.",
			Tags:    []string{"refleksi", "beban-pikiran", "prioritas"}, DaysAgo: 1,
			IsPrivate: true, ShareWithAI: true, // Explicit example of per-entry sharing; user settings permit AI access.
		},
		{
			UserIdx: 0, Title: "Jeda sebelum membalas pesan kerja",
			Content: "Ada pesan masuk ketika saya sedang makan. Refleks pertama saya adalah langsung membalas, tetapi saya memilih menghabiskan makan terlebih dahulu lalu memberi kabar kapan saya bisa menindaklanjuti. Rasanya agak tidak nyaman, namun saya belajar bahwa jeda singkat juga bagian dari cara menjaga kapasitas.",
			Tags:    []string{"batas-sehat", "kerja", "refleksi"}, DaysAgo: 3, IsPrivate: true,
		},
		{
			UserIdx: 0, Title: "[Contoh] Satu hal kecil yang terasa cukup",
			Content: "Contoh jurnal publik akun presentasi. Sore ini saya berjalan sebentar tanpa tujuan khusus dan melihat langit berubah warna. Tidak ada masalah yang selesai karena berjalan, tetapi jeda itu memberi ruang untuk memperhatikan apa yang sedang saya rasakan. Untuk malam ini, itu sudah cukup.",
			Tags:    []string{"jeda", "keseharian", "contoh-publik"}, DaysAgo: 5, IsPrivate: false,
		},
		{
			UserIdx: 0, Title: "Menyusun ulang rencana minggu ini",
			Content: "Beberapa rencana minggu ini berubah. Saya sempat menilai diri sendiri karena tidak menjalankan semuanya sesuai daftar. Setelah melihat ulang, ada hal yang memang bergantung pada orang lain dan ada yang tidak lagi mendesak. Saya mengganti target minggu ini menjadi lebih sederhana dan menambahkan waktu istirahat di kalender.",
			Tags:    []string{"fleksibilitas", "rencana", "self-compassion"}, DaysAgo: 8, IsPrivate: true,
		},
		{
			UserIdx: 0, Title: "[Contoh] Rutinitas malam yang realistis",
			Content: "Contoh jurnal publik akun presentasi. Saya mencoba menyiapkan barang untuk besok dan membaca sebentar sebelum tidur. Ada malam ketika rutinitas itu tidak berjalan, dan saya tidak ingin mengubahnya menjadi aturan yang membuat saya merasa gagal. Saya akan mulai lagi ketika memungkinkan.",
			Tags:    []string{"tidur", "rutinitas", "contoh-publik"}, DaysAgo: 11, IsPrivate: false,
		},
		{
			UserIdx: 1, Title: "Mengecek kebutuhan sebelum berkata iya",
			Content: "Hari ini saya diminta membantu beberapa hal sekaligus. Sebelum menjawab, saya mengecek jadwal dan bertanya mana yang paling mendesak. Ternyata ada satu tugas yang bisa dikerjakan orang lain dan satu yang bisa dijadwalkan ulang. Saya masih ingin membantu, tetapi sekarang saya mencoba melakukannya dengan melihat kapasitas yang ada.",
			Tags:    []string{"kapasitas", "batas", "kerja"}, DaysAgo: 2, IsPrivate: true,
		},
		{
			UserIdx: 1, Title: "[Contoh] Menerima perubahan ritme",
			Content: "Contoh jurnal publik akun presentasi. Pekan ini tidak serapi yang saya rencanakan. Ada beberapa hari ketika tenaga terasa berbeda, jadi saya memindahkan sebagian kegiatan dan tetap menyisakan waktu untuk hal yang penting. Saya ingin belajar mengukur kemajuan dari hal yang sungguh dapat saya lakukan, bukan hanya dari daftar yang penuh.",
			Tags:    []string{"ritme", "fleksibilitas", "contoh-publik"}, DaysAgo: 6, IsPrivate: false,
		},
		{
			UserIdx: 1, Title: "Menulis tanpa harus menemukan solusi",
			Content: "Saya menulis apa yang terjadi dan perasaan yang muncul tanpa mencoba memperbaiki semuanya sekarang. Setelah itu saya bertanya apakah saya ingin didengar, butuh bantuan praktis, atau memang ingin menyusun langkah. Jawabannya tidak selalu sama. Mengetahui perbedaannya membantu saya menyampaikan kebutuhan dengan lebih jelas.",
			Tags:    []string{"emosi", "menulis", "kebutuhan"}, DaysAgo: 9, IsPrivate: true,
		},
		{
			UserIdx: 2, Title: "Membuat pekan kuliah terasa lebih teratur",
			Content: "Saya menuliskan jadwal kelas, tenggat yang pasti, dan waktu yang dibutuhkan untuk perjalanan. Setelah itu saya memilih dua pekerjaan paling penting untuk hari pertama. Ternyata melihat jadwal dengan lebih lengkap membantu saya menyadari bahwa tidak semuanya harus dimulai malam ini. Kalau rencana berubah, saya bisa menyusunnya ulang.",
			Tags:    []string{"kuliah", "prioritas", "perencanaan"}, DaysAgo: 4, IsPrivate: true,
		},
		{
			UserIdx: 2, Title: "[Contoh] Mengakui progres yang kecil",
			Content: "Contoh jurnal publik akun presentasi. Hari ini saya membuka dokumen tugas dan menulis kerangka selama beberapa menit. Saya belum menyelesaikan tugas itu, tetapi memulainya mengurangi jarak antara niat dan tindakan. Besok saya akan melihat kembali kerangka tersebut dan menentukan langkah berikutnya.",
			Tags:    []string{"progres-kecil", "kuliah", "contoh-publik"}, DaysAgo: 7, IsPrivate: false,
		},
	}

	now := time.Now().UTC()
	for _, entry := range entries {
		if entry.UserIdx >= len(users) {
			continue
		}
		author := users[entry.UserIdx]
		createdAt := now.AddDate(0, 0, -entry.DaysAgo)
		journal := model.Journal{
			UserID: author.ID, Title: entry.Title, Content: entry.Content,
			Tags: pq.StringArray(entry.Tags), IsPrivate: entry.IsPrivate,
			ShareWithAI: entry.ShareWithAI, WordCount: len(strings.Fields(entry.Content)),
			CreatedAt: createdAt, UpdatedAt: createdAt,
		}
		var existing model.Journal
		findResult := db.Where("title = ? AND user_id = ?", entry.Title, author.ID).First(&existing)
		if errors.Is(findResult.Error, gorm.ErrRecordNotFound) {
			if err := db.Create(&journal).Error; err != nil {
				return err
			}
			continue
		}
		if findResult.Error != nil {
			return findResult.Error
		}
		if err := db.Model(&existing).Updates(map[string]any{
			"content": entry.Content, "tags": pq.StringArray(entry.Tags),
			"is_private": entry.IsPrivate, "share_with_ai": entry.ShareWithAI,
			"word_count": journal.WordCount, "created_at": createdAt, "updated_at": createdAt,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}
