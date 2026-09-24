package presentation

import (
	"errors"
	"fmt"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedArticles publishes a compact editorial library with practical steps,
// safety notes, and links to primary public-health sources.
func SeedArticles(db *gorm.DB) error {
	var admin model.User
	if err := db.Where("email = ? AND role = ?", presentationAdminEmail, model.RoleAdmin).First(&admin).Error; err != nil {
		return err
	}
	var categoryIDs = make(map[string]uint)
	for _, name := range []string{"Kesehatan Mental", "Tips & Trik", "Meditasi", "Mindfulness", "Self-Care", "Hubungan"} {
		var category model.ArticleCategory
		if err := db.Where("name = ?", name).First(&category).Error; err != nil {
			return fmt.Errorf("article category %q must be seeded before articles: %w", name, err)
		}
		categoryIDs[name] = category.ID
	}
	legacyTitles := []string{
		"Mengenal Kecemasan dan Cara Mengatasinya", "5 Teknik Pernapasan untuk Menenangkan Pikiran",
		"Panduan Meditasi untuk Pemula", "Mengatasi Stres di Tempat Kerja", "Pentingnya Tidur untuk Kesehatan Mental",
		"Pengalaman Saya Bangkit dari Burnout", "Tips Mengatur Waktu Belajar Anti Cemas",
	}
	seedAccounts, err := presentationUsers(db)
	if err != nil {
		return err
	}
	seedOwnerIDs := []uint{admin.ID}
	for _, user := range seedAccounts {
		seedOwnerIDs = append(seedOwnerIDs, user.ID)
	}
	if err := db.Where("title IN ? AND user_id IN ?", legacyTitles, seedOwnerIDs).Delete(&model.Article{}).Error; err != nil {
		return err
	}

	type articleSeed struct {
		Title    string
		Content  string
		Category string
		Image    string
	}
	articles := []articleSeed{
		{
			Title:    "Ketika cemas terasa penuh: grounding yang fleksibel",
			Category: "Kesehatan Mental", Image: "article-calm-start.webp",
			Content: `<p>Cemas dapat hadir sebagai pikiran yang bergerak cepat, tubuh tegang, atau dorongan untuk menghindari sesuatu. Pengalaman tiap orang berbeda. Latihan grounding tidak menghilangkan sumber masalah, tetapi bisa menjadi cara untuk kembali memperhatikan keadaan di sekitar sebelum memilih langkah berikutnya.</p>
<h2>Latihan singkat dengan mata terbuka</h2>
<ol>
  <li>Jika terasa aman, duduk atau berdiri dengan posisi yang nyaman. Tidak perlu mengubah napas.</li>
  <li>Perhatikan telapak kaki yang menyentuh lantai atau kursi yang menopang tubuh.</li>
  <li>Sebutkan perlahan tiga benda yang kamu lihat dan satu suara yang terdengar.</li>
  <li>Tanyakan pada diri sendiri: “Apa satu hal yang perlu saya lakukan sekarang?” Pilih langkah kecil atau hubungi orang yang kamu percaya.</li>
</ol>
<p>Latihan ini bukan ujian. Kamu boleh berhenti, menjaga mata tetap terbuka, atau memilih kegiatan lain bila memperhatikan tubuh dan napas terasa tidak nyaman. Bila kecemasan berulang kali mengganggu kegiatan, tidur, atau hubungan, pertimbangkan berbicara dengan tenaga kesehatan.</p>
<h2>Bacaan lanjutan</h2><p><a href="https://www.who.int/publications/i/item/9789240003927" target="_blank" rel="noopener noreferrer">Panduan WHO: Doing What Matters in Times of Stress</a> membahas grounding, mengambil jarak dari pikiran, dan bertindak sesuai nilai.</p>
<p><small>Artikel ini untuk edukasi umum, bukan diagnosis atau pengganti layanan profesional.</small></p>`,
		},
		{
			Title:    "Bernapas pelan tanpa memaksa diri",
			Category: "Tips & Trik", Image: "article-pause.webp",
			Content: `<p>Latihan napas sering disarankan saat seseorang sedang tegang, tetapi hitungan tertentu tidak cocok untuk semua orang. Kamu tidak perlu menahan napas atau menarik napas sedalam mungkin. Tujuan latihan singkat ini hanya mengundang ritme yang sedikit lebih lambat dan tetap nyaman.</p>
<h2>Coba versi yang sederhana</h2>
<ol>
  <li>Pilih posisi yang terasa aman; mata boleh tetap terbuka.</li>
  <li>Perhatikan napas sebagaimana adanya selama beberapa saat.</li>
  <li>Jika nyaman, biarkan hembusan sedikit lebih panjang tanpa menghitung atau menahan napas.</li>
  <li>Berhenti setelah beberapa putaran, lalu perhatikan apakah kamu ingin melanjutkan atau melakukan hal lain.</li>
</ol>
<p>Jika terasa pusing, sesak, panik, atau tidak nyaman, kembali bernapas seperti biasa dan hentikan latihan. Kamu juga bisa memilih grounding melalui benda yang terlihat atau menghubungi seseorang. Latihan relaksasi adalah salah satu pilihan, bukan satu-satunya jalan menghadapi stres.</p>
<h2>Bacaan lanjutan</h2><p><a href="https://www.emro.who.int/mhps/dealing_with_stress.html" target="_blank" rel="noopener noreferrer">Materi WHO tentang menghadapi stres</a> menyarankan napas yang pelan dan lembut, serta mengingatkan untuk tidak bernapas terlalu cepat atau terlalu dalam.</p>
<p><small>Artikel ini untuk edukasi umum dan tidak menggantikan saran tenaga kesehatan.</small></p>`,
		},
		{
			Title:    "Meditasi untuk pemula: kembali, bukan mengosongkan pikiran",
			Category: "Meditasi", Image: "article-calm-start.webp",
			Content: `<p>Dalam latihan perhatian, pikiran tetap bisa muncul. Tujuannya bukan membuat kepala kosong, melainkan menyadari bahwa perhatian sedang berpindah lalu memilih apakah ingin kembali ke hal yang sedang dilakukan.</p>
<h2>Latihan tiga menit yang bisa diubah</h2>
<ol>
  <li>Duduk, berdiri, atau berjalan pelan dengan cara yang nyaman.</li>
  <li>Pilih satu jangkar perhatian: suara sekitar, warna benda, atau sensasi tangan menyentuh meja.</li>
  <li>Saat sadar pikiran mengembara, beri nama singkat seperti “sedang merencanakan” atau “sedang mengingat”.</li>
  <li>Kembali memperhatikan jangkar itu jika terasa membantu. Tidak perlu mengkritik diri karena terdistraksi.</li>
</ol>
<p>Mulailah sebentar dan sesuaikan durasinya. Jika memejamkan mata, berfokus pada tubuh, atau duduk diam membuatmu tidak nyaman, buka mata, bergerak, atau hentikan latihan. Respons terhadap meditasi berbeda-beda.</p>
<p><a href="https://www.nccih.nih.gov/health/meditation-and-mindfulness-effectiveness-and-safety" target="_blank" rel="noopener noreferrer">NCCIH menjelaskan manfaat dan keamanan mindfulness</a>, termasuk bahwa meditasi tidak seharusnya menggantikan perawatan konvensional atau menunda konsultasi untuk masalah kesehatan.</p>
<p><small>Artikel ini adalah informasi umum, bukan terapi atau diagnosis.</small></p>`,
		},
		{
			Title:    "Mengatur stres kerja lewat percakapan prioritas",
			Category: "Tips & Trik", Image: "article-calm-start.webp",
			Content: `<p>Saat beberapa tenggat bertabrakan, menyuruh diri sendiri bekerja lebih cepat belum tentu menyelesaikan persoalan kapasitas. Sebagian stres berkaitan dengan tuntutan, kejelasan peran, dukungan tim, dan kendali atas cara kerja. Karena itu, mengatur beban bukan tanggung jawab individu saja.</p>
<h2>Persiapan sebelum berbicara</h2>
<ul>
  <li>Tuliskan tugas yang berjalan, tenggat, dan perkiraan waktu yang diperlukan.</li>
  <li>Tandai bagian yang saling bertabrakan atau membutuhkan keputusan orang lain.</li>
  <li>Pilih permintaan yang konkret: menentukan urutan, mengubah tenggat, atau mencari dukungan.</li>
</ul>
<p>Contoh pembuka: “Saya sedang mengerjakan A dan B dengan tenggat yang berdekatan. Mana yang perlu saya prioritaskan lebih dulu, dan apakah tenggat yang lain bisa disesuaikan?” Kamu dapat meminta percakapan lanjutan atau membawa pendamping bila itu membantu.</p>
<p>Jeda, rutinitas, dan dukungan sosial dapat menjadi bagian dari pengelolaan stres, tetapi tidak memperbaiki lingkungan kerja yang tidak aman dengan sendirinya. Jika stres menetap atau mengganggu fungsi sehari-hari, bicarakan dengan tenaga kesehatan atau layanan dukungan yang tersedia.</p>
<p><a href="https://www.who.int/news-room/questions-and-answers/item/stress" target="_blank" rel="noopener noreferrer">WHO: Stress</a> menjelaskan respons stres dan beberapa pilihan dukungan sehari-hari.</p>
<p><small>Informasi umum; bukan penilaian kondisi kerja atau saran klinis.</small></p>`,
		},
		{
			Title:    "Membangun rutinitas tidur yang realistis",
			Category: "Kesehatan Mental", Image: "article-sleep-routine.webp",
			Content: `<p>Rutinitas tidur tidak harus sempurna untuk terasa membantu. Mulailah dari satu perubahan kecil yang sesuai dengan jadwal dan kondisi tempat tinggalmu. Orang dewasa usia 18–60 tahun umumnya disarankan tidur sedikitnya tujuh jam per malam; kebutuhan dapat berbeda menurut usia dan individu.</p>
<h2>Pilih satu langkah untuk dicoba</h2>
<ul>
  <li>Usahakan waktu bangun yang cukup konsisten pada sebagian besar hari.</li>
  <li>Buat kamar terasa tenang, nyaman, dan sejuk sejauh keadaan memungkinkan.</li>
  <li>Matikan perangkat elektronik setidaknya 30 menit sebelum tidur jika itu realistis bagimu.</li>
  <li>Perhatikan apakah kafein pada sore atau malam hari memengaruhi tidurmu.</li>
</ul>
<p>Catat pola tidur dan hal yang mungkin memengaruhinya tanpa menyalahkan diri. Bila masalah tidur menetap atau mengganggu aktivitas di siang hari, bicarakan dengan penyedia layanan kesehatan untuk menilai penyebab dan pilihan bantuan. Jangan mengubah atau menghentikan obat tanpa berkonsultasi.</p>
<p><a href="https://www.cdc.gov/sleep/about/" target="_blank" rel="noopener noreferrer">CDC: About Sleep</a> merangkum kebutuhan tidur menurut kelompok usia dan kebiasaan yang dapat mendukung tidur.</p>
<p><small>Informasi ini tidak mendiagnosis gangguan tidur.</small></p>`,
		},
		{
			Title:    "Menjadi teman yang mendengarkan tanpa memaksakan solusi",
			Category: "Hubungan", Image: "story-community-support.webp",
			Content: `<p>Ketika seseorang bercerita bahwa ia sedang kesulitan, kita mungkin ingin segera memperbaiki keadaan. Namun orang tersebut bisa jadi lebih membutuhkan ruang untuk didengarkan, bantuan praktis, atau ditemani mencari dukungan. Bertanya lebih dulu membantu kita tidak berasumsi.</p>
<h2>Kalimat pembuka yang memberi pilihan</h2>
<ul>
  <li>“Terima kasih sudah cerita. Kamu ingin aku mendengarkan dulu, membantu mencari langkah, atau menemani menghubungi seseorang?”</li>
  <li>“Tidak perlu menjawab sekarang. Aku bisa menghubungimu lagi besok.”</li>
  <li>“Ada hal praktis yang bisa kubantu hari ini?”</li>
</ul>
<p>Dengarkan tanpa membandingkan pengalaman atau menjanjikan bahwa semuanya pasti cepat membaik. Jaga batas kemampuanmu sendiri juga; kamu boleh membantu mencari orang tepercaya atau layanan profesional. Jika ada risiko keselamatan yang mendesak, libatkan bantuan langsung dan layanan darurat setempat.</p>
<p>Ruang Tenang adalah tempat berbagi dukungan, bukan layanan darurat atau pengganti konseling profesional.</p>`,
		},
		{
			Title:    "Menulis jurnal tanpa harus membagikannya",
			Category: "Self-Care", Image: "article-pause.webp",
			Content: `<p>Jurnal dapat menjadi ruang untuk merapikan pengalaman, mencatat hal yang ingin dibicarakan, atau sekadar menulis tanpa kesimpulan. Tidak ada format yang wajib. Kamu tidak harus membagikan tulisan pribadi agar proses refleksi tetap berguna.</p>
<h2>Prompt singkat untuk memulai</h2>
<ol>
  <li>Apa yang terjadi hari ini, dengan kata-kata yang paling sederhana?</li>
  <li>Apa yang saya rasakan atau butuhkan saat itu?</li>
  <li>Apa satu hal yang ingin saya ingat, tanyakan, atau coba besok?</li>
</ol>
<p>Kamu boleh menulis beberapa kalimat lalu berhenti. Jika menulis membuat perasaan semakin berat, tutup jurnal dan pilih dukungan lain. Sebelum berbagi entri dengan orang lain atau fitur digital, tinjau pengaturan privasi dan izin yang diminta. Jurnal bukan rekam medis dan tidak dimaksudkan untuk mendiagnosis diri.</p>
<p><a href="https://www.who.int/publications/i/item/9789240003927" target="_blank" rel="noopener noreferrer">Panduan pengelolaan stres WHO</a> memuat latihan refleksi dan tindakan kecil yang dapat dicoba sesuai kebutuhan.</p>
<p><small>Informasi umum untuk refleksi pribadi, bukan terapi.</small></p>`,
		},
		{
			Title:    "Mencari bantuan saat beban terasa tidak aman",
			Category: "Kesehatan Mental", Image: "story-community-support.webp",
			Content: `<p>Kamu tidak perlu menunggu sampai menemukan kata-kata yang sempurna untuk mencari bantuan. Kamu bisa menghubungi orang tepercaya, fasilitas kesehatan, psikolog, psikiater, atau layanan dukungan di daerahmu dan menyampaikan bahwa kamu sedang merasa tidak aman.</p>
<p>Di Indonesia, FAQ Kementerian Kesehatan tentang Healing119.id menjelaskan bahwa layanan ini memberi dukungan emosional dan konsultasi dasar, serta dapat menghubungkan pengguna ke layanan lanjutan. FAQ tersebut mencantumkan akses panggilan melalui 119 ekstensi 8 dan chat melalui situs <a href="https://www.healing119.id/" target="_blank" rel="noopener noreferrer">Healing119.id</a>. Layanan ini bukan terapi jangka panjang; lihat situs resmi untuk informasi terbaru dan ketersediaan.</p>
<p>Jika ada bahaya langsung, jangan menunggu balasan aplikasi: hubungi layanan darurat setempat atau pergi ke fasilitas kesehatan terdekat, dan minta seseorang yang kamu percaya untuk menemani.</p>
<p><small>Ruang Tenang bukan layanan krisis. Informasi kontak ditinjau dari <a href="https://kesprimkom.kemkes.go.id/assets/uploads/contents/others/FAQ_Cegah_Bunuh_Diri%2C_Dukung_Kesehatan_Jiwa__Kenali_Layanan_Healing119.id.pdf" target="_blank" rel="noopener noreferrer">FAQ resmi Kementerian Kesehatan tentang Healing119.id</a> pada September 2026.</small></p>`,
		},
	}

	for _, seed := range articles {
		categoryID := categoryIDs[seed.Category]
		thumbnail := getSeedAsset(seed.Image, "images")
		if thumbnail == "" {
			return fmt.Errorf("thumbnail not found for article %q (%s)", seed.Title, seed.Image)
		}

		var existing model.Article
		findResult := db.Where("title = ? AND user_id = ?", seed.Title, admin.ID).First(&existing)
		if findResult.Error == nil {
			if err := db.Model(&existing).Updates(map[string]any{
				"content": seed.Content, "article_category_id": categoryID,
				"thumbnail": thumbnail, "status": model.ArticleStatusPublished,
				"user_id": admin.ID, "is_user_generated": false,
			}).Error; err != nil {
				return err
			}
			continue
		}
		if !errors.Is(findResult.Error, gorm.ErrRecordNotFound) {
			return findResult.Error
		}
		article := model.Article{
			Title: seed.Title, Thumbnail: thumbnail, Content: seed.Content,
			ArticleCategoryID: categoryID, UserID: admin.ID,
			Status: model.ArticleStatusPublished,
		}
		if err := db.Create(&article).Error; err != nil {
			return err
		}
	}

	// Retire the old rejected fixture because it contained a false medical claim.
	if err := db.Where("title = ? AND user_id IN ?", "Obat Herbal Ajaib yang Pasti Menyembuhkan Depresi", seedOwnerIDs).Delete(&model.Article{}).Error; err != nil {
		return err
	}
	return seedUserGeneratedArticles(db, seedAccounts, categoryIDs["Kesehatan Mental"], categoryIDs["Tips & Trik"])
}

// seedUserGeneratedArticles supplies safe draft examples for the moderation
// queue. Drafts are attached only to a fixed presentation account.
func seedUserGeneratedArticles(db *gorm.DB, seedAccounts []model.User, healthCategoryID, tipsCategoryID uint) error {
	var author model.User
	if err := db.Where("role = ? AND email IN ?", model.RoleUser, presentationAccountEmails).Order("id ASC").First(&author).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	seedOwnerIDs := make([]uint, 0, len(seedAccounts))
	for _, user := range seedAccounts {
		seedOwnerIDs = append(seedOwnerIDs, user.ID)
	}

	entries := []struct {
		Title            string
		Content          string
		CategoryID       uint
		ModerationStatus model.ArticleModerationStatus
		ModerationNotes  string
	}{
		{
			Title:      "Contoh kiriman: refleksi setelah hari yang padat",
			Content:    "<p>Contoh kiriman akun presentasi yang masih berupa draf. Catatan ini mencoba membedakan hal yang perlu dikerjakan hari ini dan hal yang dapat menunggu. Sebelum diterbitkan, penulis dapat menambahkan konteks dan sumber untuk saran yang diberikan.</p>",
			CategoryID: healthCategoryID, ModerationStatus: model.ArticleModerationPending,
		},
		{
			Title:      "Contoh kiriman: membuat rencana belajar lebih ringan",
			Content:    "<p>Contoh kiriman akun presentasi untuk ditinjau moderator. Naskah menyarankan pemecahan tugas menjadi langkah kecil, tetapi perlu menyebut bahwa cara tersebut tidak selalu sesuai untuk semua orang.</p>",
			CategoryID: tipsCategoryID, ModerationStatus: model.ArticleModerationFlagged,
			ModerationNotes: "Contoh pemeriksaan moderasi: tinjau bahasa yang terdengar terlalu menjanjikan dan tambahkan konteks pengalaman pribadi.",
		},
		{
			Title:      "Contoh kiriman: sumber bacaan belum dicantumkan",
			Content:    "<p>Contoh draf yang ditolak untuk diterbitkan karena berisi rangkuman informasi kesehatan tanpa sumber yang dapat diperiksa. Penulis diminta menulis ulang dengan rujukan yang jelas sebelum mengirim kembali.</p>",
			CategoryID: healthCategoryID, ModerationStatus: model.ArticleModerationRejected,
			ModerationNotes: "Tidak ada sumber primer yang dapat diverifikasi; kiriman diminta ditulis ulang sebelum diterbitkan.",
		},
	}

	for _, seed := range entries {
		var existing model.Article
		findResult := db.Where("title = ? AND user_id IN ?", seed.Title, seedOwnerIDs).First(&existing)
		if findResult.Error == nil {
			if err := db.Model(&existing).Updates(map[string]any{
				"content": seed.Content, "article_category_id": seed.CategoryID,
				"user_id": author.ID, "status": model.ArticleStatusDraft,
				"moderation_status": seed.ModerationStatus, "moderation_notes": seed.ModerationNotes,
				"is_user_generated": true,
			}).Error; err != nil {
				return err
			}
			continue
		}
		if !errors.Is(findResult.Error, gorm.ErrRecordNotFound) {
			return findResult.Error
		}
		article := model.Article{
			Title: seed.Title, Content: seed.Content,
			ArticleCategoryID: seed.CategoryID, UserID: author.ID,
			Status: model.ArticleStatusDraft, ModerationStatus: seed.ModerationStatus,
			ModerationNotes: seed.ModerationNotes, IsUserGenerated: true,
		}
		if err := db.Create(&article).Error; err != nil {
			return err
		}
	}
	return nil
}
