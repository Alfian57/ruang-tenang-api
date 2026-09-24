package presentation

import (
	"errors"
	"fmt"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedForums adds seeded discussions only for presentation accounts. Each
// discussion has topic-specific replies, likes and votes instead of randomly
// selected generic responses.
func SeedForums(db *gorm.DB) error {
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
	if len(members) == 0 {
		return nil
	}
	var categories []model.ForumCategory
	if err := db.Find(&categories).Error; err != nil {
		return err
	}
	categoryByName := make(map[string]uint, len(categories))
	for _, category := range categories {
		categoryByName[category.Name] = category.ID
	}

	type replySeed struct {
		AuthorIdx    int
		Content      string
		UpvotedBy    []int
		DownvotedBy  []int
		Accepted     bool
		CommunityFav bool
	}
	type threadSeed struct {
		Title     string
		Content   string
		Category  string
		AuthorIdx int
		DaysAgo   int
		LikedBy   []int
		Replies   []replySeed
	}

	threads := []threadSeed{
		{
			Title:    "Membagi beban kerja tanpa langsung mengambil keputusan besar",
			Content:  "Dalam beberapa minggu ini daftar pekerjaan saya makin panjang dan saya merasa terus tertinggal. Saya belum ingin mengambil keputusan besar saat sedang sangat lelah. Apa langkah kecil yang pernah membantu kalian memetakan beban dan membicarakan prioritas dengan atasan?",
			Category: "Kesehatan Mental di Tempat Kerja", AuthorIdx: 0, DaysAgo: 2, LikedBy: []int{1, 2},
			Replies: []replySeed{
				{AuthorIdx: 1, Content: "Saya pernah menulis semua tugas, lalu menandai mana yang benar-benar harus selesai hari ini. Daftar itu saya bawa saat bicara dengan atasan: mana yang perlu diprioritaskan, mana yang bisa mundur. Percakapannya tetap canggung, tapi lebih jelas daripada mencoba menanggung semuanya sendirian.", UpvotedBy: []int{0, 2}, CommunityFav: true},
				{AuthorIdx: 2, Content: "Kalau memungkinkan, coba catat jam kerja dan jeda selama beberapa hari. Bukan untuk menyalahkan diri, melainkan supaya pembicaraan tentang kapasitas punya contoh yang konkret. Semoga ada orang tepercaya yang bisa menemani menyiapkan obrolannya.", UpvotedBy: []int{0}},
			},
		},
		{
			Title:    "Apa yang kalian lakukan sebelum rapat saat rasa cemas naik?",
			Content:  "Presentasi singkat di rapat membuat saya memikirkan banyak kemungkinan buruk. Saya sedang mencari cara yang realistis supaya tetap bisa ikut rapat tanpa menuntut diri harus merasa tenang dulu. Apa persiapan kecil yang membantu kalian?",
			Category: "Tips Mengelola Stres", AuthorIdx: 1, DaysAgo: 4, LikedBy: []int{0, 2},
			Replies: []replySeed{
				{AuthorIdx: 2, Content: "Saya menyiapkan tiga poin yang wajib disampaikan dan menaruh catatan itu di dekat laptop. Sebelum mulai, saya melihat sekeliling dan mengingatkan diri bahwa saya sedang berada di ruang rapat, bukan di skenario yang saya bayangkan. Kalau latihan napas membuatmu tidak nyaman, tidak perlu dipaksakan.", UpvotedBy: []int{0}, Accepted: true},
				{AuthorIdx: 0, Content: "Saya minta rekan kerja mengirim agenda lebih awal, lalu berlatih satu kali dengan teman. Ternyata persiapan yang jelas lebih membantu daripada mengulang seluruh presentasi sampai larut malam.", UpvotedBy: []int{1, 2}},
			},
		},
		{
			Title:    "Bagaimana menemani teman yang mulai menarik diri?",
			Content:  "Seorang teman terlihat lebih jarang membalas pesan belakangan ini. Saya ingin hadir tanpa mendesaknya untuk cerita atau membuat asumsi tentang keadaannya. Bagaimana kalian biasanya membuka percakapan dengan lembut?",
			Category: "Dukungan Emosional", AuthorIdx: 2, DaysAgo: 6, LikedBy: []int{0, 1},
			Replies: []replySeed{
				{AuthorIdx: 0, Content: "Saya biasanya mengirim pesan yang tidak menuntut jawaban cepat, misalnya: ‘Aku ingat kamu. Tidak harus membalas sekarang; kalau mau ditemani jalan atau butuh bantuan praktis, kabari ya.’ Lalu saya menjaga konsistensi tanpa membanjiri pesannya.", UpvotedBy: []int{1, 2}, CommunityFav: true},
				{AuthorIdx: 1, Content: "Boleh juga bertanya langsung, ‘Kamu ingin didengarkan, ditemani, atau dibantu mencari solusi?’ Kalau ada kekhawatiran soal keselamatannya, lebih baik hubungi orang tepercaya yang dekat dengannya atau layanan darurat setempat.", UpvotedBy: []int{0}},
			},
		},
		{
			Title:    "Prompt jurnal yang membantu saya memisahkan fakta dan asumsi",
			Content:  "Saya sering menulis panjang lalu makin sulit memilah apa yang sebenarnya terjadi dan apa yang saya takutkan akan terjadi. Ada prompt singkat yang membantu refleksi tanpa menghakimi diri sendiri?",
			Category: "Diskusi Umum", AuthorIdx: 0, DaysAgo: 8, LikedBy: []int{1},
			Replies: []replySeed{
				{AuthorIdx: 2, Content: "Coba tiga kolom: ‘yang saya tahu sebagai fakta’, ‘yang sedang saya perkirakan’, dan ‘satu hal yang bisa saya lakukan atau tanyakan’. Tidak harus selesai dalam satu sesi; menulis ‘belum tahu’ juga boleh.", UpvotedBy: []int{0, 1}, Accepted: true},
				{AuthorIdx: 1, Content: "Saya menutup jurnal dengan satu pertanyaan: ‘Apa yang akan saya katakan kepada teman yang mengalami hal serupa?’ Bukan untuk mengecilkan masalah, tetapi supaya suara di kepala tidak selalu keras.", UpvotedBy: []int{0}},
			},
		},
		{
			Title:    "Membangun rutinitas malam saat jam tidur berantakan",
			Content:  "Jadwal tidur saya berubah-ubah karena kuliah dan kerja. Saya ingin mencoba rutinitas yang sederhana, bukan daftar aturan panjang yang sulit dijalankan. Perubahan kecil apa yang paling masuk akal untuk dimulai?",
			Category: "Kesehatan Mental di Sekolah", AuthorIdx: 1, DaysAgo: 10, LikedBy: []int{0, 2},
			Replies: []replySeed{
				{AuthorIdx: 0, Content: "Saya mulai dari jam bangun yang lebih konsisten, lalu menyiapkan barang untuk pagi sebelum masuk tempat tidur. Tidak selalu berhasil, jadi saya menganggapnya eksperimen, bukan ujian disiplin.", UpvotedBy: []int{2}},
				{AuthorIdx: 2, Content: "Rutinitas yang sama selama 10 menit terasa lebih mungkin buat saya: redupkan lampu, siapkan air, lalu baca beberapa halaman. Kalau gangguan tidur terus memengaruhi aktivitas, ngobrol dengan tenaga kesehatan bisa membantu mencari penyebabnya.", UpvotedBy: []int{0, 1}, Accepted: true},
			},
		},
		{
			Title:    "Hal apa yang perlu ditanyakan saat mencari psikolog?",
			Content:  "Saya sedang mempertimbangkan konsultasi pertama dan ingin tahu cara memilih layanan yang sesuai. Selain biaya dan jadwal, apa yang sebaiknya ditanyakan sebelum membuat janji?",
			Category: "Pertanyaan & Jawaban", AuthorIdx: 2, DaysAgo: 12, LikedBy: []int{0, 1},
			Replies: []replySeed{
				{AuthorIdx: 1, Content: "Kamu bisa bertanya soal latar belakang dan pendekatan yang digunakan, alur sesi pertama, kerahasiaan, biaya pembatalan, serta pilihan jika setelah beberapa sesi ternyata kurang cocok. Tidak perlu menyiapkan cerita yang sempurna.", UpvotedBy: []int{0, 2}, Accepted: true},
				{AuthorIdx: 0, Content: "Saya mencatat dua atau tiga hal yang paling ingin dibahas, tapi konselor juga membantu mengarahkan percakapan. Kalau bingung mulai dari mana, mengatakan ‘saya belum tahu harus mulai dari mana’ pun cukup.", UpvotedBy: []int{2}},
			},
		},
	}

	legacyReplies := []string{
		"Semangat ya! Kamu nggak sendirian. Kalau butuh apa-apa, feel free to reach out 💙",
		"Terima kasih sudah berbagi. Ini sangat menginspirasi!",
		"Saya juga pernah mengalami hal serupa. Yang penting jangan menyerah dan terus mencoba.",
		"Coba teknik pomodoro untuk mengatasi prokrastinasi. Kerja 25 menit, istirahat 5 menit.",
		"Jangan lupa untuk istirahat juga ya. Kesehatan mental itu penting banget.",
		"I feel you. Semoga lekas membaik! 🤗",
		"Pengalaman yang sangat berharga. Terima kasih sudah mau berbagi.",
		"Semangat! Satu langkah kecil setiap hari itu sudah progress yang bagus.",
	}
	memberIDs := make([]uint, len(members))
	for i, user := range members {
		memberIDs[i] = user.ID
	}
	legacyTitles := []string{
		"Bagaimana cara kalian mengatasi burnout?", "Cerita sukses sembuh dari anxiety",
		"Butuh teman curhat", "Rekomendasi buku self-improvement",
		"Tips menghadapi social anxiety di tempat kerja", "Apakah terapi worth it?",
		"Meditasi pagi bikin produktif!", "Stress karena deadline skripsi",
	}
	if err := db.Where("title IN ? AND user_id IN ?", legacyTitles, memberIDs).Delete(&model.Forum{}).Error; err != nil {
		return err
	}
	now := time.Now().UTC()

	for _, data := range threads {
		author := members[data.AuthorIdx%len(members)]
		categoryID, found := categoryByName[data.Category]
		if !found {
			return fmt.Errorf("forum category %q must be seeded before forums", data.Category)
		}
		createdAt := now.AddDate(0, 0, -data.DaysAgo)
		forum := model.Forum{
			UserID: author.ID, CategoryID: &categoryID, Title: data.Title,
			Content: data.Content, CreatedAt: createdAt, UpdatedAt: createdAt,
			HasAcceptedAnswer: false,
		}
		var existing model.Forum
		findResult := db.Where("title = ? AND user_id IN ?", data.Title, memberIDs).First(&existing)
		if findResult.Error == nil {
			forum.ID = existing.ID
			if err := db.Model(&existing).Updates(map[string]any{
				"user_id": author.ID, "category_id": categoryID, "content": data.Content,
				"created_at": createdAt, "updated_at": createdAt,
			}).Error; err != nil {
				return err
			}
		} else if errors.Is(findResult.Error, gorm.ErrRecordNotFound) {
			if err := db.Create(&forum).Error; err != nil {
				return err
			}
			existing = forum
		} else {
			return findResult.Error
		}

		// Replace only canned replies and reactions from known demo accounts that
		// the previous seeder generated. Real community replies are preserved.
		var oldPosts []model.ForumPost
		if err := db.Where("forum_id = ? AND user_id IN ? AND content IN ?", existing.ID, memberIDs, legacyReplies).Find(&oldPosts).Error; err != nil {
			return err
		}
		for _, post := range oldPosts {
			if err := db.Where("post_id = ?", post.ID).Delete(&model.ForumPostVote{}).Error; err != nil {
				return err
			}
			if err := db.Unscoped().Delete(&post).Error; err != nil {
				return err
			}
		}
		if err := db.Where("forum_id = ? AND user_id IN ?", existing.ID, memberIDs).Delete(&model.ForumLike{}).Error; err != nil {
			return err
		}

		for _, likerIdx := range data.LikedBy {
			userID := memberIDs[likerIdx%len(memberIDs)]
			if userID == author.ID {
				continue
			}
			like := model.ForumLike{ForumID: existing.ID, UserID: userID}
			if err := db.FirstOrCreate(&like, "forum_id = ? AND user_id = ?", existing.ID, userID).Error; err != nil {
				return err
			}
		}

		accepted := false
		for replyIndex, seed := range data.Replies {
			replyUserID := memberIDs[seed.AuthorIdx%len(memberIDs)]
			var post model.ForumPost
			err := db.Where("forum_id = ? AND user_id = ? AND content = ?", existing.ID, replyUserID, seed.Content).First(&post).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				post = model.ForumPost{
					ForumID: existing.ID, UserID: replyUserID, Content: seed.Content,
					CreatedAt: createdAt.Add(time.Duration(replyIndex+2) * time.Hour),
					UpdatedAt: createdAt.Add(time.Duration(replyIndex+2) * time.Hour),
				}
				if err := db.Create(&post).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			}

			if err := db.Model(&post).Updates(map[string]any{
				"is_accepted_answer":    seed.Accepted,
				"is_community_favorite": seed.CommunityFav,
			}).Error; err != nil {
				return err
			}
			accepted = accepted || seed.Accepted

			for _, voterIdx := range seed.UpvotedBy {
				voterID := memberIDs[voterIdx%len(memberIDs)]
				if voterID == replyUserID {
					continue
				}
				vote := model.ForumPostVote{PostID: post.ID, UserID: voterID, VoteType: model.VoteTypeUpvote}
				if err := db.Where("post_id = ? AND user_id = ?", post.ID, voterID).Assign("vote_type", model.VoteTypeUpvote).FirstOrCreate(&vote).Error; err != nil {
					return err
				}
			}
			for _, voterIdx := range seed.DownvotedBy {
				voterID := memberIDs[voterIdx%len(memberIDs)]
				if voterID == replyUserID {
					continue
				}
				vote := model.ForumPostVote{PostID: post.ID, UserID: voterID, VoteType: model.VoteTypeDownvote}
				if err := db.Where("post_id = ? AND user_id = ?", post.ID, voterID).Assign("vote_type", model.VoteTypeDownvote).FirstOrCreate(&vote).Error; err != nil {
					return err
				}
			}

			var upvotes, downvotes int64
			if err := db.Model(&model.ForumPostVote{}).Where("post_id = ? AND vote_type = ?", post.ID, model.VoteTypeUpvote).Count(&upvotes).Error; err != nil {
				return err
			}
			if err := db.Model(&model.ForumPostVote{}).Where("post_id = ? AND vote_type = ?", post.ID, model.VoteTypeDownvote).Count(&downvotes).Error; err != nil {
				return err
			}
			if err := db.Model(&post).Updates(map[string]any{"upvotes_count": upvotes, "downvotes_count": downvotes}).Error; err != nil {
				return err
			}
		}
		if err := db.Model(&existing).Update("has_accepted_answer", accepted).Error; err != nil {
			return err
		}
	}

	return nil
}
