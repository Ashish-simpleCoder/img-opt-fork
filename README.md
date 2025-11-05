# 🖼️ img-opt

**img-opt** is a fast, cross-platform command-line tool built in Go for optimizing and converting images.

Currently, it focuses on **converting PNG/JPEG files to WebP**.

---

## 🚀 Features

- 🧩 Convert **local images or remote URLs** to WebP
- ⚙️ **Concurrent processing** — converts multiple images at once
- 🧭 **Recursive folder scanning** (optional)
- 💾 Automatically saves optimized images to your **Downloads** folder
- 📦 **Error logging** to a file (`webp-errors.log`)
- 🔧 Non-interactive mode — ideal for automation and CI/CD

---

## 📥 Installation

Just run this single command:

```bash
curl -sSL https://raw.githubusercontent.com/Yagnik-Gohil/img-opt/main/install.sh | bash
```

Once installed, verify:

```bash
img-opt --help
```

You’re ready to go!

---

## 🧠 Usage

```bash
img-opt [flags]
```

### Flags

| Flag | Description | Example |
|------|--------------|----------|
| `--dir` | Path to a local folder containing images | `--dir ./images` |
| `--urls` | Comma-separated list of image URLs to convert | `--urls "https://a.jpg,https://b.png"` |
| `--quality` | Output quality (1–100, default `80`) | `--quality 75` |
| `--workers` | Number of concurrent workers (default `8`) | `--workers 12` |
| `--lossless` | Use lossless compression (better for PNGs) | `--lossless` |
| `--recursive` | Include subfolders when scanning directories | `--recursive` |
| `--help` | Show help message | `--help` |

---

## 💡 Examples

### 🔹 Convert a local folder
Convert all `.png` and `.jpg` images in the current directory to WebP:

```bash
img-opt --dir .
```

Output will be saved in your `~/Downloads/webp-YYYYMMDD-HHMMSS/` folder.

---

### 🔹 Convert from URLs
Download and convert multiple images directly from the web:

```bash
img-opt --urls "https://example.com/photo1.jpg,https://example.com/photo2.png"
```

---

### 🔹 Convert recursively
Process a folder and all its subfolders:

```bash
img-opt --dir ./assets --recursive
```

---

### 🔹 Adjust quality
Lower quality to reduce file size (smaller output, slightly less sharpness):

```bash
img-opt --dir ./photos --quality 70
```

---

### 🔹 Use lossless mode (best for PNGs)
Keeps perfect visual quality but slightly larger files:

```bash
img-opt --dir ./screenshots --lossless
```

---

### 🔹 Increase performance (more concurrency)
If you have many CPU cores:

```bash
img-opt --dir ./large-set --workers 16
```

---

## 📂 Output structure

Converted files are automatically placed in your **Downloads** directory inside a timestamped folder, for example:

```
~/Downloads/webp-20251105-140205/
├── image1.webp
├── image2.webp
└── webp-errors.log
```

The `webp-errors.log` file lists any files that failed to convert.

---

## 🧰 Build manually

If you have Go installed and want to build locally:

```bash
go build -o img-opt
./img-opt --help
```

To build for multiple platforms:

```bash
GOOS=linux   GOARCH=amd64 go build -o bin/img-opt-linux
GOOS=darwin  GOARCH=arm64 go build -o bin/img-opt-macos-arm64
GOOS=windows GOARCH=amd64 go build -o bin/img-opt.exe
```
---

## 💡 Motivation
The fastest, simplest way to convert and optimize images.

## 🤖 Built with AI
AI-assisted for speed and precision.

