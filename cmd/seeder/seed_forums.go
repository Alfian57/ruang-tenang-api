package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/Alfian57/ruang-tenang-api/internal/models"
	"gorm.io/gorm"
)

func seedForums(db *gorm.DB) {
	log.Println("📝 Seeding forum categories...")
	forumCategories := []string{
		"Diskusi Umum",
		"Curhat & Keluh Kesah",
		"Dukungan Emosional",
		"Tips Mengelola Stres",
		"Kisah Inspiratif",
		"Kesehatan Mental di Tempat Kerja",
	}

	for _, catName := range forumCategories {
		var existing models.ForumCategory
		if db.Where("name = ?", catName).First(&existing).RowsAffected == 0 {
			cat := models.ForumCategory{Name: catName}
			db.Create(&cat)
			log.Printf("  ✓ Created forum category: %s", catName)
		}
	}

	// Fetch users for generic usage
	var users []models.User
	if err := db.Find(&users).Error; err != nil || len(users) == 0 {
		log.Println("  ⚠️ No users found, skipping forum seeding details")
		return
	}

	// Fetch categories
	var categories []models.ForumCategory
	db.Find(&categories)

	log.Println("📝 Seeding forums threads...")
	sampleForums := []struct {
		Title   string
		Content string
		CatName string
	}{
		{
			Title:   "Bagaimana cara kalian mengatasi burnout?",
			Content: "Belakangan ini saya merasa sangat lelah dengan pekerjaan. Ada tips untuk mengatasi burnout tanpa harus resign?",
			CatName: "Kesehatan Mental di Tempat Kerja",
		},
		{
			Title:   "Cerita sukses sembuh dari anxiety",
			Content: "Saya ingin berbagi pengalaman saya sembuh dari anxiety disorder setelah 2 tahun berjuang. Semoga bisa menginspirasi teman-teman semua.",
			CatName: "Kisah Inspiratif",
		},
		{
			Title:   "Butuh teman curhat",
			Content: "Lagi merasa down banget hari ini, ada yang bersedia mendengarkan?",
			CatName: "Curhat & Keluh Kesah",
		},
		{
			Title:   "Rekomendasi buku self-improvement",
			Content: "Ada rekomendasi buku bagus untuk meningkatkan kepercayaan diri?",
			CatName: "Diskusi Umum",
		},
	}

	for _, f := range sampleForums {
		var catID uint
		for _, c := range categories {
			if c.Name == f.CatName {
				catID = c.ID
				break
			}
		}

		// Random user
		user := users[rand.Intn(len(users))]

		var existing models.Forum
		if db.Where("title = ?", f.Title).First(&existing).RowsAffected == 0 {
			forum := models.Forum{
				UserID:     user.ID,
				CategoryID: &catID,
				Title:      f.Title,
				Content:    f.Content,
				CreatedAt:  time.Now().Add(-time.Duration(rand.Intn(100)) * time.Hour),
			}
			db.Create(&forum)
			log.Printf("  ✓ Created forum thread: %s", f.Title)

			// Create some replies
			numReplies := rand.Intn(5) + 1
			for i := 0; i < numReplies; i++ {
				replyUser := users[rand.Intn(len(users))]
				post := models.ForumPost{
					ForumID:   forum.ID,
					UserID:    replyUser.ID,
					Content:   "Semangat ya! Terima kasih sudah berbagi.",
					CreatedAt: time.Now().Add(-time.Duration(rand.Intn(50)) * time.Minute),
				}
				db.Create(&post)
			}
			log.Printf("    - Added %d replies", numReplies)

			// Add likes
			numLikes := rand.Intn(10)
			for i := 0; i < numLikes; i++ {
				likeUser := users[rand.Intn(len(users))]
				var like models.ForumLike
				if db.Where("forum_id = ? AND user_id = ?", forum.ID, likeUser.ID).First(&like).RowsAffected == 0 {
					db.Create(&models.ForumLike{ForumID: forum.ID, UserID: likeUser.ID})
				}
			}
			log.Printf("    - Added %d likes", numLikes)

		}
	}
}
