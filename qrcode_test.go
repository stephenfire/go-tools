package tools

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/skip2/go-qrcode"
)

func TestGenerateQRCodeWithoutLogo(t *testing.T) {
	tmp := "."
	out := filepath.Join(tmp, "plain.png")

	err := GenerateQRCode("hello world", out, QRCodeOptions{
		Level:   QRCodeRecoveryMedium,
		Version: 0,
		Size:    256,
	})
	if err != nil {
		t.Fatalf("GenerateQRCode without logo failed: %v", err)
	}

	if _, err = os.Stat(out); err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	assertPNGFileDecodable(t, out)
}

func TestGenerateQRCodeWithLogo(t *testing.T) {
	tmp := "."
	logoPath := filepath.Join(tmp, "logo.png")
	out := filepath.Join(tmp, "logo-qr.png")

	if err := writeTestLogo(logoPath); err != nil {
		t.Fatalf("write logo failed: %v", err)
	}

	err := GenerateQRCode("hello with logo", out, QRCodeOptions{
		Level:    QRCodeRecoveryLow,
		Version:  0,
		Size:     256,
		LogoPath: logoPath,
	})
	if err != nil {
		t.Fatalf("GenerateQRCode with logo failed: %v", err)
	}

	if _, err = os.Stat(out); err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	assertPNGFileDecodable(t, out)
}

func TestGenerateQRCodeLogoCoverNeedsLogo(t *testing.T) {
	err := GenerateQRCode("hello", filepath.Join(".", "x.png"), QRCodeOptions{
		Level:     QRCodeRecoveryMedium,
		Version:   0,
		Size:      256,
		LogoCover: 0.2,
	})
	if err == nil {
		t.Fatalf("expect error when logo-cover is used without logo")
	}
	if !errors.Is(err, ErrLogoCoverNeedsLogo) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateQRCodeEmptyOutput(t *testing.T) {
	err := GenerateQRCode("hello", "", QRCodeOptions{Level: QRCodeRecoveryMedium, Version: 0, Size: 256})
	if err == nil {
		t.Fatalf("expect error when output path is empty")
	}
	if !errors.Is(err, ErrOutputPathEmpty) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateQRCodeToWriterWithoutLogo(t *testing.T) {
	var out bytes.Buffer
	err := GenerateQRCodeToWriter("hello writer", &out, QRCodeOptions{
		Level:   QRCodeRecoveryMedium,
		Version: 0,
		Size:    256,
	})
	if err != nil {
		t.Fatalf("GenerateQRCodeToWriter without logo failed: %v", err)
	}

	if out.Len() == 0 {
		t.Fatalf("output should not be empty")
	}

	if _, _, err = image.Decode(bytes.NewReader(out.Bytes())); err != nil {
		t.Fatalf("output should be a decodable image: %v", err)
	}
}

func TestGenerateQRCodeToWriterNilOutput(t *testing.T) {
	err := GenerateQRCodeToWriter("hello", nil, QRCodeOptions{
		Level:   QRCodeRecoveryMedium,
		Version: 0,
		Size:    256,
	})
	if err == nil {
		t.Fatalf("expect error when output writer is nil")
	}
	if !errors.Is(err, ErrOutputWriterNil) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateQRCodeToWriterMissingText(t *testing.T) {
	var out bytes.Buffer
	err := GenerateQRCodeToWriter("", &out, QRCodeOptions{Level: QRCodeRecoveryMedium, Version: 0, Size: 256})
	if err == nil {
		t.Fatalf("expect error when text is empty")
	}
	if !errors.Is(err, ErrMissingText) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateQRCodeToWriterWithLogo(t *testing.T) {
	logoData, err := buildTestLogoPNGBytes()
	if err != nil {
		t.Fatalf("build logo bytes failed: %v", err)
	}

	var out bytes.Buffer
	err = GenerateQRCodeToWriter("hello writer with logo", &out, QRCodeOptions{
		Level:      QRCodeRecoveryLow,
		Version:    0,
		Size:       256,
		LogoReader: bytes.NewReader(logoData),
	})
	if err != nil {
		t.Fatalf("GenerateQRCodeToWriter with logo failed: %v", err)
	}

	if out.Len() == 0 {
		t.Fatalf("output should not be empty")
	}

	if _, _, err = image.Decode(bytes.NewReader(out.Bytes())); err != nil {
		t.Fatalf("output should be a decodable image: %v", err)
	}
}

func TestGenerateQRCodeToWriterWithLogoPath(t *testing.T) {
	tmp := "."
	logoPath := filepath.Join(tmp, "logo.png")
	if err := writeTestLogo(logoPath); err != nil {
		t.Fatalf("write logo failed: %v", err)
	}

	var out bytes.Buffer
	err := GenerateQRCodeToWriter("hello writer with logo path", &out, QRCodeOptions{
		Level:    QRCodeRecoveryLow,
		Version:  0,
		Size:     256,
		LogoPath: logoPath,
	})
	if err != nil {
		t.Fatalf("GenerateQRCodeToWriter with logo path failed: %v", err)
	}

	if out.Len() == 0 {
		t.Fatalf("output should not be empty")
	}
}

func TestGenerateQRCodeLogoSourceConflict(t *testing.T) {
	tmp := "."
	logoPath := filepath.Join(tmp, "logo.png")
	if err := writeTestLogo(logoPath); err != nil {
		t.Fatalf("write logo failed: %v", err)
	}
	logoData, err := os.ReadFile(logoPath)
	if err != nil {
		t.Fatalf("read logo failed: %v", err)
	}

	var out bytes.Buffer
	err = GenerateQRCodeToWriter("hello", &out, QRCodeOptions{
		Level:      QRCodeRecoveryMedium,
		Version:    0,
		Size:       256,
		LogoPath:   logoPath,
		LogoReader: bytes.NewReader(logoData),
	})
	if err == nil {
		t.Fatalf("expect error when LogoPath and LogoReader are both set")
	}
	if !errors.Is(err, ErrLogoSourceConflict) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateQRCodeToWriterLogoCoverNeedsLogo(t *testing.T) {
	var out bytes.Buffer
	err := GenerateQRCodeToWriter("hello", &out, QRCodeOptions{
		Level:     QRCodeRecoveryMedium,
		Version:   0,
		Size:      256,
		LogoCover: 0.2,
	})
	if err == nil {
		t.Fatalf("expect error when logo-cover is used without logo reader")
	}
	if !errors.Is(err, ErrLogoCoverNeedsLogo) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateQRCodeToWriterInvalidRecoveryLevel(t *testing.T) {
	var out bytes.Buffer
	err := GenerateQRCodeToWriter("hello", &out, QRCodeOptions{
		Level:   QRCodeRecoveryLevel(99),
		Version: 0,
		Size:    256,
	})
	if err == nil {
		t.Fatalf("expect error when recovery level is invalid")
	}
	if !strings.Contains(err.Error(), "invalid recovery level") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestQRCodeOptionsToParamsLogoPolicy(t *testing.T) {
	logoBytes, err := buildTestLogoPNGBytes()
	if err != nil {
		t.Fatalf("build logo bytes failed: %v", err)
	}

	ps, err := (QRCodeOptions{
		Level:      QRCodeRecoveryLow,
		Version:    0,
		Size:       256,
		LogoReader: bytes.NewReader(logoBytes),
	}).toParams()
	if err != nil {
		t.Fatalf("toParams failed: %v", err)
	}
	if ps.level != qrcode.Highest {
		t.Fatalf("logo mode should force highest recovery by default, got %v", ps.level)
	}
	if ps.coverRatio != QRCodeDefaultLogoCover {
		t.Fatalf("default logo cover mismatch, want %v got %v", QRCodeDefaultLogoCover, ps.coverRatio)
	}

	ps2, err := (QRCodeOptions{
		Level:                       QRCodeRecoveryLow,
		Version:                     0,
		Size:                        256,
		DisableForceHighestWhenLogo: true,
		LogoReader:                  bytes.NewReader(logoBytes),
	}).toParams()
	if err != nil {
		t.Fatalf("toParams failed: %v", err)
	}
	if ps2.level != qrcode.Low {
		t.Fatalf("disable force-highest should keep configured level, got %v", ps2.level)
	}
}

func TestQRCodeOptionsToParamsValidation(t *testing.T) {
	tests := []struct {
		name    string
		options QRCodeOptions
		check   func(error) bool
	}{
		{
			name:    "invalid size",
			options: QRCodeOptions{Level: QRCodeRecoveryMedium, Version: 0, Size: 0},
			check:   func(err error) bool { return err != nil && strings.Contains(err.Error(), "size") },
		},
		{
			name:    "invalid version",
			options: QRCodeOptions{Level: QRCodeRecoveryMedium, Version: 41, Size: 256},
			check:   func(err error) bool { return err != nil && strings.Contains(err.Error(), "version") },
		},
		{
			name: "logo cover requires logo",
			options: QRCodeOptions{Level: QRCodeRecoveryMedium, Version: 0, Size: 256,
				LogoCover: 0.3,
			},
			check: func(err error) bool { return errors.Is(err, ErrLogoCoverNeedsLogo) },
		},
		{
			name: "logo source conflict",
			options: QRCodeOptions{Level: QRCodeRecoveryMedium, Version: 0, Size: 256,
				LogoPath:   "/tmp/logo.png",
				LogoReader: bytes.NewReader([]byte("x")),
			},
			check: func(err error) bool { return errors.Is(err, ErrLogoSourceConflict) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.options.toParams()
			if !tt.check(err) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func writeTestLogo(path string) error {
	img := image.NewRGBA(image.Rect(0, 0, 60, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 60; x++ {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	return png.Encode(f, img)
}

func buildTestLogoPNGBytes() ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, 60, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 60; x++ {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return io.ReadAll(&buf)
}

func assertPNGFileDecodable(t *testing.T, path string) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open output file failed: %v", err)
	}
	defer func() {
		_ = f.Close()
	}()

	if _, _, err = image.Decode(f); err != nil {
		t.Fatalf("output should be a decodable image: %v", err)
	}
}
