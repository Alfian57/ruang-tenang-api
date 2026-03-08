package production

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

func SeedTimedChallengeTemplates(db *gorm.DB) error {
	templates := []model.TimedChallengeTemplate{
		{Title: "Napas Tenang", Description: "Selesaikan 3 sesi latihan pernapasan dalam 1 jam", ChallengeType: "breathing", TargetValue: 3, DurationMinutes: 60, XPReward: 50, CoinReward: 10, Icon: "🌬️", IsActive: true},
		{Title: "Penulis Kilat", Description: "Tulis 2 jurnal dalam 30 menit", ChallengeType: "journal", TargetValue: 2, DurationMinutes: 30, XPReward: 40, CoinReward: 8, Icon: "✍️", IsActive: true},
		{Title: "Pelacak Mood", Description: "Catat mood-mu 3 kali dalam 2 jam", ChallengeType: "mood", TargetValue: 3, DurationMinutes: 120, XPReward: 35, CoinReward: 5, Icon: "😊", IsActive: true},
		{Title: "Pembaca Cepat", Description: "Baca 3 artikel kesehatan mental dalam 45 menit", ChallengeType: "article", TargetValue: 3, DurationMinutes: 45, XPReward: 45, CoinReward: 8, Icon: "📖", IsActive: true},
		{Title: "Pendengar Aktif", Description: "Dengarkan 2 lagu relaksasi dalam 20 menit", ChallengeType: "music", TargetValue: 2, DurationMinutes: 20, XPReward: 30, CoinReward: 5, Icon: "🎵", IsActive: true},
		{Title: "Teman Curhat", Description: "Kirim 5 pesan di forum dalam 30 menit", ChallengeType: "forum", TargetValue: 5, DurationMinutes: 30, XPReward: 40, CoinReward: 10, Icon: "💬", IsActive: true},
		{Title: "Penjelajah Peta", Description: "Selesaikan 2 landmark di peta perjalanan dalam 1 jam", ChallengeType: "map", TargetValue: 2, DurationMinutes: 60, XPReward: 60, CoinReward: 15, Icon: "🗺️", IsActive: true},
		{Title: "Sprint Pagi", Description: "Selesaikan 3 aktivitas apapun dalam 15 menit", ChallengeType: "any", TargetValue: 3, DurationMinutes: 15, XPReward: 55, CoinReward: 12, Icon: "⚡", IsActive: true},
		{Title: "Meditasi Marathon", Description: "Lakukan 5 sesi pernapasan dalam 2 jam", ChallengeType: "breathing", TargetValue: 5, DurationMinutes: 120, XPReward: 75, CoinReward: 20, Icon: "🧘", IsActive: true},
		{Title: "Kurator Konten", Description: "Baca 5 artikel dan tulis 1 jurnal dalam 1 jam", ChallengeType: "mixed", TargetValue: 6, DurationMinutes: 60, XPReward: 80, CoinReward: 25, Icon: "🎯", IsActive: true},
	}

	for _, t := range templates {
		var existing model.TimedChallengeTemplate
		if db.Where("title = ?", t.Title).First(&existing).RowsAffected == 0 {
			if err := db.Create(&t).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
