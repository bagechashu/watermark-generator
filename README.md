# Generate Watermark PNG

## Font use LxgwWenKai
- https://github.com/lxgw/LxgwWenKai/releases


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