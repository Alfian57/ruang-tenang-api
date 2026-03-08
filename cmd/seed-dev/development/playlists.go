package development

import (
	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedPlaylists seeds admin-curated and user playlists for development
func SeedPlaylists(db *gorm.DB) error {
	// Get admin user for admin playlists
	var admin model.User
	if err := db.Where("role = ?", model.RoleAdmin).First(&admin).Error; err != nil {
		return nil
	}

	// Get member users for user playlists
	var members []model.User
	if err := db.Where("role = ?", model.RoleMember).Order("id ASC").Find(&members).Error; err != nil {
		return err
	}

	// Get all songs
	var songs []model.Song
	if err := db.Order("id ASC").Find(&songs).Error; err != nil {
		return err
	}
	if len(songs) < 5 {
		return nil // Need at least some songs
	}

	type playlistData struct {
		UserID          uint
		Name            string
		Description     string
		IsPublic        bool
		IsAdminPlaylist bool
		SongIndices     []int // indices into the songs slice
	}

	playlists := []playlistData{
		// Admin-curated playlists
		{
			UserID:          admin.ID,
			Name:            "Relaksasi Malam",
			Description:     "Koleksi musik menenangkan untuk menemani tidur Anda. Dipilih khusus oleh tim Ruang Tenang.",
			IsPublic:        true,
			IsAdminPlaylist: true,
			SongIndices:     []int{0, 2, 4, 6},
		},
		{
			UserID:          admin.ID,
			Name:            "Fokus dan Produktif",
			Description:     "Musik instrumental yang membantu konsentrasi saat belajar atau bekerja.",
			IsPublic:        true,
			IsAdminPlaylist: true,
			SongIndices:     []int{1, 3, 5},
		},
		{
			UserID:          admin.ID,
			Name:            "Meditasi Pagi",
			Description:     "Mulai hari Anda dengan ketenangan melalui koleksi musik meditasi pilihan.",
			IsPublic:        true,
			IsAdminPlaylist: true,
			SongIndices:     []int{0, 1, 2, 3, 4},
		},
	}

	// Add user playlists if members exist
	if len(members) > 0 {
		playlists = append(playlists, playlistData{
			UserID:      members[0].ID,
			Name:        "Playlist Santai Saya",
			Description: "Kumpulan lagu favorit untuk bersantai.",
			IsPublic:    false,
			SongIndices: []int{0, 3, 5, 7},
		})
	}
	if len(members) > 1 {
		playlists = append(playlists, playlistData{
			UserID:      members[1].ID,
			Name:        "Mood Booster",
			Description: "Musik yang selalu bikin mood lebih baik.",
			IsPublic:    true,
			SongIndices: []int{1, 2, 6},
		})
	}

	for _, pd := range playlists {
		var existing model.Playlist
		if db.Where("name = ? AND user_id = ?", pd.Name, pd.UserID).First(&existing).RowsAffected > 0 {
			continue
		}

		playlist := model.Playlist{
			UserID:          pd.UserID,
			Name:            pd.Name,
			Description:     pd.Description,
			IsPublic:        pd.IsPublic,
			IsAdminPlaylist: pd.IsAdminPlaylist,
		}

		if err := db.Create(&playlist).Error; err != nil {
			return err
		}

		// Add songs to playlist
		for pos, songIdx := range pd.SongIndices {
			if songIdx >= len(songs) {
				continue
			}
			item := model.PlaylistItem{
				PlaylistID: playlist.ID,
				SongID:     songs[songIdx].ID,
				Position:   pos,
			}
			if err := db.Create(&item).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
