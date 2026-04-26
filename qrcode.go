package tools

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/skip2/go-qrcode"
)

const (
	// QRCodeDefaultLogoCover is the default logo cover ratio when logo is enabled.
	QRCodeDefaultLogoCover = 0.20
)

var (
	ErrMissingText        = errors.New("tools/qr: text is missing")
	ErrOutputPathEmpty    = errors.New("tools/qr: output path cannot be empty")
	ErrOutputWriterNil    = errors.New("tools/qr: output writer cannot be nil")
	ErrLogoSourceConflict = errors.New("tools/qr: logo source conflict")
	ErrLogoCoverNeedsLogo = errors.New("tools/qr: logo-cover requires logo")
)

// QRCodeRecoveryLevel is the QR error correction level used by this package's public API.
type QRCodeRecoveryLevel int

const (
	QRCodeRecoveryLow     QRCodeRecoveryLevel = QRCodeRecoveryLevel(qrcode.Low)
	QRCodeRecoveryMedium  QRCodeRecoveryLevel = QRCodeRecoveryLevel(qrcode.Medium)
	QRCodeRecoveryHigh    QRCodeRecoveryLevel = QRCodeRecoveryLevel(qrcode.High)
	QRCodeRecoveryHighest QRCodeRecoveryLevel = QRCodeRecoveryLevel(qrcode.Highest)
)

// qrcodeCoverRatioTolerance allows tiny rounding drift when comparing actual
// cover ratio and requested target ratio.
const qrcodeCoverRatioTolerance = 0.995

// QRCodeOptions contains common QR generation options shared by file/writer APIs.
type (
	QRCodeOptions struct {
		Level QRCodeRecoveryLevel

		// Version accepts 0..40. 0 means auto-select the minimum valid version.
		Version int
		// Size is the output PNG width/height in pixels.
		Size int

		// LogoCover is the requested logo area ratio over QR code area.
		// When logo is present and LogoCover is 0, QRCodeDefaultLogoCover is used.
		LogoCover float64

		// DisableForceHighestWhenLogo disables the default behavior that upgrades
		// recovery level to highest when logo is present.
		DisableForceHighestWhenLogo bool

		// LogoPath is an optional logo image file path.
		LogoPath string
		// LogoReader is an optional logo image reader.
		// Set either LogoPath or LogoReader, not both.
		LogoReader io.Reader
	}

	params struct {
		level      qrcode.RecoveryLevel
		version    int
		size       int
		logoImg    image.Image
		coverRatio float64

		logoCloser io.Closer
	}
)

func (l QRCodeRecoveryLevel) toRecoveryLevel() (qrcode.RecoveryLevel, error) {
	switch l {
	case QRCodeRecoveryLow:
		return qrcode.Low, nil
	case QRCodeRecoveryMedium:
		return qrcode.Medium, nil
	case QRCodeRecoveryHigh:
		return qrcode.High, nil
	case QRCodeRecoveryHighest:
		return qrcode.Highest, nil
	default:
		return 0, errors.New("tools/qr: invalid recovery level")
	}
}

func (q QRCodeOptions) hasLogo() bool {
	return q.LogoPath != "" || q.LogoReader != nil
}

func (q QRCodeOptions) toParams() (*params, error) {
	if q.Size <= 0 {
		return nil, errors.New("tools/qr: size must be greater than zero")
	}
	if q.Version < 0 || q.Version > 40 {
		return nil, errors.New("tools/qr: version allowed value 0..40")
	}
	nativeLevel, err := q.Level.toRecoveryLevel()
	if err != nil {
		return nil, err
	}

	ps := &params{level: nativeLevel, version: q.Version, size: q.Size}

	if q.LogoPath != "" && q.LogoReader != nil {
		return nil, ErrLogoSourceConflict
	}
	if q.LogoCover < 0 || q.LogoCover >= 1 {
		return nil, errors.New("tools/qr: logo cover allowed value 0..1")
	}
	if !q.hasLogo() {
		if q.LogoCover != 0 {
			return nil, ErrLogoCoverNeedsLogo
		}
	} else {
		// By default, logo mode forces highest recovery; callers can opt out.
		if !q.DisableForceHighestWhenLogo {
			ps.level = qrcode.Highest
		}
		if q.LogoCover == 0 {
			ps.coverRatio = QRCodeDefaultLogoCover
		} else {
			ps.coverRatio = q.LogoCover
		}
		logoReader := q.LogoReader
		if q.LogoPath != "" {
			logoFile, err := os.Open(filepath.Clean(q.LogoPath))
			if err != nil {
				return nil, fmt.Errorf("tools/qr: open logo file failed: %w", err)
			}
			ps.logoCloser = logoFile
			logoReader = logoFile
		}
		ps.logoImg, _, err = image.Decode(logoReader)
		if err != nil {
			return nil, fmt.Errorf("tools/qr: decode logo image failed: %w", err)
		}
	}
	return ps, nil
}

func (ps *params) hasLogo() bool {
	return ps.logoImg != nil
}

func (ps *params) Close() {
	if ps != nil && ps.logoCloser != nil {
		_ = ps.logoCloser.Close()
	}
}

// GenerateQRCode generates a QR image file with explicit text and options.
func GenerateQRCode(text, output string, options QRCodeOptions) error {
	// Normalize output path and ensure parent directory exists.
	output = strings.TrimSpace(output)
	if output == "" {
		return ErrOutputPathEmpty
	}
	output = filepath.Clean(output)

	// Render into memory first; persist to disk only after full generation succeeds.
	var outBuffer bytes.Buffer
	var err error
	if err = GenerateQRCodeToWriter(text, &outBuffer, options); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return fmt.Errorf("tools/qr: prepare output dir failed: %w", err)
	}
	if err = os.WriteFile(output, outBuffer.Bytes(), 0o644); err != nil {
		return fmt.Errorf("tools/qr: write output file failed: %w", err)
	}
	return nil
}

// GenerateQRCodeToWriter generates a QR code PNG to writer with explicit text and options.
func GenerateQRCodeToWriter(text string, output io.Writer, options QRCodeOptions) error {
	if text == "" {
		return ErrMissingText
	}
	// Writer-based API mirrors file-based behavior but writes PNG bytes to caller output.
	if output == nil {
		return ErrOutputWriterNil
	}

	ps, err := options.toParams()
	if err != nil {
		return err
	}
	defer ps.Close()
	if !ps.hasLogo() {
		return generateWithoutLogo(text, output, ps)
	}
	return generateWithLogo(text, output, ps)
}

// isCoverSatisfied applies a small tolerance to absorb pixel rounding drift.
func isCoverSatisfied(actualCover, targetCover float64) bool {
	return actualCover >= targetCover*qrcodeCoverRatioTolerance
}

// writePNGToWriter encodes image into PNG stream and wraps encode errors.
func writePNGToWriter(w io.Writer, img image.Image) error {
	if err := png.Encode(w, img); err != nil {
		return fmt.Errorf("tools/qr: encode output png failed: %w", err)
	}
	return nil
}

// buildQRCode creates a QRCode object in auto or forced-version mode.
func buildQRCode(text string, level qrcode.RecoveryLevel, version int) (*qrcode.QRCode, error) {
	if version == 0 {
		qr, err := qrcode.New(text, level)
		if err != nil {
			return nil, fmt.Errorf("tools/qr: generate qrcode failed: %w", err)
		}
		return qr, nil
	}

	qr, err := qrcode.NewWithForcedVersion(text, version, level)
	if err != nil {
		return nil, fmt.Errorf("tools/qr: generate qrcode with version %d failed: %w", version, err)
	}
	return qr, nil
}

// generateWithLogo builds QR image, overlays centered logo with finder
// protection, validates effective cover ratio, and writes final PNG to output.
func generateWithLogo(text string, output io.Writer, ps *params) error {
	// Auto-version mode: start from minimal encodable version and increase only
	// until cover ratio constraints can be satisfied.
	if ps.version == 0 {
		baseQR, qerr := qrcode.New(text, ps.level)
		if qerr != nil {
			return fmt.Errorf("tools/qr: generate qrcode failed: %w", qerr)
		}

		// Try progressively larger versions to increase code area for the same logo.
		for v := baseQR.VersionNumber; v <= 40; v++ {
			// If forced version cannot encode payload, skip to next version.
			candidateQR, ferr := qrcode.NewWithForcedVersion(text, v, ps.level)
			if ferr != nil {
				continue
			}

			// Compose and check whether effective cover is acceptable with tolerance.
			merged, actualCover, merr := mergeCenterLogo(candidateQR.Image(ps.size), ps.logoImg, candidateQR.VersionNumber, ps.coverRatio)
			if merr != nil {
				return merr
			}
			// First valid version wins to keep output QR as small as possible.
			if isCoverSatisfied(actualCover, ps.coverRatio) {
				return writePNGToWriter(output, merged)
			}
		}

		// Even v40 cannot satisfy requested cover under finder-protection rules.
		return fmt.Errorf("tools/qr: unable to satisfy logo-cover %.4f with qr-version auto (max version 40)", ps.coverRatio)
	}

	// Fixed-version mode: honor caller version strictly (single attempt).
	qr, err := buildQRCode(text, ps.level, ps.version)
	if err != nil {
		return err
	}

	merged, actualCover, err := mergeCenterLogo(qr.Image(ps.size), ps.logoImg, qr.VersionNumber, ps.coverRatio)
	if err != nil {
		return err
	}
	// In fixed-version mode we cannot scale version up, so fail with max cover hint.
	if !isCoverSatisfied(actualCover, ps.coverRatio) {
		return fmt.Errorf("tools/qr: logo-cover %.4f is too large for qr-version %d (max %.4f with finder protection)", ps.coverRatio, qr.VersionNumber, actualCover)
	}

	return writePNGToWriter(output, merged)
}

// generateWithoutLogo builds plain QR image and writes it as PNG stream.
func generateWithoutLogo(text string, output io.Writer, ps *params) error {
	qr, err := buildQRCode(text, ps.level, ps.version)
	if err != nil {
		return err
	}

	return writePNGToWriter(output, qr.Image(ps.size))
}

// mergeCenterLogo overlays a centered logo onto QR image while preserving scan
// reliability by avoiding finder patterns and tracking actual covered ratio.
func mergeCenterLogo(base, logo image.Image, version int, coverRatio float64) (image.Image, float64, error) {
	// Input constraints for geometric math.
	if coverRatio <= 0 || coverRatio >= 1 {
		return nil, 0, fmt.Errorf("tools/qr: invalid cover ratio: %f", coverRatio)
	}
	if version < 1 || version > 40 {
		return nil, 0, fmt.Errorf("tools/qr: invalid qrcode version: %d", version)
	}

	// Base QR image dimensions must be valid.
	baseBounds := base.Bounds()
	baseW := baseBounds.Dx()
	baseH := baseBounds.Dy()
	if baseW == 0 || baseH == 0 {
		return nil, 0, errors.New("tools/qr: invalid qrcode image size")
	}

	// Logo image dimensions must be valid.
	logoBounds := logo.Bounds()
	logoW := logoBounds.Dx()
	logoH := logoBounds.Dy()
	if logoW == 0 || logoH == 0 {
		return nil, 0, errors.New("tools/qr: invalid logo image size")
	}

	// Reconstruct module geometry from version and rendered image size.
	// totalModuleCount = code modules + quiet zone (4 modules on each side).
	moduleCount := 17 + version*4
	totalModuleCount := moduleCount + 8
	pxPerModuleX := float64(baseW) / float64(totalModuleCount)
	pxPerModuleY := float64(baseH) / float64(totalModuleCount)
	quietZoneX := int(math.Round(pxPerModuleX * 4.0))
	quietZoneY := int(math.Round(pxPerModuleY * 4.0))

	// codeRect excludes quiet zone; cover ratio is measured on code area only.
	codeRect := image.Rect(
		baseBounds.Min.X+quietZoneX,
		baseBounds.Min.Y+quietZoneY,
		baseBounds.Max.X-quietZoneX,
		baseBounds.Max.Y-quietZoneY,
	)
	if codeRect.Dx() <= 0 || codeRect.Dy() <= 0 {
		return nil, 0, errors.New("tools/qr: invalid qrcode code area")
	}

	// Finder protection blocks reserve 8 modules around three corner finders
	// (finder 7 + separator 1) to avoid masking essential detection patterns.
	finderProtectModules := 8.0
	finderProtectX := int(math.Ceil(pxPerModuleX * finderProtectModules))
	finderProtectY := int(math.Ceil(pxPerModuleY * finderProtectModules))
	finderRects := []image.Rectangle{
		image.Rect(codeRect.Min.X, codeRect.Min.Y, codeRect.Min.X+finderProtectX, codeRect.Min.Y+finderProtectY),
		image.Rect(codeRect.Max.X-finderProtectX, codeRect.Min.Y, codeRect.Max.X, codeRect.Min.Y+finderProtectY),
		image.Rect(codeRect.Min.X, codeRect.Max.Y-finderProtectY, codeRect.Min.X+finderProtectX, codeRect.Max.Y),
	}

	// Clone base image into RGBA canvas for alpha-aware logo composition.
	baseRGBA := image.NewRGBA(baseBounds)
	draw.Draw(baseRGBA, baseBounds, base, baseBounds.Min, draw.Src)

	// Convert target cover ratio into target logo width/height while preserving
	// original logo aspect ratio.
	targetArea := float64(codeRect.Dx()*codeRect.Dy()) * coverRatio
	logoAspect := float64(logoW) / float64(logoH)
	targetW := int(math.Sqrt(targetArea * logoAspect))
	targetH := int(float64(targetW) / logoAspect)
	// Clamp dimensions into valid code area range.
	if targetW < 1 {
		targetW = 1
	}
	if targetH < 1 {
		targetH = 1
	}
	if targetW > codeRect.Dx() {
		targetW = codeRect.Dx()
		targetH = int(float64(targetW) / logoAspect)
	}
	if targetH > codeRect.Dy() {
		targetH = codeRect.Dy()
		targetW = int(float64(targetH) * logoAspect)
	}

	// If centered logo overlaps protected finder zones, shrink iteratively.
	targetW, targetH, err := shrinkForFinderSafety(targetW, targetH, codeRect, finderRects)
	if err != nil {
		return nil, 0, err
	}

	// Scale source logo to final size and blend onto center region.
	overlayRect := centeredRect(codeRect, targetW, targetH)
	scaledLogo, err := scaleLogoToTarget(logo, overlayRect.Dx(), overlayRect.Dy())
	if err != nil {
		return nil, 0, err
	}

	draw.Draw(baseRGBA, overlayRect, scaledLogo, scaledLogo.Bounds().Min, draw.Over)
	// Return actual cover ratio after all constraints (clamping/shrink) applied.
	actualCover := float64(targetW*targetH) / float64(codeRect.Dx()*codeRect.Dy())

	return baseRGBA, actualCover, nil
}

// shrinkForFinderSafety repeatedly scales down centered logo rectangle until it
// no longer overlaps any finder protection rectangle.
func shrinkForFinderSafety(targetW, targetH int, codeRect image.Rectangle, finderRects []image.Rectangle) (int, int, error) {
	for i := 0; i < 200; i++ {
		// Keep logo centered while only changing size.
		candidate := centeredRect(codeRect, targetW, targetH)
		overlapped := false
		for _, finderRect := range finderRects {
			if candidate.Overlaps(finderRect) {
				overlapped = true
				break
			}
		}
		// All finder zones are clear: current size is acceptable.
		if !overlapped {
			return targetW, targetH, nil
		}

		// No feasible centered size left.
		if targetW <= 1 || targetH <= 1 {
			break
		}
		// Shrink by 5% each iteration to converge smoothly.
		targetW = int(math.Floor(float64(targetW) * 0.95))
		targetH = int(math.Floor(float64(targetH) * 0.95))
		if targetW < 1 {
			targetW = 1
		}
		if targetH < 1 {
			targetH = 1
		}
	}

	return 0, 0, errors.New("tools/qr: unable to place centered logo without covering finder patterns")
}

// centeredRect creates a rectangle with given size centered inside bounds.
func centeredRect(bounds image.Rectangle, w, h int) image.Rectangle {
	centerX := (bounds.Min.X + bounds.Max.X) / 2
	centerY := (bounds.Min.Y + bounds.Max.Y) / 2
	left := centerX - w/2
	top := centerY - h/2
	return image.Rect(left, top, left+w, top+h)
}

// scaleLogoToTarget validates dimensions then applies nearest-neighbor scaling.
func scaleLogoToTarget(src image.Image, width, height int) (*image.RGBA, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("tools/qr: invalid scaled logo size: %dx%d", width, height)
	}

	return resizeNearest(src, width, height), nil
}

// resizeNearest scales image using nearest-neighbor sampling.
// This keeps implementation dependency-free and deterministic.
func resizeNearest(src image.Image, width, height int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	for y := 0; y < height; y++ {
		sy := srcBounds.Min.Y + y*srcH/height
		for x := 0; x < width; x++ {
			sx := srcBounds.Min.X + x*srcW/width
			dst.Set(x, y, src.At(sx, sy))
		}
	}

	return dst
}
