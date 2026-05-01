package presentation

import (
	"errors"
	"os"
	"strings"

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
			Title: "Mengenal Kecemasan dan Cara Mengatasinya",
			Content: `<h2>Apa Itu Kecemasan?</h2>
<p>Kecemasan adalah respons alami tubuh saat kita merasa ada ancaman, tekanan, atau ketidakpastian. Dalam kadar ringan, kecemasan bisa membantu kita lebih waspada. Namun bila terjadi terlalu sering, terlalu kuat, atau bertahan lama, kecemasan dapat mengganggu fokus belajar, kualitas tidur, hubungan sosial, dan rasa percaya diri.</p>
<p>Gejalanya bisa muncul dalam bentuk jantung berdebar, napas pendek, sulit rileks, pikiran berputar tanpa henti, hingga dorongan untuk menghindari situasi tertentu. Penting dipahami bahwa mengalami kecemasan bukan tanda kelemahan, melainkan sinyal bahwa tubuh dan pikiran sedang membutuhkan dukungan.</p>

<h2>Tanda Kecemasan yang Perlu Diperhatikan</h2>
<ul>
  <li><strong>Fisik:</strong> otot tegang, pusing, gangguan pencernaan, sulit tidur.</li>
  <li><strong>Pikiran:</strong> overthinking, skenario terburuk, sulit mengambil keputusan.</li>
  <li><strong>Emosi:</strong> mudah panik, mudah tersinggung, merasa tidak aman.</li>
  <li><strong>Perilaku:</strong> menunda tugas, menarik diri, menghindari percakapan penting.</li>
</ul>

<h2>Strategi Praktis Mengelola Kecemasan</h2>
<p>Gunakan pendekatan bertahap. Pilih satu teknik sederhana, lakukan konsisten selama 7-14 hari, lalu evaluasi dampaknya:</p>
<ol>
  <li><strong>Teknik napas 4-6:</strong> tarik napas 4 hitungan, hembuskan 6 hitungan, ulang 3-5 menit.</li>
  <li><strong>Grounding 5-4-3-2-1:</strong> sebutkan 5 hal yang dilihat, 4 yang disentuh, 3 yang didengar, 2 yang dicium, 1 yang dirasa.</li>
  <li><strong>Jurnal pikiran:</strong> tulis kekhawatiran, lalu bedakan fakta, asumsi, dan rencana tindakan.</li>
  <li><strong>Batasi stimulasi:</strong> kurangi kafein berlebih, doom-scrolling, dan multitasking berlebihan.</li>
</ol>

<h2>Kapan Perlu Bantuan Profesional?</h2>
<p>Segera cari bantuan bila kecemasan membuat aktivitas harian terasa berat, mengganggu kuliah/kerja, atau menurunkan kualitas hidup secara signifikan. Konseling dan terapi psikologis dapat membantu menemukan akar masalah sekaligus membangun keterampilan regulasi emosi yang lebih sehat.</p>
<blockquote><p>Ingat: kemajuan tidak selalu linear. Hari yang berat bukan berarti kamu gagal, tetapi bagian dari proses pemulihan.</p></blockquote>`,
			CategoryID: healthCategory.ID,
			Image:      "article-mental.jpg",
		},
		{
			Title: "5 Teknik Pernapasan untuk Menenangkan Pikiran",
			Content: `<h2>Mengapa Pernapasan Membantu Menenangkan Pikiran?</h2>
<p>Napas adalah jembatan tercepat antara tubuh dan emosi. Saat cemas, pola napas cenderung pendek dan cepat. Dengan memperlambat ritme napas secara sadar, kita mengirim sinyal aman ke sistem saraf sehingga tubuh lebih mudah kembali tenang.</p>

<h2>1) Teknik 4-7-8</h2>
<p>Tarik napas melalui hidung selama 4 hitungan, tahan 7 hitungan, lalu hembuskan perlahan selama 8 hitungan. Ulang 4 siklus. Teknik ini efektif untuk menurunkan ketegangan sebelum tidur atau sebelum presentasi.</p>

<h2>2) Box Breathing (4-4-4-4)</h2>
<p>Tarik 4 hitungan, tahan 4 hitungan, hembuskan 4 hitungan, tahan 4 hitungan. Bayangkan membentuk sisi-sisi kotak. Cocok dilakukan saat jeda kerja agar fokus kembali stabil.</p>

<h2>3) Pernapasan Diafragma</h2>
<p>Letakkan satu tangan di dada dan satu tangan di perut. Saat menarik napas, utamakan perut mengembang lebih dulu. Pola ini membantu oksigenasi lebih efisien dan menurunkan ketegangan otot.</p>

<h2>4) Alternate Nostril Breathing</h2>
<p>Tutup satu lubang hidung, tarik napas dari sisi lain, lalu berganti saat menghembuskan napas. Lakukan 1-2 menit untuk meningkatkan rasa seimbang dan konsentrasi.</p>

<h2>5) Extended Exhale</h2>
<p>Gunakan rasio napas sederhana: tarik 3 hitungan, hembuskan 6 hitungan. Hembusan lebih panjang membantu aktivasi respon relaksasi. Teknik ini aman untuk pemula dan mudah dilakukan di mana saja.</p>

<h2>Tips Agar Konsisten</h2>
<ul>
  <li>Mulai dari 2-3 menit, 2 kali sehari.</li>
  <li>Pilih satu teknik favorit selama seminggu sebelum ganti teknik.</li>
  <li>Gunakan pengingat rutin: setelah bangun, sebelum belajar, sebelum tidur.</li>
</ul>
<p>Konsistensi kecil setiap hari biasanya lebih berdampak daripada sesi panjang tetapi jarang dilakukan.</p>`,
			CategoryID: tipsCategory.ID,
			Image:      "article-tips.jpg",
		},
		{
			Title: "Panduan Meditasi untuk Pemula",
			Content: `<h2>Meditasi untuk Pemula: Mulai dari yang Sederhana</h2>
<p>Banyak orang mengira meditasi berarti pikiran harus kosong total. Faktanya, meditasi adalah latihan untuk menyadari apa yang sedang terjadi dalam diri kita tanpa buru-buru bereaksi. Pikiran tetap muncul, dan itu normal.</p>

<h2>Persiapan 5 Menit</h2>
<ul>
  <li>Pilih tempat yang relatif tenang.</li>
  <li>Duduk nyaman dengan punggung tegak namun rileks.</li>
  <li>Set timer 5 menit agar kamu tidak terus melihat jam.</li>
</ul>

<h2>Langkah Latihan Dasar</h2>
<ol>
  <li>Tutup mata perlahan atau arahkan pandangan ke satu titik.</li>
  <li>Perhatikan sensasi napas di hidung, dada, atau perut.</li>
  <li>Saat pikiran mengembara, beri label singkat: “pikiran”.</li>
  <li>Kembalikan perhatian ke napas tanpa menyalahkan diri.</li>
  <li>Akhiri dengan satu napas dalam dan peregangan ringan.</li>
</ol>

<h2>Tantangan yang Umum Terjadi</h2>
<p><strong>“Saya tidak bisa diam.”</strong> Tidak apa-apa. Mulailah dari 2 menit.
<strong>“Pikiran saya ramai.”</strong> Itu bagian dari proses observasi.
<strong>“Saya bosan.”</strong> Coba variasi meditasi berjalan atau body scan.</p>

<h2>Manfaat yang Bisa Dirasakan</h2>
<ul>
  <li>Meningkatkan kemampuan fokus dan hadir pada momen saat ini.</li>
  <li>Membantu regulasi emosi saat menghadapi tekanan.</li>
  <li>Mendukung kualitas tidur melalui penurunan ketegangan mental.</li>
  <li>Meningkatkan kesadaran diri sehingga keputusan lebih tenang.</li>
</ul>

<p>Target realistis: latihan 5-10 menit per hari selama 2 minggu. Setelah itu, evaluasi perubahan pada kualitas tidur, fokus, dan respons emosionalmu.</p>`,
			CategoryID: meditasiCategory.ID,
			Image:      "article-meditation.jpg",
		},
		{
			Title: "Mengatasi Stres di Tempat Kerja",
			Content: `<h2>Stres Kerja: Normal, Tapi Perlu Dikelola</h2>
<p>Target yang padat, notifikasi tanpa henti, dan tuntutan multitasking dapat membuat energi mental terkuras. Stres kerja yang tidak dikelola berisiko menurunkan produktivitas, meningkatkan konflik interpersonal, dan memicu kelelahan emosional.</p>

<h2>Sinyal Awal Burnout</h2>
<ul>
  <li>Merasa lelah bahkan setelah istirahat.</li>
  <li>Sulit fokus pada tugas sederhana.</li>
  <li>Sinis, mudah marah, atau kehilangan motivasi.</li>
  <li>Kinerja menurun dan sering menunda pekerjaan.</li>
</ul>

<h2>Strategi yang Bisa Langsung Diterapkan</h2>
<ol>
  <li><strong>Peta prioritas harian:</strong> bedakan tugas penting-mendesak agar energi tidak habis untuk hal kecil.</li>
  <li><strong>Kerja berblok waktu:</strong> gunakan siklus fokus 25-50 menit diikuti jeda 5-10 menit.</li>
  <li><strong>Batas komunikasi:</strong> tentukan jam respons pesan agar tidak selalu “siaga”.</li>
  <li><strong>Ritual transisi:</strong> setelah kerja, lakukan aktivitas penutup seperti jalan 10 menit atau journaling singkat.</li>
  <li><strong>Komunikasi asertif:</strong> sampaikan kapasitas kerja secara jelas, termasuk estimasi waktu realistis.</li>
</ol>

<h2>Peran Tim dan Atasan</h2>
<p>Kesehatan mental bukan tanggung jawab individu saja. Lingkungan kerja yang sehat memerlukan distribusi beban yang adil, ekspektasi jelas, dan ruang diskusi saat kapasitas tim menurun.</p>

<h2>Kapan Harus Mencari Dukungan?</h2>
<p>Jika stres sudah mengganggu tidur, relasi, atau performa secara konsisten selama beberapa minggu, pertimbangkan konsultasi profesional. Intervensi lebih awal biasanya membuat pemulihan lebih cepat dan mencegah burnout berkepanjangan.</p>`,
			CategoryID: tipsCategory.ID,
			Image:      "article-stress.jpg",
		},
		{
			Title: "Pentingnya Tidur untuk Kesehatan Mental",
			Content: `<h2>Tidur Bukan Kemewahan, Melainkan Kebutuhan Dasar</h2>
<p>Saat tidur, otak melakukan “perawatan malam”: memperkuat memori, memproses emosi, dan memulihkan energi kognitif. Kurang tidur membuat kita lebih reaktif, sulit fokus, dan rentan overthinking.</p>

<h2>Dampak Kurang Tidur pada Kesehatan Mental</h2>
<ul>
  <li>Regulasi emosi menurun, sehingga lebih mudah cemas atau mudah tersinggung.</li>
  <li>Konsentrasi dan kemampuan mengambil keputusan menurun.</li>
  <li>Motivasi menurun dan kelelahan mental meningkat.</li>
  <li>Risiko gejala depresi dan kecemasan dapat meningkat bila berlangsung lama.</li>
</ul>

<h2>Target Durasi Tidur yang Direkomendasikan</h2>
<p>Mayoritas dewasa muda membutuhkan sekitar <strong>7-9 jam</strong> tidur per malam. Bukan hanya durasi, kualitas tidur (tidak sering terbangun, bangun lebih segar) juga sangat penting.</p>

<h2>Ritual Sleep Hygiene yang Efektif</h2>
<ol>
  <li>Tidur dan bangun di jam yang konsisten, termasuk akhir pekan.</li>
  <li>Kurangi paparan layar 60 menit sebelum tidur.</li>
  <li>Hindari kafein 6-8 jam sebelum waktu tidur.</li>
  <li>Gunakan tempat tidur khusus untuk tidur, bukan untuk kerja.</li>
  <li>Lakukan rutinitas menenangkan: mandi hangat, peregangan ringan, atau napas lambat.</li>
</ol>

<h2>Jika Sulit Tidur Terus-Menerus</h2>
<p>Bila keluhan insomnia terjadi lebih dari 3 malam per minggu selama beberapa minggu, pertimbangkan evaluasi profesional. Dengan bantuan yang tepat, pola tidur dapat dipulihkan secara bertahap.</p>
<p>Tidur yang baik adalah fondasi untuk belajar lebih optimal, emosi lebih stabil, dan relasi sosial yang lebih sehat.</p>`,
			CategoryID: healthCategory.ID,
			Image:      "article-sleep.jpg",
		},
	}

	for _, a := range articles {
		var existing model.Article
		findResult := db.Where("title = ?", a.Title).First(&existing)
		if findResult.Error == nil {
			// Repair broken/missing thumbnail file for existing seeded article.
			if !uploadAssetExists(existing.Thumbnail) {
				url := ""
				if u, ok := placeholderImages[a.Image]; ok {
					url = u
				}
				thumbnail := getOrDownloadImage(url, a.Image)

				if thumbnail != "" {
					if err := db.Model(&existing).Update("thumbnail", thumbnail).Error; err != nil {
						return err
					}
				}
			}

			updates := map[string]any{
				"content":             a.Content,
				"article_category_id": a.CategoryID,
				"status":              model.ArticleStatusPublished,
			}

			if err := db.Model(&existing).Updates(updates).Error; err != nil {
				return err
			}
			continue
		}
		if !errors.Is(findResult.Error, gorm.ErrRecordNotFound) {
			return findResult.Error
		}

		// Get thumbnail
		url := ""
		if u, ok := placeholderImages[a.Image]; ok {
			url = u
		}
		thumbnail := getOrDownloadImage(url, a.Image)

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

func uploadAssetExists(path string) bool {
	cleanPath := strings.TrimSpace(path)
	if cleanPath == "" {
		return false
	}

	// Absolute URLs are considered externally managed.
	if strings.HasPrefix(cleanPath, "http://") || strings.HasPrefix(cleanPath, "https://") {
		return true
	}

	localPath := strings.TrimPrefix(cleanPath, "/")
	if localPath == "" {
		return false
	}

	info, err := os.Stat(localPath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
