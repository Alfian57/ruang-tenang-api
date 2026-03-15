package production

import (
	"fmt"
	"strings"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

type landmarkDef struct {
	LandmarkKey    string
	Name           string
	Description    string
	Icon           string
	UnlockType     model.MapUnlockType
	UnlockActivity string
	UnlockValue    int
	PositionX      int
	PositionY      int
	XPReward       int
	CoinReward     int
	DisplayOrder   int
}

type regionDef struct {
	RegionKey   string
	Name        string
	Description string
	Icon        string
	UnlockType  model.MapUnlockType
	UnlockValue int
	PositionX   int
	PositionY   int
	Order       int
	Landmarks   []landmarkDef
}

func buildLevelTaskSummary(landmarks []landmarkDef) string {
	parts := make([]string, 0, len(landmarks))
	for i, l := range landmarks {
		parts = append(parts, fmt.Sprintf("%d. %s", i+1, l.Description))
	}
	return strings.Join(parts, "\n")
}

// SeedMapRegions seeds the progress map regions and landmarks
func SeedMapRegions(db *gorm.DB) error {
	regions := []regionDef{
		{
			RegionKey:   "gerbang_awal",
			Name:        "Tier 1: Fondasi",
			Description: "Tahap awal untuk membangun kebiasaan positif dan konsisten.",
			Icon:        "🌅",
			UnlockType:  model.MapUnlockLevel,
			UnlockValue: 1,
			PositionX:   50,
			PositionY:   90,
			Order:       1,
			Landmarks: []landmarkDef{
				{"la_pertama_login", "Login Pertama", "Masuk ke aplikasi untuk pertama kali", "🚪", model.MapUnlockActivityCount, "login", 1, 30, 88, 10, 5, 1},
				{"la_pertama_mood", "Catat Mood Pertama", "Catat mood pertamamu", "😊", model.MapUnlockActivityCount, "mood", 1, 50, 85, 15, 5, 2},
				{"la_pertama_chat", "Chat AI Pertama", "Mulai percakapan dengan AI companion", "💬", model.MapUnlockActivityCount, "chat", 1, 70, 88, 15, 5, 3},
			},
		},
		{
			RegionKey:   "taman_ketenangan",
			Name:        "Tier 2: Stabilitas",
			Description: "Perkuat ritme aktivitas untuk menjaga kestabilan emosi.",
			Icon:        "🌿",
			UnlockType:  model.MapUnlockLevel,
			UnlockValue: 2,
			PositionX:   30,
			PositionY:   75,
			Order:       2,
			Landmarks: []landmarkDef{
				{"la_napas_pertama", "Napas Pertama", "Selesaikan sesi pernapasan pertama", "🌬️", model.MapUnlockActivityCount, "breathing", 1, 20, 78, 20, 10, 1},
				{"la_5_sesi_napas", "Penapas Terampil", "Selesaikan 5 sesi pernapasan", "🧘", model.MapUnlockActivityCount, "breathing", 5, 35, 72, 30, 15, 2},
				{"la_streak_3", "Api Konsisten", "Raih streak 3 hari berturut-turut", "🔥", model.MapUnlockStreak, "", 3, 45, 76, 25, 10, 3},
			},
		},
		{
			RegionKey:   "perpustakaan_bijak",
			Name:        "Tier 3: Eksplorasi",
			Description: "Eksplorasi wawasan baru dan perluas pemahaman kesehatan mental.",
			Icon:        "📚",
			UnlockType:  model.MapUnlockLevel,
			UnlockValue: 3,
			PositionX:   70,
			PositionY:   70,
			Order:       3,
			Landmarks: []landmarkDef{
				{"la_baca_artikel", "Pembaca Pemula", "Baca artikel pertamamu", "📖", model.MapUnlockActivityCount, "article", 1, 60, 72, 20, 10, 1},
				{"la_5_artikel", "Kutu Buku", "Baca 5 artikel berbeda", "📕", model.MapUnlockActivityCount, "article", 5, 75, 68, 30, 15, 2},
				{"la_tulis_artikel", "Penulis Pertama", "Tulis artikel pertamamu", "✍️", model.MapUnlockActivityCount, "write_article", 1, 80, 73, 35, 20, 3},
			},
		},
		{
			RegionKey:   "lembah_refleksi",
			Name:        "Tier 4: Refleksi",
			Description: "Perdalam refleksi diri lewat jurnal dan pelacakan emosi.",
			Icon:        "🏞️",
			UnlockType:  model.MapUnlockLevel,
			UnlockValue: 4,
			PositionX:   25,
			PositionY:   55,
			Order:       4,
			Landmarks: []landmarkDef{
				{"la_jurnal_pertama", "Jurnal Pertama", "Tulis entri jurnal pertamamu", "📝", model.MapUnlockActivityCount, "journal", 1, 15, 58, 25, 10, 1},
				{"la_5_jurnal", "Penulis Harian", "Tulis 5 entri jurnal", "📓", model.MapUnlockActivityCount, "journal", 5, 30, 52, 35, 15, 2},
				{"la_mood_7", "Pelacak Emosi", "Catat mood selama 7 hari", "📊", model.MapUnlockActivityCount, "mood", 7, 35, 57, 30, 15, 3},
			},
		},
		{
			RegionKey:   "alun_komunitas",
			Name:        "Tier 5: Koneksi",
			Description: "Bangun koneksi sehat lewat partisipasi komunitas.",
			Icon:        "🏛️",
			UnlockType:  model.MapUnlockLevel,
			UnlockValue: 5,
			PositionX:   65,
			PositionY:   50,
			Order:       5,
			Landmarks: []landmarkDef{
				{"la_forum_pertama", "Suara Pertama", "Buat postingan forum pertama", "🗣️", model.MapUnlockActivityCount, "forum", 1, 55, 52, 30, 15, 1},
				{"la_5_forum", "Kontributor Aktif", "Buat 5 postingan forum", "💬", model.MapUnlockActivityCount, "forum", 5, 70, 48, 40, 20, 2},
				{"la_streak_7", "Semangat Membara", "Raih streak 7 hari", "🔥", model.MapUnlockStreak, "", 7, 75, 53, 40, 20, 3},
			},
		},
		{
			RegionKey:   "puncak_harmoni",
			Name:        "Tier 6: Harmoni",
			Description: "Konsisten lintas aktivitas untuk menjaga keseimbangan yang lebih matang.",
			Icon:        "⛰️",
			UnlockType:  model.MapUnlockLevel,
			UnlockValue: 6,
			PositionX:   40,
			PositionY:   40,
			Order:       6,
			Landmarks: []landmarkDef{
				{"la_10_napas", "Master Pernapasan", "Selesaikan 10 sesi pernapasan", "🌊", model.MapUnlockActivityCount, "breathing", 10, 30, 42, 40, 20, 1},
				{"la_cerita_inspirasi", "Pencerita", "Bagikan cerita inspirasi pertama", "✨", model.MapUnlockActivityCount, "story", 1, 45, 38, 45, 25, 2},
				{"la_20_chat", "Teman Bicara", "Lakukan 20 sesi chat AI", "🤖", model.MapUnlockActivityCount, "chat", 20, 50, 43, 40, 20, 3},
			},
		},
		{
			RegionKey:   "hutan_kebijaksanaan",
			Name:        "Tier 7: Ketangguhan",
			Description: "Latih ketangguhan melalui konsistensi dan target jangka menengah.",
			Icon:        "🌳",
			UnlockType:  model.MapUnlockLevel,
			UnlockValue: 7,
			PositionX:   20,
			PositionY:   30,
			Order:       7,
			Landmarks: []landmarkDef{
				{"la_streak_14", "Tekad Baja", "Raih streak 14 hari berturut", "💪", model.MapUnlockStreak, "", 14, 10, 32, 50, 25, 1},
				{"la_15_jurnal", "Perenung Mendalam", "Tulis 15 entri jurnal", "📜", model.MapUnlockActivityCount, "journal", 15, 25, 28, 50, 25, 2},
				{"la_500_xp", "Pengumpul XP", "Kumpulkan 500 XP total", "⭐", model.MapUnlockXP, "", 500, 30, 33, 45, 20, 3},
			},
		},
		{
			RegionKey:   "danau_kedamaian",
			Name:        "Tier 8: Kematangan",
			Description: "Perjalananmu semakin matang dengan capaian berkelanjutan.",
			Icon:        "🏖️",
			UnlockType:  model.MapUnlockLevel,
			UnlockValue: 8,
			PositionX:   70,
			PositionY:   25,
			Order:       8,
			Landmarks: []landmarkDef{
				{"la_streak_21", "Kebiasaan Terbentuk", "Raih streak 21 hari", "🌟", model.MapUnlockStreak, "", 21, 60, 27, 60, 30, 1},
				{"la_50_chat", "Sahabat AI", "Lakukan 50 sesi chat AI", "🧠", model.MapUnlockActivityCount, "chat", 50, 75, 22, 55, 30, 2},
				{"la_1000_xp", "Veteran", "Kumpulkan 1000 XP total", "🏅", model.MapUnlockXP, "", 1000, 80, 27, 55, 25, 3},
			},
		},
		{
			RegionKey:   "menara_guardian",
			Name:        "Tier 9: Guardian",
			Description: "Tahap lanjutan bagi pengguna dengan kedisiplinan tinggi.",
			Icon:        "🏰",
			UnlockType:  model.MapUnlockLevel,
			UnlockValue: 9,
			PositionX:   45,
			PositionY:   15,
			Order:       9,
			Landmarks: []landmarkDef{
				{"la_streak_30", "Konsistensi Sempurna", "Raih streak 30 hari", "👑", model.MapUnlockStreak, "", 30, 35, 17, 70, 35, 1},
				{"la_30_jurnal", "Penulis Ulung", "Tulis 30 entri jurnal", "📘", model.MapUnlockActivityCount, "journal", 30, 50, 12, 65, 30, 2},
				{"la_2000_xp", "Elite", "Kumpulkan 2000 XP total", "💎", model.MapUnlockXP, "", 2000, 55, 17, 65, 35, 3},
			},
		},
		{
			RegionKey:   "nirwana",
			Name:        "Tier 10: Mastery",
			Description: "Puncak perjalanan dengan tantangan dan reward tertinggi.",
			Icon:        "🌈",
			UnlockType:  model.MapUnlockLevel,
			UnlockValue: 10,
			PositionX:   50,
			PositionY:   5,
			Order:       10,
			Landmarks: []landmarkDef{
				{"la_streak_60", "Legendaris", "Raih streak 60 hari", "🌠", model.MapUnlockStreak, "", 60, 40, 7, 100, 50, 1},
				{"la_5000_xp", "Grandmaster", "Kumpulkan 5000 XP total", "🏆", model.MapUnlockXP, "", 5000, 55, 3, 100, 50, 2},
				{"la_guardian", "Guardian Sejati", "Capai level 10 dan klaim semua landmark", "🛡️", model.MapUnlockLevel, "", 10, 50, 8, 150, 75, 3},
			},
		},
	}

	for _, r := range regions {
		if r.UnlockType == model.MapUnlockLevel && len(r.Landmarks) < 3 {
			return fmt.Errorf("region %s harus punya minimal 3 tugas/landmark, saat ini %d", r.RegionKey, len(r.Landmarks))
		}

		region := model.MapRegion{}
		if err := db.Where("region_key = ?", r.RegionKey).First(&region).Error; err != nil {
			if err != gorm.ErrRecordNotFound {
				return err
			}

			region = model.MapRegion{
				RegionKey:    r.RegionKey,
				Name:         r.Name,
				Description:  r.Description,
				Icon:         r.Icon,
				UnlockType:   r.UnlockType,
				UnlockValue:  r.UnlockValue,
				PositionX:    r.PositionX,
				PositionY:    r.PositionY,
				DisplayOrder: r.Order,
				IsActive:     true,
			}
			if createErr := db.Create(&region).Error; createErr != nil {
				return createErr
			}
		} else {
			region.Name = r.Name
			region.Description = r.Description
			region.Icon = r.Icon
			region.UnlockType = r.UnlockType
			region.UnlockValue = r.UnlockValue
			region.PositionX = r.PositionX
			region.PositionY = r.PositionY
			region.DisplayOrder = r.Order
			region.IsActive = true
			if updateErr := db.Save(&region).Error; updateErr != nil {
				return updateErr
			}
		}

		for _, l := range r.Landmarks {
			landmark := model.MapLandmark{}
			if err := db.Where("landmark_key = ?", l.LandmarkKey).First(&landmark).Error; err != nil {
				if err != gorm.ErrRecordNotFound {
					return err
				}

				landmark = model.MapLandmark{
					RegionID:       region.ID,
					LandmarkKey:    l.LandmarkKey,
					Name:           l.Name,
					Description:    l.Description,
					Icon:           l.Icon,
					UnlockType:     l.UnlockType,
					UnlockActivity: l.UnlockActivity,
					UnlockValue:    l.UnlockValue,
					PositionX:      l.PositionX,
					PositionY:      l.PositionY,
					XPReward:       l.XPReward,
					CoinReward:     l.CoinReward,
					DisplayOrder:   l.DisplayOrder,
					IsActive:       true,
				}
				if createErr := db.Create(&landmark).Error; createErr != nil {
					return createErr
				}
				continue
			}

			landmark.RegionID = region.ID
			landmark.Name = l.Name
			landmark.Description = l.Description
			landmark.Icon = l.Icon
			landmark.UnlockType = l.UnlockType
			landmark.UnlockActivity = l.UnlockActivity
			landmark.UnlockValue = l.UnlockValue
			landmark.PositionX = l.PositionX
			landmark.PositionY = l.PositionY
			landmark.XPReward = l.XPReward
			landmark.CoinReward = l.CoinReward
			landmark.DisplayOrder = l.DisplayOrder
			landmark.IsActive = true
			if updateErr := db.Save(&landmark).Error; updateErr != nil {
				return updateErr
			}
		}

		if r.UnlockType == model.MapUnlockLevel {
			taskSummary := buildLevelTaskSummary(r.Landmarks)
			if err := db.Model(&model.LevelConfig{}).
				Where("level = ?", r.UnlockValue).
				Update("task_description", taskSummary).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
