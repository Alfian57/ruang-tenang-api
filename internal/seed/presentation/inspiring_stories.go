package presentation

import (
	"errors"
	"fmt"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedInspiringStories fills the presentation feed with editorial composites.
// They are explicitly described as illustrations in their content so the demo
// dataset cannot be mistaken for a real member's recovery testimonial.
func SeedInspiringStories(db *gorm.DB) error {
	allUsers, err := presentationUsers(db)
	if err != nil {
		return err
	}
	var members []model.User
	for _, user := range allUsers {
		if user.Role == model.RoleUser {
			members = append(members, user)
		}
	}
	if len(members) < 2 {
		return nil
	}
	memberIDs := make([]uint, len(members))
	for i, member := range members {
		memberIDs[i] = member.ID
	}
	legacyTitles := []string{
		"Dari Kegelapan Menuju Cahaya: Perjalanan Saya Melawan Kecemasan",
		"Belajar Mencintai Diri Sendiri: Perjalanan Self-Care Saya",
		"Ketika Terapi Mengubah Hidup Saya", "Menemukan Harapan di Tengah Badai", "Healing Bukan Garis Lurus",
	}
	if err := db.Where("title IN ? AND author_id IN ?", legacyTitles, memberIDs).Delete(&model.InspiringStory{}).Error; err != nil {
		return err
	}

	var moderator model.User
	if err := db.Where("email = ? AND role = ?", presentationAdminEmail, model.RoleAdmin).First(&moderator).Error; err != nil {
		return nil
	}
	var categories []model.StoryCategory
	if err := db.Where("is_active = ?", true).Order("display_order ASC").Find(&categories).Error; err != nil {
		return err
	}
	categoryMap := make(map[string]model.StoryCategory, len(categories))
	for _, category := range categories {
		categoryMap[category.Slug] = category
	}

	type commentSeed struct {
		AuthorIdx int
		Content   string
		HeartIdx  int
	}
	type storySeed struct {
		AuthorIdx     int
		Title         string
		Content       string
		HasWarning    bool
		WarningText   string
		IsFeatured    bool
		CategorySlugs []string
		Tags          []string
		DaysAgo       int
		ViewCount     int
		HeartUsers    []int
		Comments      []commentSeed
	}

	illustrationNote := "Ilustrasi editorial komposit untuk dataset presentasi; ini bukan kesaksian satu orang atau nasihat klinis.\n\n"
	stories := []storySeed{
		{
			AuthorIdx: 0, Title: "Langkah kecil ketika hari terasa terlalu penuh",
			Content: illustrationNote + `Dalam ilustrasi ini, seseorang mulai merasa sulit memulai hari setelah beberapa minggu menanggung banyak hal sekaligus. Ia tidak langsung menemukan jawaban besar. Ia mulai dengan memberi nama pada apa yang terasa berat, mengirim pesan kepada teman yang dipercaya, lalu membagi satu urusan menjadi langkah yang lebih kecil.

Ketika kesulitan mulai mengganggu kegiatan sehari-hari, ia mencari bantuan profesional. Dukungan itu tidak membuat setiap hari langsung mudah, tetapi membantunya memahami pilihan dan membangun cara menghadapi hari yang berat.

Cerita ini mengingatkan bahwa meminta ditemani adalah langkah yang sah. Kamu boleh mencari bantuan bahkan ketika belum tahu harus mulai dari mana.`,
			IsFeatured: true, CategorySlugs: []string{"recovery-journey", "finding-hope"},
			Tags: []string{"langkah-kecil", "dukungan", "mencari-bantuan"}, DaysAgo: 4, ViewCount: 82,
			HeartUsers: []int{1, 2}, Comments: []commentSeed{
				{1, "Bagian tentang tidak harus punya jawaban besar terasa menenangkan. Terima kasih sudah mengangkat pentingnya mencari dukungan.", 2},
				{2, "Mengingatkan saya untuk bertanya dulu: bantuan seperti apa yang saya butuhkan hari ini?", 0},
			},
		},
		{
			AuthorIdx: 1, Title: "Belajar membuat batas tanpa harus merasa bersalah",
			Content: illustrationNote + `Tokoh dalam ilustrasi ini terbiasa berkata “iya” sebelum sempat memeriksa tenaganya sendiri. Lama-lama, waktu istirahat terasa seperti sesuatu yang harus diminta maafkan. Ia mulai berlatih menjawab dengan jeda: “Boleh saya cek kapasitas dulu dan kembali dengan jawaban?”

Ia juga mencoba menyepakati prioritas yang lebih jelas di rumah dan di tempat kerja. Ada percakapan yang canggung, dan tidak semua keadaan dapat diubah seketika. Namun batas yang disampaikan dengan tenang memberinya ruang untuk merawat kebutuhan dasar tanpa menganggap dirinya egois.

Menjaga diri bukan berarti tidak peduli pada orang lain. Kadang itu cara agar kepedulian dapat bertahan lebih lama.`,
			IsFeatured: false, CategorySlugs: []string{"self-care-journey", "recovery-journey"},
			Tags: []string{"batas-sehat", "istirahat", "self-care"}, DaysAgo: 7, ViewCount: 56,
			HeartUsers: []int{0, 2}, Comments: []commentSeed{
				{0, "Kalimat untuk mengambil jeda sebelum menjawab ini praktis sekali. Akan saya coba simpan.", 2},
			},
		},
		{
			AuthorIdx: 2, Title: "Memberi jarak pada pikiran yang berisik",
			Content: illustrationNote + `Saat banyak hal datang bersamaan, tokoh dalam cerita ini merasa pikirannya seperti membuka banyak tab sekaligus. Ia mencoba berhenti sebentar, menyebut apa yang sedang ia pikirkan sebagai “kekhawatiran”, lalu mengarahkan perhatian ke telapak kaki yang menyentuh lantai dan benda-benda di sekitarnya.

Latihan grounding tidak menghapus masalah yang sedang dihadapi. Dalam ilustrasi ini, latihan hanya membantu tokoh kembali memperhatikan keadaan sekarang agar ia dapat memilih satu tindakan yang mungkin dilakukan. Jika fokus pada napas atau tubuh terasa tidak nyaman, ia berhenti dan memilih melihat sekitar atau menghubungi orang tepercaya.

Latihan sederhana dapat dicoba, tetapi tidak harus cocok untuk semua orang.`,
			IsFeatured: true, CategorySlugs: []string{"anxiety-management", "finding-hope"},
			Tags: []string{"grounding", "kecemasan", "mindfulness"}, DaysAgo: 10, ViewCount: 113,
			HeartUsers: []int{0, 1}, Comments: []commentSeed{
				{1, "Terima kasih juga sudah menulis bahwa latihan boleh dihentikan kalau terasa tidak nyaman. Itu penting.", 0},
				{0, "Saya suka fokus ke benda di sekitar daripada memaksa harus tenang. Pelan-pelan saja.", 2},
			},
		},
		{
			AuthorIdx: 0, Title: "Mencari bantuan profesional untuk pertama kali",
			Content: illustrationNote + `Dalam kisah ilustratif ini, seseorang ingin berbicara dengan psikolog tetapi khawatir tidak tahu apa yang harus dikatakan. Ia menuliskan tiga hal sebelum membuat janji: apa yang paling mengganggu akhir-akhir ini, kapan hal itu terasa lebih berat, dan perubahan apa yang ingin ia pahami.

Pertemuan pertama menjadi ruang untuk saling mengenal dan membicarakan kebutuhan. Ia berhak bertanya tentang pendekatan, biaya, kerahasiaan, serta langkah berikutnya. Bila belum merasa cocok, ia dapat membicarakannya atau mencari rujukan lain.

Meminta bantuan bukan tanda gagal mengatasi masalah sendiri. Konsultasi juga bukan janji bahwa semua persoalan selesai dalam satu pertemuan; proses dan pilihan bantuan berbeda untuk setiap orang.`,
			HasWarning: true, WarningText: "Menyebut kecemasan dan proses mencari bantuan profesional.",
			CategorySlugs: []string{"professional-help", "anxiety-management"},
			Tags:          []string{"psikolog", "konsultasi", "dukungan-profesional"}, DaysAgo: 13, ViewCount: 67,
			HeartUsers: []int{1, 2}, Comments: []commentSeed{
				{2, "Daftar pertanyaannya membantu mengubah hal yang terasa menakutkan menjadi lebih terarah.", 1},
			},
		},
		{
			AuthorIdx: 1, Title: "Kemajuan tidak selalu terlihat seperti garis lurus",
			Content: illustrationNote + `Tokoh dalam ilustrasi ini pernah mengira kemajuan berarti tidak pernah merasa cemas lagi. Ketika kecemasan muncul kembali, ia merasa semua latihan dan percakapan sebelumnya sia-sia. Dengan bantuan orang yang ia percaya, ia mulai melihat kemajuan dengan ukuran lain: lebih cepat menyadari tanda tubuhnya, lebih berani meminta jeda, dan lebih tahu siapa yang dapat dihubungi.

Hari yang sulit tidak membatalkan hal yang sudah dipelajari. Namun tidak perlu juga memaksa diri menyebut setiap kesulitan sebagai pelajaran. Ada kalanya yang dibutuhkan hanya istirahat, dukungan praktis, atau pertolongan profesional.

Setiap perjalanan berbeda. Tidak ada satu ukuran kemajuan yang wajib dipakai semua orang.`,
			CategorySlugs: []string{"recovery-journey", "finding-hope"},
			Tags:          []string{"proses", "self-compassion", "dukungan"}, DaysAgo: 16, ViewCount: 49,
			HeartUsers: []int{0, 2}, Comments: []commentSeed{
				{0, "Saya suka ukuran kemajuan yang lebih dekat ke keseharian, seperti tahu kapan harus minta jeda.", 2},
			},
		},
	}

	now := time.Now().UTC()
	moderatedAt := now.Add(-time.Hour)
	moderatorID := moderator.ID
	for _, data := range stories {
		author := members[data.AuthorIdx%len(members)]
		createdAt := now.AddDate(0, 0, -data.DaysAgo)
		publishedAt := createdAt.Add(3 * time.Hour)
		cover := getSeedAsset("story-community-support.webp", "images")
		if cover == "" {
			return fmt.Errorf("story cover image is missing")
		}

		var story model.InspiringStory
		findResult := db.Where("title = ? AND author_id IN ?", data.Title, memberIDs).First(&story)
		if findResult.Error != nil && !errors.Is(findResult.Error, gorm.ErrRecordNotFound) {
			return findResult.Error
		}
		if errors.Is(findResult.Error, gorm.ErrRecordNotFound) {
			story = model.InspiringStory{
				AuthorID: author.ID, Title: data.Title, Content: data.Content,
				Status: model.StoryStatusApproved,
			}
			if err := db.Create(&story).Error; err != nil {
				return err
			}
		}

		if err := db.Model(&story).Updates(map[string]any{
			"author_id": author.ID, "content": data.Content, "cover_image": cover,
			"is_anonymous": true, "has_trigger_warning": data.HasWarning,
			"trigger_warning_text": data.WarningText, "status": model.StoryStatusApproved,
			"moderator_id": moderatorID, "moderation_feedback": "Disetujui sebagai ilustrasi editorial komposit.",
			"moderated_at": moderatedAt, "view_count": data.ViewCount,
			"is_featured": data.IsFeatured, "published_at": publishedAt,
			"created_at": createdAt, "updated_at": now,
		}).Error; err != nil {
			return err
		}

		for _, slug := range data.CategorySlugs {
			category, ok := categoryMap[slug]
			if !ok {
				continue
			}
			relation := model.StoryCategoryRelation{StoryID: story.ID, CategoryID: category.ID}
			if err := db.FirstOrCreate(&relation, relation).Error; err != nil {
				return err
			}
		}
		for _, tag := range data.Tags {
			storyTag := model.StoryTag{StoryID: story.ID, Tag: tag}
			if err := db.FirstOrCreate(&storyTag, "story_id = ? AND tag = ?", story.ID, tag).Error; err != nil {
				return err
			}
		}
		for _, authorIdx := range data.HeartUsers {
			userID := memberIDs[authorIdx%len(memberIDs)]
			if userID == author.ID {
				continue
			}
			heart := model.StoryHeart{StoryID: story.ID, UserID: userID}
			if err := db.FirstOrCreate(&heart, "story_id = ? AND user_id = ?", story.ID, userID).Error; err != nil {
				return err
			}
		}

		for _, seed := range data.Comments {
			commenterID := memberIDs[seed.AuthorIdx%len(memberIDs)]
			var comment model.StoryComment
			err := db.Where("story_id = ? AND user_id = ? AND content = ?", story.ID, commenterID, seed.Content).First(&comment).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				comment = model.StoryComment{
					StoryID: story.ID, UserID: commenterID, Content: seed.Content,
					CreatedAt: createdAt.Add(4 * time.Hour), UpdatedAt: createdAt.Add(4 * time.Hour),
				}
				if err := db.Create(&comment).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}

			heartUserID := memberIDs[seed.HeartIdx%len(memberIDs)]
			if heartUserID != commenterID {
				commentHeart := model.StoryCommentHeart{CommentID: comment.ID, UserID: heartUserID}
				if err := db.FirstOrCreate(&commentHeart, "comment_id = ? AND user_id = ?", comment.ID, heartUserID).Error; err != nil {
					return err
				}
			}
			var commentHearts int64
			if err := db.Model(&model.StoryCommentHeart{}).Where("comment_id = ?", comment.ID).Count(&commentHearts).Error; err != nil {
				return err
			}
			if err := db.Model(&comment).Update("heart_count", commentHearts).Error; err != nil {
				return err
			}
		}

		var heartCount, commentCount int64
		if err := db.Model(&model.StoryHeart{}).Where("story_id = ?", story.ID).Count(&heartCount).Error; err != nil {
			return err
		}
		if err := db.Model(&model.StoryComment{}).Where("story_id = ? AND is_hidden = ?", story.ID, false).Count(&commentCount).Error; err != nil {
			return err
		}
		if err := db.Model(&story).Updates(map[string]any{"heart_count": heartCount, "comment_count": commentCount}).Error; err != nil {
			return err
		}
	}

	return nil
}
