package production

import (
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedInspiringStories seeds polished inspiring stories for production.
// Requires Default Accounts and Story Categories to be seeded first.
func SeedInspiringStories(db *gorm.DB) error {
	// Get member users to use as authors
	var members []model.User
	if err := db.Where("role = ?", model.RoleMember).Order("id ASC").Find(&members).Error; err != nil {
		return err
	}
	if len(members) < 2 {
		return nil // Need at least 2 members
	}

	// Get admin for moderation fields
	var moderator model.User
	if err := db.Where("role = ?", model.RoleAdmin).First(&moderator).Error; err != nil {
		return nil
	}

	// Get story categories
	var categories []model.StoryCategory
	if err := db.Where("is_active = ?", true).Order("display_order ASC").Find(&categories).Error; err != nil {
		return err
	}
	if len(categories) == 0 {
		return nil
	}

	categoryMap := make(map[string]model.StoryCategory)
	for _, c := range categories {
		categoryMap[c.Slug] = c
	}

	now := time.Now()
	modTime := now.Add(-24 * time.Hour)
	pubTime := now.Add(-20 * time.Hour)

	stories := []struct {
		AuthorIdx     int
		Title         string
		Content       string
		IsAnonymous   bool
		HasTrigger    bool
		TriggerText   string
		IsFeatured    bool
		CategorySlugs []string
		Tags          []string
		ViewCount     int
		HeartCount    int
		CommentCount  int
	}{
		{
			AuthorIdx: 0,
			Title:     "Dari Kegelapan Menuju Cahaya: Perjalanan Saya Melawan Kecemasan",
			Content: `Dua tahun lalu, saya hampir tidak bisa keluar rumah. Setiap kali mencoba membuka pintu, jantung saya berdebar sangat kencang, tangan berkeringat, dan pikiran dipenuhi skenario terburuk. Kecemasan sosial yang saya alami membuat dunia terasa begitu menakutkan.

Titik balik terjadi ketika seorang teman dekat mengajak saya mencoba teknik pernapasan sederhana—teknik 4-7-8. Awalnya saya skeptis, tapi saya coba karena tidak ada lagi yang bisa saya lakukan. Perlahan tapi pasti, latihan pernapasan itu menjadi ritual harian saya.

Saya mulai menulis jurnal tentang perasaan saya setiap hari. Awalnya hanya beberapa kalimat, tapi lama-kelamaan jurnal itu menjadi tempat curhat yang aman. Saya bisa jujur tentang ketakutan saya tanpa dihakimi.

Langkah terbesar adalah ketika saya memutuskan untuk menemui psikolog. Butuh keberanian besar, tapi itu adalah keputusan terbaik dalam hidup saya. Terapis saya membantu saya memahami bahwa kecemasan bukanlah kelemahan—itu adalah sinyal dari tubuh yang perlu dipahami.

Sekarang, saya masih mengalami kecemasan. Tapi saya sudah punya "toolkit" untuk menghadapinya: pernapasan, jurnal, olahraga, dan dukungan profesional. Bagi siapa pun yang membaca ini dan sedang berjuang—Anda tidak sendirian, dan langkah pertama tidak harus besar. Cukup satu napas dalam saja.`,
			IsFeatured:    true,
			CategorySlugs: []string{"anxiety-management", "recovery-journey"},
			Tags:          []string{"kecemasan", "pernapasan", "journaling", "terapi"},
			ViewCount:     156,
			HeartCount:    3,
			CommentCount:  2,
		},
		{
			AuthorIdx: 1,
			Title:     "Belajar Mencintai Diri Sendiri: Perjalanan Self-Care Saya",
			Content: `Selama bertahun-tahun, saya selalu mendahulukan orang lain. Sebagai anak sulung, saya merasa harus selalu kuat dan menjadi panutan. Tapi di balik senyum, saya kelelahan. Burnout datang tanpa peringatan.

Suatu hari, tubuh saya benar-benar menyerah. Saya tidak bisa bangun dari tempat tidur selama tiga hari. Itu adalah wake-up call terbesar dalam hidup saya.

Saya mulai belajar tentang self-care—bukan yang glamor seperti di media sosial, tapi yang sederhana dan nyata. Tidur cukup. Makan teratur. Mengatakan "tidak" tanpa rasa bersalah. Meluangkan waktu untuk hal yang saya nikmati.

Saya menemukan bahwa mendengarkan musik yang menenangkan sebelum tidur sangat membantu. Suara alam, piano lembut, atau white noise menjadi teman setia saya. Saya juga mulai rutin berjalan kaki di pagi hari—hanya 15 menit, tapi efeknya luar biasa untuk mood saya.

Hal terpenting yang saya pelajari: self-care bukan egois. Anda tidak bisa menuangkan dari gelas yang kosong. Sekarang saya lebih baik dalam menjaga batasan dan mendengarkan kebutuhan tubuh saya. Dan percayalah, hidup terasa jauh lebih ringan.`,
			IsFeatured:    false,
			CategorySlugs: []string{"self-care-journey", "finding-hope"},
			Tags:          []string{"self-care", "burnout", "batasan", "musik"},
			ViewCount:     98,
			HeartCount:    2,
			CommentCount:  1,
		},
		{
			AuthorIdx: 0,
			Title:     "Ketika Terapi Mengubah Hidup Saya",
			Content: `Saya pernah berpikir bahwa pergi ke psikolog berarti saya "gila". Stigma itu begitu kuat di lingkungan saya. Keluarga saya bilang, "Kamu cuma kurang bersyukur." Teman-teman bilang, "Itu cuma fase."

Tapi perasaan hampa yang terus-menerus itu bukan "cuma fase." Saya merasa seperti menonton hidup saya dari luar—tidak benar-benar hadir, tidak benar-benar merasakan apa-apa. Numbernya? Hampir dua tahun saya hidup seperti itu.

Akhirnya, dengan keberanian yang tersisa, saya membuat janji dengan psikolog klinis. Sesi pertama sangat menegangkan. Tapi terapis saya luar biasa sabar. Dia tidak pernah menghakimi. Dia mendengarkan—benar-benar mendengarkan.

Melalui CBT (Cognitive Behavioral Therapy), saya belajar mengenali pola pikir negatif saya. Saya belajar bahwa pikiran bukan fakta. Dan perlahan, dinding yang saya bangun selama bertahun-tahun mulai runtuh.

Sekarang, setelah hampir setahun terapi, saya bisa bilang: terapi adalah investasi terbaik yang pernah saya lakukan untuk diri sendiri. Jika Anda ragu untuk mencoba, izinkan saya menjadi bukti bahwa itu worth it. Anda layak mendapat bantuan.`,
			IsAnonymous:   true,
			HasTrigger:    true,
			TriggerText:   "Konten ini membahas tentang perasaan hampa dan pengalaman depresi.",
			IsFeatured:    false,
			CategorySlugs: []string{"professional-help", "overcoming-depression"},
			Tags:          []string{"terapi", "CBT", "stigma", "depresi"},
			ViewCount:     212,
			HeartCount:    2,
			CommentCount:  2,
		},
		{
			AuthorIdx: 1,
			Title:     "Menemukan Harapan di Tengah Badai",
			Content: `Pandemi mengubah segalanya. Saya kehilangan pekerjaan, rutinitas, dan yang paling berat—saya kehilangan koneksi sosial. Isolasi membuat pikiran saya menjadi musuh terbesar saya.

Di masa-masa gelap itu, saya menemukan sebuah komunitas online tentang kesehatan mental. Untuk pertama kalinya, saya merasa tidak sendirian. Ada orang-orang yang memahami apa yang saya rasakan tanpa saya harus menjelaskan panjang lebar.

Saya mulai mencoba hal-hal kecil: meditasi 5 menit di pagi hari, menulis 3 hal yang saya syukuri sebelum tidur, dan menelepon satu orang setiap hari. Langkah-langkah kecil ini pelan-pelan membawa perubahan besar.

Yang paling membantu adalah menulis. Saya mulai menulis jurnal setiap malam. Menuangkan semua yang ada di kepala ke atas kertas (atau layar) terasa seperti melepas beban dari pundak. Journal ini juga membantu saya melacak pola mood saya—kapan saya merasa baik, kapan saya merasa buruk, dan apa pemicunya.

Hari ini, saya sudah bekerja lagi dan memiliki rutinitas yang sehat. Saya masih punya hari-hari buruk, tapi sekarang saya tahu: badai selalu berlalu, dan setelah hujan selalu ada pelangi. Jika Anda sedang di tengah badai sekarang, bertahanlah. Cahaya itu ada, meskipun belum terlihat.`,
			IsFeatured:    true,
			CategorySlugs: []string{"finding-hope", "recovery-journey"},
			Tags:          []string{"pandemi", "komunitas", "gratitude", "journaling"},
			ViewCount:     134,
			HeartCount:    3,
			CommentCount:  1,
		},
		{
			AuthorIdx: 0,
			Title:     "Healing Bukan Garis Lurus",
			Content: `Saya ingin jujur: perjalanan healing itu tidak semulus yang sering digambarkan di media sosial. Tidak ada "sebelum dan sesudah" yang dramatis. Healing itu berantakan, penuh kemunduran, dan kadang terasa seperti berjalan di tempat.

Saya pernah merasa sangat baik selama berbulan-bulan, lalu tiba-tiba jatuh lagi. Dan itu membuat saya merasa gagal. "Kenapa saya kembali ke titik nol?" Tapi terapis saya mengajarkan sesuatu yang mengubah perspektif saya: Anda tidak kembali ke titik nol. Anda kembali dengan lebih banyak pengalaman dan alat untuk menghadapinya.

Sekarang saya menerima bahwa healing bukan tujuan akhir—itu adalah proses seumur hidup. Ada hari-hari baik dan ada hari-hari buruk. Dan keduanya valid.

Yang saya pelajari:
- Tidak apa-apa untuk meminta bantuan
- Kemunduran bukan kegagalan
- Membandingkan perjalanan Anda dengan orang lain itu tidak adil
- Setiap langkah kecil tetaplah langkah maju
- Self-compassion adalah keterampilan yang bisa dipelajari

Terima kasih sudah membaca. Jika cerita ini membuat satu orang merasa tidak sendirian, maka itu sudah cukup bagi saya.`,
			IsFeatured:    false,
			CategorySlugs: []string{"healing-from-trauma", "recovery-journey"},
			Tags:          []string{"healing", "kemunduran", "self-compassion", "proses"},
			ViewCount:     178,
			HeartCount:    2,
			CommentCount:  1,
		},
	}

	comments := []struct {
		StoryIdx  int
		AuthorIdx int
		Content   string
	}{
		{0, 1, "Terima kasih sudah berbagi cerita ini. Saya juga mengalami hal serupa dan teknik pernapasan sangat membantu saya! 💙"},
		{0, 0, "Sangat menginspirasi. Semoga semakin banyak orang yang berani mengambil langkah pertama."},
		{1, 0, "Self-care memang bukan egois. Terima kasih sudah mengingatkan kita semua! 🌸"},
		{2, 1, "Cerita ini sangat berani. Terima kasih sudah membantu menghapus stigma tentang terapi."},
		{2, 0, "Saya akhirnya memutuskan mencoba terapi setelah membaca ini. Terima kasih. ❤️"},
		{3, 0, "Journaling benar-benar membantu saya juga! Menulis itu terapi tersendiri."},
		{4, 1, "\"Healing bukan garis lurus\" — kalimat ini sangat saya butuhkan hari ini. Terima kasih."},
	}

	for i, s := range stories {
		// Check if already seeded
		var existing model.InspiringStory
		if db.Where("title = ?", s.Title).First(&existing).RowsAffected > 0 {
			continue
		}

		author := members[s.AuthorIdx%len(members)]
		moderatorID := moderator.ID

		story := model.InspiringStory{
			AuthorID:           author.ID,
			Title:              s.Title,
			Content:            s.Content,
			IsAnonymous:        s.IsAnonymous,
			HasTriggerWarning:  s.HasTrigger,
			TriggerWarningText: s.TriggerText,
			Status:             model.StoryStatusApproved,
			ModeratorID:        &moderatorID,
			ModerationFeedback: "Cerita yang sangat menginspirasi. Disetujui.",
			ModeratedAt:        &modTime,
			ViewCount:          s.ViewCount,
			HeartCount:         s.HeartCount,
			CommentCount:       s.CommentCount,
			IsFeatured:         s.IsFeatured,
			PublishedAt:        &pubTime,
		}

		if s.IsFeatured {
			featuredAt := now.Add(-18 * time.Hour)
			story.FeaturedAt = &featuredAt
		}

		if err := db.Create(&story).Error; err != nil {
			return err
		}

		// Create category relations
		for _, slug := range s.CategorySlugs {
			if cat, ok := categoryMap[slug]; ok {
				relation := model.StoryCategoryRelation{
					StoryID:    story.ID,
					CategoryID: cat.ID,
				}
				if err := db.Create(&relation).Error; err != nil {
					return err
				}
			}
		}

		// Create tags
		for _, tag := range s.Tags {
			storyTag := model.StoryTag{
				StoryID: story.ID,
				Tag:     tag,
			}
			if err := db.Create(&storyTag).Error; err != nil {
				return err
			}
		}

		// Create hearts from other users
		heartCount := 0
		for _, member := range members {
			if member.ID == author.ID {
				continue
			}
			if heartCount >= s.HeartCount {
				break
			}
			heart := model.StoryHeart{
				StoryID: story.ID,
				UserID:  member.ID,
			}
			if err := db.Create(&heart).Error; err != nil {
				return err
			}
			heartCount++
		}

		// Create comments for this story
		for _, c := range comments {
			if c.StoryIdx != i {
				continue
			}
			commenter := members[c.AuthorIdx%len(members)]
			comment := model.StoryComment{
				StoryID: story.ID,
				UserID:  commenter.ID,
				Content: c.Content,
			}
			if err := db.Create(&comment).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
