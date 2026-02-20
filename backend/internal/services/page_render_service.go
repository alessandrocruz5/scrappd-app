package services

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alessandrocruz5/scrappd-app/backend/internal/repository"
	"github.com/alessandrocruz5/scrappd-app/backend/pkg/utils"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	_ "golang.org/x/image/webp"
)

const (
	defaultRenderScale     = 2.0
	maxRenderScale         = 4.0
	maxRenderPixelCount    = 25_000_000
	defaultJPEGQuality     = 92
	backgroundFetchTimeout = 6 * time.Second
)

type PageRenderOptions struct {
	Scale        float64
	Format       string
	Quality      int
	TargetWidth  int
	TargetHeight int
}

type PageRenderService interface {
	RenderPage(ctx context.Context, userID, pageID uuid.UUID, opts PageRenderOptions) ([]byte, string, error)
}

type pageRenderService struct {
	pagesRepo     repository.PagesRepository
	pageItemsRepo repository.PageItemsRepository
	itemsRepo     repository.ItemsRepository
	storage       Storage
}

func NewPageRenderService(
	pagesRepo repository.PagesRepository,
	pageItemsRepo repository.PageItemsRepository,
	itemsRepo repository.ItemsRepository,
	storage Storage,
) PageRenderService {
	return &pageRenderService{
		pagesRepo:     pagesRepo,
		pageItemsRepo: pageItemsRepo,
		itemsRepo:     itemsRepo,
		storage:       storage,
	}
}

func (s *pageRenderService) RenderPage(ctx context.Context, userID, pageID uuid.UUID, opts PageRenderOptions) ([]byte, string, error) {
	scale := opts.Scale
	if scale <= 0 {
		scale = defaultRenderScale
	}
	if scale > maxRenderScale {
		return nil, "", utils.ErrBadRequest(fmt.Sprintf("Scale exceeds max %.1f", maxRenderScale), nil)
	}

	format := strings.ToLower(strings.TrimSpace(opts.Format))
	if format == "" {
		format = "png"
	}
	if format == "jpg" {
		format = "jpeg"
	}
	if format != "png" && format != "jpeg" {
		return nil, "", utils.ErrBadRequest("Unsupported render format", nil)
	}

	quality := opts.Quality
	if quality <= 0 || quality > 100 {
		quality = defaultJPEGQuality
	}

	page, err := s.pagesRepo.GetByID(ctx, userID, pageID)
	if err != nil {
		return nil, "", err
	}

	width := int(math.Round(float64(page.CanvasWidth) * scale))
	height := int(math.Round(float64(page.CanvasHeight) * scale))
	if width <= 0 || height <= 0 {
		return nil, "", utils.ErrBadRequest("Invalid page canvas size", nil)
	}
	if width*height > maxRenderPixelCount {
		return nil, "", utils.ErrBadRequest("Render size too large", nil)
	}

	bgColor := parseHexColor(page.BackgroundColor, color.White)
	canvas := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{C: bgColor}, image.Point{}, draw.Src)

	if page.BackgroundImageURL != nil && *page.BackgroundImageURL != "" {
		bgImg, err := fetchRemoteImage(ctx, *page.BackgroundImageURL)
		if err == nil {
			bgFilled := imaging.Fill(bgImg, width, height, imaging.Center, imaging.Lanczos)
			draw.Draw(canvas, canvas.Bounds(), bgFilled, image.Point{}, draw.Over)
		}
	}

	items, err := s.pageItemsRepo.ListByPage(ctx, userID, pageID)
	if err != nil {
		return nil, "", err
	}

	for _, pageItem := range items {
		item, err := s.itemsRepo.GetByID(ctx, userID, pageItem.ItemID)
		if err != nil {
			return nil, "", err
		}

		imageKey := item.OriginalImageKey
		if item.ProcessedImageKey != nil && *item.ProcessedImageKey != "" {
			imageKey = *item.ProcessedImageKey
		}

		imgBytes, err := s.storage.Download(ctx, imageKey)
		if err != nil {
			return nil, "", utils.ErrStorage("Failed to download item image", err)
		}

		decoded, _, err := image.Decode(bytes.NewReader(imgBytes))
		if err != nil {
			return nil, "", utils.ErrInternalServer("Failed to decode item image", err)
		}

		targetW := int(math.Round(pageItem.Width * scale))
		targetH := int(math.Round(pageItem.Height * scale))
		if targetW <= 0 || targetH <= 0 {
			continue
		}

		// Match the editor's BoxFit.contain behavior to keep the full subject visible.
		resized := fitImageInFrame(decoded, targetW, targetH)
		angleDeg := pageItem.Rotation * 180 / math.Pi
		rotated := imaging.Rotate(resized, angleDeg, color.Transparent)

		if pageItem.Opacity > 0 && pageItem.Opacity < 1 {
			rotated = applyOpacity(rotated, pageItem.Opacity)
		}

		centerX := (pageItem.PositionX*scale + float64(targetW)/2)
		centerY := (pageItem.PositionY*scale + float64(targetH)/2)
		drawX := int(math.Round(centerX - float64(rotated.Bounds().Dx())/2))
		drawY := int(math.Round(centerY - float64(rotated.Bounds().Dy())/2))

		draw.Draw(
			canvas,
			image.Rect(drawX, drawY, drawX+rotated.Bounds().Dx(), drawY+rotated.Bounds().Dy()),
			rotated,
			image.Point{},
			draw.Over,
		)
	}

	outputImage := image.Image(canvas)
	if opts.TargetWidth > 0 || opts.TargetHeight > 0 {
		targetW := opts.TargetWidth
		targetH := opts.TargetHeight
		if targetW > 0 && targetH == 0 {
			targetH = int(math.Round(float64(targetW) * float64(page.CanvasHeight) / float64(page.CanvasWidth)))
		}
		if targetH > 0 && targetW == 0 {
			targetW = int(math.Round(float64(targetH) * float64(page.CanvasWidth) / float64(page.CanvasHeight)))
		}
		if targetW <= 0 || targetH <= 0 {
			return nil, "", utils.ErrBadRequest("Invalid target size", nil)
		}
		if targetW*targetH > maxRenderPixelCount {
			return nil, "", utils.ErrBadRequest("Target size too large", nil)
		}
		outputImage = imaging.Resize(canvas, targetW, targetH, imaging.Lanczos)
	}

	var output bytes.Buffer
	switch format {
	case "jpeg":
		if err := jpeg.Encode(&output, outputImage, &jpeg.Options{Quality: quality}); err != nil {
			return nil, "", utils.ErrInternalServer("Failed to encode JPEG", err)
		}
		return output.Bytes(), "image/jpeg", nil
	default:
		if err := png.Encode(&output, outputImage); err != nil {
			return nil, "", utils.ErrInternalServer("Failed to encode PNG", err)
		}
		return output.Bytes(), "image/png", nil
	}
}

func parseHexColor(input string, fallback color.Color) color.Color {
	s := strings.TrimSpace(strings.TrimPrefix(input, "#"))
	if len(s) != 6 {
		return fallback
	}
	r, err1 := strconv.ParseUint(s[0:2], 16, 8)
	g, err2 := strconv.ParseUint(s[2:4], 16, 8)
	b, err3 := strconv.ParseUint(s[4:6], 16, 8)
	if err1 != nil || err2 != nil || err3 != nil {
		return fallback
	}
	return color.NRGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 0xFF}
}

func applyOpacity(src image.Image, opacity float64) *image.NRGBA {
	if opacity <= 0 {
		opacity = 0
	}
	if opacity >= 1 {
		opacity = 1
	}
	bounds := src.Bounds()
	dst := image.NewNRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := src.At(x, y).RGBA()
			alpha := uint8(float64(a>>8) * opacity)
			dst.SetNRGBA(x, y, color.NRGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: alpha,
			})
		}
	}
	return dst
}

func fitImageInFrame(src image.Image, targetW, targetH int) *image.NRGBA {
	fitted := imaging.Fit(src, targetW, targetH, imaging.Lanczos)
	frame := image.NewNRGBA(image.Rect(0, 0, targetW, targetH))
	offsetX := (targetW - fitted.Bounds().Dx()) / 2
	offsetY := (targetH - fitted.Bounds().Dy()) / 2
	draw.Draw(
		frame,
		image.Rect(
			offsetX,
			offsetY,
			offsetX+fitted.Bounds().Dx(),
			offsetY+fitted.Bounds().Dy(),
		),
		fitted,
		image.Point{},
		draw.Over,
	)
	return frame
}

func fetchRemoteImage(ctx context.Context, url string) (image.Image, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: backgroundFetchTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("background fetch failed: %s", resp.Status)
	}
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}
	return img, nil
}
