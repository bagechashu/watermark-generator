package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

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
	width := flag.Int("width", 0, "Image width (required for generate mode)")
	height := flag.Int("height", 0, "Image height (required for generate mode)")
	fontPath := flag.String("font", "font.ttf", "Path to TTF font file")
	alpha := flag.Uint("alpha", 255, "Alpha transparency (0-255)")
	input := flag.String("input", "", "Input image path (activates apply mode)")
	output := flag.String("output", "", "Output image path (optional)")
	help := flag.Bool("help", false, "Show help message")

	flag.Parse()

	if *help {
		fmt.Println("Watermark Tool - CLI Usage")
		fmt.Println()
		fmt.Println("MODE 1: Generate new watermark")
		fmt.Println("  watermark -text <text> -width <w> -height <h> [options]")
		fmt.Println()
		fmt.Println("MODE 2: Apply watermark to existing image")
		fmt.Println("  watermark -text <text> -input <image> [options]")
		fmt.Println()
		fmt.Println("Common Flags:")
		fmt.Println("  -text <str>     Watermark text (required in both modes)")
		fmt.Println("  -font <path>    Path to TTF font file (default: font.ttf)")
		fmt.Println("  -alpha <0-255>  Alpha transparency (default: 255)")
		fmt.Println()
		fmt.Println("Generate Mode Flags:")
		fmt.Println("  -width <int>    Image width in pixels (required)")
		fmt.Println("  -height <int>   Image height in pixels (required)")
		fmt.Println()
		fmt.Println("Apply Mode Flags:")
		fmt.Println("  -input <path>   Input image path (activates apply mode)")
		fmt.Println("  -output <path>  Output image path (default: input_watermarked.png)")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  ./watermark -text 'Brand' -width 400 -height 200")
		fmt.Println("  ./watermark -text 'Logo' -input photo.jpg -alpha 200")
		fmt.Println("  ./watermark -text 'Mark' -input photo.png -output marked.png")
		return
	}

	// Validate required: text is always needed
	if *text == "" {
		fmt.Fprintf(os.Stderr, "Error: -text is required\n")
		os.Exit(1)
	}

	// Validate alpha
	if *alpha > 255 {
		fmt.Fprintf(os.Stderr, "Error: -alpha must be between 0 and 255\n")
		os.Exit(1)
	}

	// Mode detection: apply mode if -input is provided, else generate mode
	if *input != "" {
		// Apply mode: add watermark to existing image
		outputPath := *output
		if outputPath == "" {
			// Default output naming: input_watermarked.png
			ext := filepath.Ext(*input)
			base := strings.TrimSuffix(*input, ext)
			outputPath = base + "_watermarked.png"
		}

		err := applyWatermark(*input, outputPath, *text, *fontPath, uint8(*alpha))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error applying watermark: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ Watermark applied: %s\n", outputPath)
	} else {
		// Generate mode: create new watermark image
		if *width == 0 || *height == 0 {
			fmt.Fprintf(os.Stderr, "Error: -width and -height are required in generate mode\n")
			os.Exit(1)
		}

		outputPath := fmt.Sprintf("%s-%dx%d.png", *text, *width, *height)
		cfg := WatermarkConfig{
			Text:     *text,
			Width:    *width,
			Height:   *height,
			FontPath: *fontPath,
			Alpha:    uint8(*alpha),
			Output:   outputPath,
		}

		err := generate(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating %s: %v\n", cfg.Output, err)
			os.Exit(1)
		}
		fmt.Printf("✓ Generated: %s (%dx%d)\n", cfg.Output, cfg.Width, cfg.Height)
	}
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

// applyWatermark adds a tiled watermark text to an existing image
func applyWatermark(inputPath, outputPath, text, fontPath string, alpha uint8) error {
	// 1. Read input image
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input image: %w", err)
	}
	defer inputFile.Close()

	// Decode image based on file extension
	var img image.Image
	ext := strings.ToLower(filepath.Ext(inputPath))

	switch ext {
	case ".png":
		img, err = png.Decode(inputFile)
	case ".jpg", ".jpeg":
		img, err = jpeg.Decode(inputFile)
	default:
		return fmt.Errorf("unsupported image format: %s", ext)
	}

	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// 2. Convert to RGBA for drawing
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(dst, dst.Bounds(), img, bounds.Min, draw.Src)

	// 3. Read font
	fontBytes, err := os.ReadFile(fontPath)
	if err != nil {
		return fmt.Errorf("failed to read font: %w", err)
	}

	ft, err := opentype.Parse(fontBytes)
	if err != nil {
		return fmt.Errorf("failed to parse font: %w", err)
	}

	// 4. Dynamic font size (based on image height: 5% of height per line)
	fontSize := float64(height) * 0.05

	face, err := opentype.NewFace(ft, &opentype.FaceOptions{
		Size:    fontSize,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return fmt.Errorf("failed to create font face: %w", err)
	}

	// 5. White color with specified alpha
	col := color.RGBA{255, 255, 255, alpha}

	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(col),
		Face: face,
	}

	// 6. Tile watermark across entire image
	textWidth := d.MeasureString(text).Round()
	textHeight := int(fontSize)

	// Calculate spacing for even tiling
	horizSpacing := textWidth + int(fontSize)
	vertSpacing := textHeight + int(fontSize/2)

	// Draw watermark in a grid pattern across the entire image
	for y := -textHeight; y < height; y += vertSpacing {
		for x := -textWidth; x < width; x += horizSpacing {
			d.Dot = fixed.P(x, y+int(fontSize))
			d.DrawString(text)
		}
	}

	// 7. Save output image
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Encode as PNG or based on output extension
	if strings.ToLower(filepath.Ext(outputPath)) == ".jpg" || strings.ToLower(filepath.Ext(outputPath)) == ".jpeg" {
		return jpeg.Encode(outFile, dst, &jpeg.Options{Quality: 95})
	}

	return png.Encode(outFile, dst)
}
