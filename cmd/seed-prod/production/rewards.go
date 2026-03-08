package production

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedRewards seeds the reward catalog for the gold coin shop
func SeedRewards(db *gorm.DB) error {
	rewards := []model.Reward{
		{Name: "Tema Profil: Aurora Borealis", Description: "Tema profil eksklusif dengan gradasi warna aurora yang indah untuk halaman profil Anda.", CoinCost: 150, Stock: -1, IsActive: true},
		{Name: "Badge Khusus: Bintang Harapan", Description: "Badge langka yang menunjukkan dedikasi Anda dalam menjaga kesehatan mental.", CoinCost: 300, Stock: 50, IsActive: true},
		{Name: "Avatar Frame: Pelangi", Description: "Bingkai avatar berwarna pelangi yang cerah dan membuat profil Anda lebih menarik.", CoinCost: 100, Stock: -1, IsActive: true},
		{Name: "Streak Freeze x2", Description: "Dua buah streak freeze tambahan untuk menjaga streak Anda tetap aman saat sedang istirahat.", CoinCost: 200, Stock: -1, IsActive: true},
		{Name: "XP Boost 2x (24 Jam)", Description: "Dapatkan XP double selama 24 jam untuk semua aktivitas Anda di platform.", CoinCost: 250, Stock: -1, IsActive: true},
		{Name: "Tema Profil: Sunset Beach", Description: "Tema profil dengan pemandangan pantai saat matahari terbenam yang menenangkan.", CoinCost: 150, Stock: -1, IsActive: true},
		{Name: "Banner Profil: Alam Pegunungan", Description: "Banner profil dengan pemandangan pegunungan yang indah dan menyegarkan pikiran.", CoinCost: 120, Stock: -1, IsActive: true},
		{Name: "Donasi ke Yayasan Kesehatan Mental", Description: "Coin Anda akan didonasikan untuk mendukung program kesehatan mental di Indonesia.", CoinCost: 500, Stock: -1, IsActive: true},
	}

	for _, reward := range rewards {
		var existing model.Reward
		if db.Where("name = ?", reward.Name).First(&existing).RowsAffected == 0 {
			if err := db.Create(&reward).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
