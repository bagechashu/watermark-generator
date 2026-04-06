package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// ====== 配置 ======
type WatermarkConfig struct {
	Text     string
	Width    int
	Height   int
	FontPath string
	Alpha    uint8
	Output   string
}

// Default sizes with approximate 2.29:1 aspect ratio
var defaultSizes = map[string][2]int{
	"small":  {50, 34},
	"medium": {237, 105},
	"large":  {407, 178},
}

// 字体文件可以使用系统自带的，或者下载免费的字体。以下示例使用 Arial Bold 字体。
// cp /System/Library/Fonts/Supplemental/Arial\ Bold.ttf font.ttf

func main() {
	// CLI flags
	text := flag.String("text", "", "Watermark text (required)")
	width := flag.Int("width", 0, "Image width (required). Use preset: small/medium/large")
	height := flag.Int("height", 0, "Image height (required)")
	fontPath := flag.String("font", "font.ttf", "Path to TTF font file")
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	if *help {
		fmt.Println("Watermark Generator - CLI Usage")
		fmt.Println()
		fmt.Println("Usage: watermark -text <text> -width <w> -height <h> [options]")
		fmt.Println()
		fmt.Println("Flags:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Default sizes (width x height):")
		fmt.Println("  small:  50 x 34")
		fmt.Println("  medium: 237 x 105")
		fmt.Println("  large:  407 x 178")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  ./watermark -text 'MyBrand' -width 407 -height 178")
		fmt.Println("  ./watermark -text 'Logo' -width 1920 -height 1080 -font custom.ttf")
		return
	}

	// Validate required parameters
	if *text == "" {
		fmt.Fprintf(os.Stderr, "Error: -text is required\n")
		os.Exit(1)
	}

	// Handle preset sizes
	if *width == 0 && *height == 0 {
		fmt.Fprintf(os.Stderr, "Error: -width and -height are required (or use preset: small/medium/large)\n")
		os.Exit(1)
	}

	// Support preset sizes
	if *width > 0 && *height == 0 {
		// Check if width is a preset name (for future extensibility)
		fmt.Fprintf(os.Stderr, "Error: both -width and -height must be specified\n")
		os.Exit(1)
	}

	if *width == 0 || *height == 0 {
		fmt.Fprintf(os.Stderr, "Error: both -width and -height must be specified\n")
		os.Exit(1)
	}

	// Generate single watermark
	output := fmt.Sprintf("%dx%d.png", *width, *height)
	cfg := WatermarkConfig{
		Text:     *text,
		Width:    *width,
		Height:   *height,
		FontPath: *fontPath,
		Alpha:    255,
		Output:   output,
	}

	err := generate(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating %s: %v\n", cfg.Output, err)
		os.Exit(1)
	}
	fmt.Printf("✓ Generated: %s (%dx%d)\n", cfg.Output, cfg.Width, cfg.Height)
}

func generate(cfg WatermarkConfig) error {
	// 1. 创建透明背景
	img := image.NewRGBA(image.Rect(0, 0, cfg.Width, cfg.Height))
	draw.Draw(img, img.Bounds(), image.Transparent, image.Point{}, draw.Src)

	// 2. 读取字体
	fontBytes, err := os.ReadFile(cfg.FontPath)
	if err != nil {
		return err
	}

	ft, err := opentype.Parse(fontBytes)
	if err != nil {
		return err
	}

	// 3. 动态字体大小（按高度比例）
	fontSize := float64(cfg.Height) * 0.4

	face, err := opentype.NewFace(ft, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return err
	}

	// 4. 白色（带透明）
	col := color.RGBA{255, 255, 255, cfg.Alpha}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: face,
	}

	// 5. 计算居中位置
	textWidth := d.MeasureString(cfg.Text).Round()
	x := (cfg.Width - textWidth) / 2
	y := (cfg.Height / 2) + int(fontSize/3)

	d.Dot = fixed.P(x, y)

	// 6. 绘制
	d.DrawString(cfg.Text)

	// 7. 输出
	out, err := os.Create(cfg.Output)
	if err != nil {
		return err
	}
	defer out.Close()

	return png.Encode(out, img)
}
