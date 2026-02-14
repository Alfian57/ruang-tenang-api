package development

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedArticles seeds test articles for development
func SeedArticles(db *gorm.DB) error {
	// Get admin user
	var admin model.User
	if err := db.Where("email = ?", "admin@ruangtenang.id").First(&admin).Error; err != nil {
		if err := db.First(&admin).Error; err != nil {
			return err
		}
	}

	// Get categories
	var healthCategory, tipsCategory, meditasiCategory model.ArticleCategory
	db.Where("name = ?", "Kesehatan Mental").First(&healthCategory)
	db.Where("name = ?", "Tips & Trik").First(&tipsCategory)
	db.Where("name = ?", "Meditasi").First(&meditasiCategory)

	articles := []struct {
		Title      string
		Content    string
		CategoryID uint
		Image      string
	}{
		{
			Title:      "Mengenal Kecemasan dan Cara Mengatasinya",
			Content:    "Kecemasan adalah respons alami tubuh terhadap stres. Ini adalah perasaan takut atau khawatir tentang apa yang akan datang.\n\nCara Mengatasi Kecemasan:\n1. Latihan pernapasan dalam\n2. Meditasi teratur\n3. Olahraga rutin\n4. Tidur yang cukup\n5. Mengurangi kafein\n\nJika kecemasan mulai mengganggu aktivitas sehari-hari, pertimbangkan untuk berkonsultasi dengan profesional kesehatan mental.",
			CategoryID: healthCategory.ID,
			Image:      "article-mental.jpg",
		},
		{
			Title:      "5 Teknik Pernapasan untuk Menenangkan Pikiran",
			Content:    "Pernapasan yang tepat dapat membantu menenangkan sistem saraf.\n\n1. Teknik 4-7-8\nTarik napas selama 4 detik, tahan 7 detik, hembuskan 8 detik.\n\n2. Pernapasan Kotak\nTarik napas 4 detik, tahan 4 detik, hembuskan 4 detik, tahan 4 detik.\n\n3. Pernapasan Diafragma\nLetakkan tangan di perut, rasakan perut mengembang saat menarik napas.\n\n4. Pernapasan Bergantian\nTutup lubang hidung kanan, tarik napas dari kiri, lalu sebaliknya.\n\n5. Pernapasan Ujjayi\nTeknik yoga dengan suara seperti ombak.",
			CategoryID: tipsCategory.ID,
			Image:      "article-tips.jpg",
		},
		{
			Title:      "Panduan Meditasi untuk Pemula",
			Content:    "Meditasi tidak harus rumit. Mulailah dengan 5 menit sehari.\n\nLangkah-langkah:\n1. Duduk dengan nyaman di tempat yang tenang\n2. Tutup mata perlahan\n3. Fokus pada napas yang masuk dan keluar\n4. Biarkan pikiran mengalir tanpa menghakimi\n5. Jika pikiran menyimpang, kembalikan fokus ke napas\n\nManfaat meditasi:\n- Mengurangi stres\n- Meningkatkan konsentrasi\n- Memperbaiki kualitas tidur\n- Meningkatkan kesadaran diri",
			CategoryID: meditasiCategory.ID,
			Image:      "article-meditation.jpg",
		},
		{
			Title:      "Mengatasi Stres di Tempat Kerja",
			Content:    "Stres kerja adalah masalah umum yang dapat mempengaruhi kesehatan mental dan fisik.\n\nTips Mengatasi:\n1. Buat batasan yang jelas antara kerja dan kehidupan pribadi\n2. Ambil break secara teratur - coba teknik Pomodoro\n3. Prioritaskan tugas dengan baik menggunakan matrix Eisenhower\n4. Jangan takut untuk meminta bantuan\n5. Praktikkan teknik relaksasi di sela-sela kerja\n6. Jaga pola tidur yang teratur\n7. Komunikasikan beban kerja dengan atasan jika diperlukan",
			CategoryID: tipsCategory.ID,
			Image:      "article-stress.jpg",
		},
		{
			Title:      "Pentingnya Tidur untuk Kesehatan Mental",
			Content:    "Tidur yang cukup sangat penting untuk menjaga kesehatan mental dan kognitif.\n\nManfaat Tidur yang Cukup:\n1. Meningkatkan konsentrasi dan produktivitas\n2. Memperbaiki mood dan regulasi emosi\n3. Mengurangi risiko depresi dan kecemasan\n4. Meningkatkan daya ingat dan pembelajaran\n5. Membantu pemulihan fisik\n\nTips untuk tidur lebih baik:\n- Tetapkan jadwal tidur yang konsisten\n- Hindari layar 1 jam sebelum tidur\n- Ciptakan lingkungan tidur yang nyaman\n- Hindari kafein di sore/malam hari",
			CategoryID: healthCategory.ID,
			Image:      "article-sleep.jpg",
		},
	}

	for _, a := range articles {
		var existing model.Article
		if db.Where("title = ?", a.Title).First(&existing).RowsAffected > 0 {
			continue
		}

		// Get thumbnail
		thumbnail := ""
		if url, ok := placeholderImages[a.Image]; ok {
			thumbnail = getOrDownloadImage(url, a.Image)
		}

		article := model.Article{
			Title:             a.Title,
			Thumbnail:         thumbnail,
			Content:           a.Content,
			ArticleCategoryID: a.CategoryID,
			UserID:            admin.ID,
			Status:            model.ArticleStatusPublished,
		}

		if err := db.Create(&article).Error; err != nil {
			return err
		}
	}

	return nil
}
