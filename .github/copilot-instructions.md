# Watermark Library - AI Assistant Instructions

## Project Overview
A Go CLI tool for generating transparent PNG watermarks with centered white text. Simplified interface with required text and dimensions parameters.

## Architecture
- **main.go**: Single file containing `WatermarkConfig` struct and `generate()` function
  - `generate(cfg WatermarkConfig)` - creates transparent image with centered white text
  - Dynamic font sizing based on image height (40% of height)
  - Uses golang.org/x/image for image manipulation and font rendering
  - Alpha is always 255 (fully opaque)

## Dependencies
- `golang.org/x/image` - image manipulation and font rendering
- Font file (font.ttf) - current default, configurable via `-font` flag

## Build & Run

### Build
```bash
go build -o watermark
```

### CLI Usage
```bash
./watermark -text 'Text' -width <w> -height <h> [-font font.ttf]
```

**Required parameters:**
- `-text` - Watermark text
- `-width` - Image width in pixels
- `-height` - Image height in pixels

**Optional parameters:**
- `-font` - Path to TTF font (default: font.ttf)

**Examples:**
```bash
# Generate large watermark
./watermark -text 'Example' -width 407 -height 178

# Generate custom size
./watermark -text 'Brand' -width 1920 -height 1080

# Use custom font
./watermark -text 'Logo' -width 400 -height 200 -font custom.ttf
```

**Default sizes for reference:**
- small: 50 × 34
- medium: 237 × 105
- large: 407 × 178

## Development Guidelines
1. **Keep it simple** - Text, width, height, font path are the only parameters
2. **Fixed transparency** - Alpha is always 255 (fully opaque)
3. **Auto output naming** - Files generated as `{width}x{height}.png`
4. **Simplified interface** - No presets or complex configuration modes
5. **Reusable `generate()` function** - Library-ready for future extraction

## Common Tasks
- **Add new size**: Just specify `-width` and `-height` at runtime
- **Adjust font sizing algorithm**: Modify fontSize calculation in `generate()` (~line 130)
- **Change font**: Pass `-font custom.ttf` or modify default in source
- **Extract library API**: Move generate() logic to separate package when ready

## Files to Know
- `go.mod` - Go 1.25.0, minimal dependencies
- `go.sum` - dependency checksums
- `font.ttf` - embedded font file (system Arial Bold)
- `main.go` - all logic currently here; 100 lines
