// Package handler untuk handler captcha
package handler

import (
	crand "crypto/rand"
	"encoding/hex"
	"image/png"
	"io"
	"log"
	"math/rand"
	"time"

	"github.com/fogleman/gg"
	"github.com/gofiber/fiber/v2"
	"github.com/qwerius/gonuxt/internal/store"
)

type CaptchaHandler struct{}

func NewCaptchaHandler() *CaptchaHandler {
	return &CaptchaHandler{}
}

const captchaChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // huruf & angka aman

func randomCaptchaText(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = captchaChars[rand.Intn(len(captchaChars))]
	}
	return string(b)
}

func randomID() string {
	b := make([]byte, 16)
	_, _ = io.ReadFull(crand.Reader, b)
	return hex.EncodeToString(b)
}

// drawCaptcha menggunakan fogleman/gg → huruf + angka, ada rotasi & noise
func drawCaptcha(text string) *gg.Context {
	const width, height = 200, 80

	dc := gg.NewContext(width, height)
	dc.SetRGB(1, 1, 1) // background putih
	dc.Clear()

	// Load font (pastikan file font ada di folder assets/fonts)
	if err := dc.LoadFontFace("./assets/fonts/RobotoSlab-Bold.ttf", 36); err != nil {
		log.Println("Failed to load font:", err)
	}

	// gambar huruf satu per satu dengan rotasi acak ±10°
	for i, ch := range text {
		x := 30 + float64(i)*35 + float64(rand.Intn(10)-5) // jarak antar huruf lebih lebar
		y := 50 + float64(rand.Intn(10)-5)
		angle := gg.Radians(float64(rand.Intn(20) - 10))
		dc.Push()
		dc.RotateAbout(angle, x, y)
		dc.SetRGB(float64(rand.Intn(120))/255.0, float64(rand.Intn(120))/255.0, float64(rand.Intn(120))/255.0)
		dc.DrawStringAnchored(string(ch), x, y, 0.5, 0.5)
		dc.Pop()
	}

	// noise titik acak
	for i := 0; i < 100; i++ {
		x := rand.Float64() * width
		y := rand.Float64() * height
		dc.SetRGB(float64(rand.Intn(220))/255.0, float64(rand.Intn(220))/255.0, float64(rand.Intn(220))/255.0)
		dc.DrawPoint(x, y, 1)
		dc.Fill()
	}

	// garis pengganggu
	for i := 0; i < rand.Intn(4)+2; i++ {
		x1 := rand.Float64() * width
		y1 := rand.Float64() * height
		x2 := rand.Float64() * width
		y2 := rand.Float64() * height
		dc.SetRGBA(float64(80+rand.Intn(176))/255.0, 0, 0, float64(140+rand.Intn(100))/255.0)
		dc.SetLineWidth(1 + rand.Float64()*2)
		dc.DrawLine(x1, y1, x2, y2)
		dc.Stroke()
	}

	return dc
}

// GenerateCaptcha endpoint untuk buat captcha
func (h *CaptchaHandler) GenerateCaptcha(c *fiber.Ctx) error {
	text := randomCaptchaText(5)
	id := randomID()

	store.Store.Add(id, text, 5*time.Minute)

	dc := drawCaptcha(text)

	c.Set(fiber.HeaderContentType, "image/png")
	c.Set("X-Captcha-ID", id)
	c.Status(fiber.StatusOK)

	return png.Encode(c.Response().BodyWriter(), dc.Image())
}
