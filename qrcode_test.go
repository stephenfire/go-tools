package tools

import (
	"bytes"
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

func TestEffectiveRecoveryLevel(t *testing.T) {
	tests := []struct {
		name                string
		input               qrcode.RecoveryLevel
		hasLogo             bool
		disableForceHighest bool
		want                qrcode.RecoveryLevel
	}{
		{name: "no logo keeps input", input: qrcode.Low, hasLogo: false, disableForceHighest: false, want: qrcode.Low},
		{name: "logo defaults to highest", input: qrcode.Low, hasLogo: true, disableForceHighest: false, want: qrcode.Highest},
		{name: "logo can disable force highest", input: qrcode.Medium, hasLogo: true, disableForceHighest: true, want: qrcode.Medium},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := effectiveRecoveryLevel(tc.input, tc.hasLogo, tc.disableForceHighest)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestGenerateQRCodeWithoutLogo(t *testing.T) {
	tmp := "."
	out := filepath.Join(tmp, "plain.png")

	err := GenerateQRCode("hello world", QRCodeRecoveryMedium, 0, 256, out, "", 0)
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

	err := GenerateQRCode("hello with logo", QRCodeRecoveryLow, 0, 256, out, logoPath, 0)
	if err != nil {
		t.Fatalf("GenerateQRCode with logo failed: %v", err)
	}

	if _, err = os.Stat(out); err != nil {
		t.Fatalf("output file not created: %v", err)
	}
}

func TestGenerateQRCodeLogoCoverNeedsLogo(t *testing.T) {
	err := GenerateQRCode("hello", QRCodeRecoveryMedium, 0, 256, filepath.Join(".", "x.png"), "", 0.2)
	if err == nil {
		t.Fatalf("expect error when logo-cover is used without logo")
	}
	if !strings.Contains(err.Error(), "logo-cover requires logo") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateQRCodeWithOptions(t *testing.T) {
	tmp := "."
	out := filepath.Join(tmp, "plain-options.png")

	err := GenerateQRCodeWithOptions(out, "", QRCodeOptions{
		Text:    "hello options",
		Level:   QRCodeRecoveryMedium,
		Version: 0,
		Size:    256,
	})
	if err != nil {
		t.Fatalf("GenerateQRCodeWithOptions failed: %v", err)
	}

	if _, err = os.Stat(out); err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	assertPNGFileDecodable(t, out)
}

func TestGenerateQRCodeToWriterWithoutLogo(t *testing.T) {
	var out bytes.Buffer
	err := GenerateQRCodeToWriter("hello writer", QRCodeRecoveryMedium, 0, 256, &out, nil, 0)
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

func TestGenerateQRCodeToWriterWithLogo(t *testing.T) {
	logoData, err := buildTestLogoPNGBytes()
	if err != nil {
		t.Fatalf("build logo bytes failed: %v", err)
	}

	var out bytes.Buffer
	err = GenerateQRCodeToWriter("hello writer with logo", QRCodeRecoveryLow, 0, 256, &out, bytes.NewReader(logoData), 0)
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

func TestGenerateQRCodeToWriterLogoCoverNeedsLogo(t *testing.T) {
	var out bytes.Buffer
	err := GenerateQRCodeToWriter("hello", QRCodeRecoveryMedium, 0, 256, &out, nil, 0.2)
	if err == nil {
		t.Fatalf("expect error when logo-cover is used without logo reader")
	}
	if !strings.Contains(err.Error(), "logo-cover requires logo") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateQRCodeToWriterWithOptions(t *testing.T) {
	var out bytes.Buffer
	err := GenerateQRCodeToWriterWithOptions(&out, nil, QRCodeOptions{
		Text:    "hello writer options",
		Level:   QRCodeRecoveryMedium,
		Version: 0,
		Size:    256,
	})
	if err != nil {
		t.Fatalf("GenerateQRCodeToWriterWithOptions failed: %v", err)
	}

	if out.Len() == 0 {
		t.Fatalf("output should not be empty")
	}

	if _, _, err = image.Decode(bytes.NewReader(out.Bytes())); err != nil {
		t.Fatalf("output should be a decodable image: %v", err)
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
