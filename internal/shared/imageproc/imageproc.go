// Package imageproc menyediakan pemrosesan gambar ringan **murni Go**
// (tanpa CGO) untuk upload: menurunkan resolusi gambar besar dan membuat
// thumbnail. Dipakai agar gambar yang disajikan lebih kecil & cepat dimuat.
//
// Catatan format: encoder WebP tidak tersedia di pustaka standar Go maupun
// `golang.org/x/image` (hanya decoder). Meng-encode WebP butuh CGO
// (mis. libwebp/govips) yang menambah kompleksiti build & dependensi sistem,
// sehingga sengaja TIDAK dipakai. Sebagai gantinya, gambar besar di-encode
// ulang ke JPEG berkualitas tinggi (mendekati ukuran WebP untuk foto) dan
// thumbnail JPEG dibuat. Di sisi web, `next/image` tetap menyajikan
// AVIF/WebP on-the-fly dari hasil ini.
package imageproc

import (
	"bytes"
	"image"
	"image/jpeg"
	_ "image/gif"  // daftarkan decoder GIF untuk image.Decode
	_ "image/png"  // daftarkan decoder PNG untuk image.Decode

	xdraw "golang.org/x/image/draw"
)

const (
	// MaxDimension membatasi sisi terpanjang gambar utama agar tak terlalu
	// besar (foto kamera bisa 4000px+). 1600px cukup tajam untuk layar.
	MaxDimension = 1600
	// ThumbDimension adalah sisi terpanjang thumbnail.
	ThumbDimension = 400
	// JPEGQuality untuk encode ulang gambar utama.
	JPEGQuality = 82
	// ThumbQuality untuk thumbnail (sedikit lebih rendah, cukup untuk preview).
	ThumbQuality = 75
)

// Result memuat byte hasil pemrosesan.
type Result struct {
	// Main adalah gambar utama (di-downscale bila perlu) sebagai JPEG.
	// Nil bila gambar tidak perlu/atau tidak dapat diproses (pakai asli).
	Main []byte
	// Thumbnail adalah versi kecil sebagai JPEG. Nil bila gagal.
	Thumbnail []byte
	// Width/Height adalah dimensi gambar utama setelah diproses.
	Width  int
	Height int
}

// Process men-decode [data], menurunkan resolusi gambar utama bila melebihi
// [MaxDimension], dan membuat thumbnail. Mengembalikan (nil, false) bila data
// bukan gambar raster yang didukung (mis. GIF animasi / format lain) sehingga
// pemanggil bisa menyimpan berkas asli apa adanya.
func Process(data []byte) (*Result, bool) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, false
	}
	// GIF animasi: jangan proses (decode hanya mengambil frame pertama).
	if format == "gif" {
		return nil, false
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	res := &Result{Width: w, Height: h}

	// Gambar utama: downscale hanya bila melebihi batas.
	if w > MaxDimension || h > MaxDimension {
		mw, mh := fitWithin(w, h, MaxDimension)
		scaled := scale(img, mw, mh)
		if jpg, err := encodeJPEG(scaled, JPEGQuality); err == nil {
			res.Main = jpg
			res.Width, res.Height = mw, mh
		}
	}

	// Thumbnail selalu dibuat (untuk daftar/preview).
	tw, th := fitWithin(w, h, ThumbDimension)
	thumb := scale(img, tw, th)
	if jpg, err := encodeJPEG(thumb, ThumbQuality); err == nil {
		res.Thumbnail = jpg
	}

	// Bila tidak ada yang dihasilkan, anggap tidak perlu diproses.
	if res.Main == nil && res.Thumbnail == nil {
		return nil, false
	}
	return res, true
}

// fitWithin menghitung dimensi baru agar sisi terpanjang = max, menjaga rasio.
func fitWithin(w, h, max int) (int, int) {
	if w <= max && h <= max {
		return w, h
	}
	if w >= h {
		return max, int(float64(h) * float64(max) / float64(w))
	}
	return int(float64(w) * float64(max) / float64(h)), max
}

func scale(src image.Image, w, h int) image.Image {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	// CatmullRom: kualitas penskalaan baik, tetap murni Go.
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)
	return dst
}

func encodeJPEG(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
