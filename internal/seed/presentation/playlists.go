package presentation

import (
	"fmt"

	"github.com/Alfian57/ruang-tenang-api/internal/model"
	"gorm.io/gorm"
)

// SeedPlaylists creates a compact editorial selection plus two playlists owned
// by presentation accounts. Existing seeded playlists and their item ordering
// are reconciled on every run.
func SeedPlaylists(db *gorm.DB) error {
	var admin model.User
	if err := db.Where("email = ? AND role = ?", presentationAdminEmail, model.RoleAdmin).First(&admin).Error; err != nil {
		return nil
	}

	allMembers, err := presentationUsers(db)
	if err != nil {
		return err
	}
	var members []model.User
	for _, user := range allMembers {
		if user.Role == model.RoleUser {
			members = append(members, user)
		}
	}

	var songs []model.Song
	if err := db.Where("title IN ?", curatedPresentationSongTitles).Order("id ASC").Find(&songs).Error; err != nil {
		return err
	}
	if len(songs) < 8 {
		return fmt.Errorf("playlist seeding needs the curated song catalog; run SeedSongCategories and SeedSongs first")
	}

	type playlistData struct {
		UserID          uint
		Name            string
		Description     string
		ThumbnailFile   string
		IsPublic        bool
		IsAdminPlaylist bool
		SongIndices     []int
	}

	playlists := []playlistData{
		{admin.ID, "Ritual Menjelang Tidur", "Piano pelan untuk menemani rutinitas malam. Musik dapat menjadi latar yang nyaman, tetapi tidak menggantikan bantuan profesional saat sulit tidur menetap.", "article-sleep-routine.webp", true, true, []int{4, 6, 7, 9}},
		{admin.ID, "Jeda Lima Menit", "Pilih satu lagu, letakkan layar sejenak, dan ambil jeda tanpa target untuk menjadi produktif.", "article-pause.webp", true, true, []int{0, 1, 2, 5}},
		{admin.ID, "Piano untuk Menata Fokus", "Instrumental tanpa lirik untuk menemani membaca, menulis jurnal, atau mengerjakan satu tugas dalam ritme yang terasa nyaman.", "article-calm-start.webp", true, true, []int{0, 3, 4, 5}},
		{admin.ID, "Suara Hening untuk Refleksi", "Pilihan ambient dan piano lembut untuk duduk sejenak bersama pikiran. Jika latihan terasa tidak nyaman, berhenti dan pilih aktivitas yang lebih aman untukmu.", "story-community-support.webp", true, true, []int{6, 7, 8, 9}},
	}

	if len(members) > 0 {
		playlists = append(playlists, playlistData{
			UserID: members[0].ID, Name: "Catatan Musik Gading",
			Description:   "Pilihan pribadi untuk jeda singkat di sela hari. Dibuat dari musik instrumental berlisensi terbuka.",
			ThumbnailFile: "article-pause.webp", IsPublic: false, SongIndices: []int{0, 3, 7},
		})
	}
	if len(members) > 1 {
		playlists = append(playlists, playlistData{
			UserID: members[1].ID, Name: "Langkah Pelan",
			Description:   "Playlist publik demo dengan pilihan piano sederhana untuk belajar atau membaca dengan jeda teratur.",
			ThumbnailFile: "article-calm-start.webp", IsPublic: true, SongIndices: []int{1, 4, 5, 8},
		})
	}
	ownerIDs := []uint{admin.ID}
	for _, member := range members {
		ownerIDs = append(ownerIDs, member.ID)
	}
	legacyNames := []string{"Relaksasi Malam", "Fokus dan Produktif", "Meditasi Pagi", "Playlist Santai Saya", "Mood Booster"}
	var legacyPlaylists []model.Playlist
	if err := db.Where("user_id IN ? AND name IN ?", ownerIDs, legacyNames).Find(&legacyPlaylists).Error; err != nil {
		return err
	}
	legacyPlaylistIDs := make([]uint, len(legacyPlaylists))
	for i, playlist := range legacyPlaylists {
		legacyPlaylistIDs[i] = playlist.ID
	}
	if len(legacyPlaylistIDs) > 0 {
		if err := db.Where("playlist_id IN ?", legacyPlaylistIDs).Delete(&model.PlaylistItem{}).Error; err != nil {
			return err
		}
		if err := db.Where("id IN ?", legacyPlaylistIDs).Delete(&model.Playlist{}).Error; err != nil {
			return err
		}
	}

	for _, data := range playlists {
		thumbnail := getSeedAsset(data.ThumbnailFile, "images")
		if thumbnail == "" {
			return fmt.Errorf("thumbnail not found for playlist %q (%s)", data.Name, data.ThumbnailFile)
		}

		var playlist model.Playlist
		findResult := db.Where("name = ? AND user_id = ?", data.Name, data.UserID).First(&playlist)
		if findResult.Error != nil && findResult.Error != gorm.ErrRecordNotFound {
			return findResult.Error
		}

		if findResult.Error == gorm.ErrRecordNotFound {
			playlist = model.Playlist{
				UserID: data.UserID, Name: data.Name, Description: data.Description,
				Thumbnail: thumbnail, IsPublic: data.IsPublic, IsAdminPlaylist: data.IsAdminPlaylist,
			}
			if err := db.Create(&playlist).Error; err != nil {
				return err
			}
		} else if err := db.Model(&playlist).Updates(map[string]any{
			"description": data.Description, "thumbnail": thumbnail,
			"is_public": data.IsPublic, "is_admin_playlist": data.IsAdminPlaylist,
		}).Error; err != nil {
			return err
		}

		if err := db.Where("playlist_id = ?", playlist.ID).Delete(&model.PlaylistItem{}).Error; err != nil {
			return err
		}
		for position, songIndex := range data.SongIndices {
			if songIndex < 0 || songIndex >= len(songs) {
				continue
			}
			if err := db.Create(&model.PlaylistItem{
				PlaylistID: playlist.ID,
				SongID:     songs[songIndex].ID,
				Position:   position,
			}).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
