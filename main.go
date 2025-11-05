package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/chai2010/webp"
	"github.com/schollz/progressbar/v3"
)

type Job struct {
	InputPath string
	URL       string
	IsURL     bool
}

func main() {
	// ---- CLI flags ----
	dirFlag := flag.String("dir", "", "Path to local directory containing images")
	urlsFlag := flag.String("urls", "", "Comma-separated list of image URLs")
	qualityFlag := flag.Int("quality", 80, "Image quality (1-100)")
	workersFlag := flag.Int("workers", 8, "Number of concurrent workers")
	losslessFlag := flag.Bool("lossless", false, "Use lossless compression (better for PNGs)")
	recursiveFlag := flag.Bool("recursive", false, "Scan all subdirectories when using --dir")
	helpFlag := flag.Bool("help", false, "Show usage")

	flag.Parse()

	if *helpFlag || (*dirFlag == "" && *urlsFlag == "") {
		fmt.Println(`
WebP CLI Converter
------------------
Convert PNG/JPEG images to WebP format quickly and efficiently.

Usage:
  img-opt --dir <folder> [options]
  img-opt --urls <url1,url2,...> [options]

Options:
  --dir         Path to folder containing images
  --urls        Comma-separated URLs to images
  --quality     Quality (1–100, default 80)
  --workers     Number of concurrent workers (default 8)
  --lossless    Use lossless compression (good for PNGs)
  --recursive   Scan all subdirectories (when using --dir)
  --help        Show this help message
`)
		return
	}

	outDir, err := createOutputFolder()
	if err != nil {
		fmt.Println("Error creating output folder:", err)
		return
	}

	logFile, err := os.Create(filepath.Join(outDir, "webp-errors.log"))
	if err != nil {
		fmt.Println("Failed to create error log file:", err)
		return
	}
	defer logFile.Close()
	logger := log.New(logFile, "", log.LstdFlags)

	var jobs []Job

	// ---- Load jobs from directory ----
	if *dirFlag != "" {
		files, err := collectLocalFiles(*dirFlag, *recursiveFlag)
		if err != nil {
			fmt.Println("Error reading directory:", err)
			return
		}
		for _, f := range files {
			jobs = append(jobs, Job{InputPath: f})
		}
	}

	// ---- Load jobs from URLs ----
	if *urlsFlag != "" {
		urlList := strings.Split(*urlsFlag, ",")
		for _, u := range urlList {
			u = strings.TrimSpace(u)
			if u != "" {
				jobs = append(jobs, Job{URL: u, IsURL: true})
			}
		}
	}

	if len(jobs) == 0 {
		fmt.Println("No valid images found.")
		return
	}

	fmt.Printf("Found %d image(s). Starting concurrent conversion...\n", len(jobs))
	bar := progressbar.Default(int64(len(jobs)), "Converting")

	var converted, failed int
	var mu sync.Mutex
	jobChan := make(chan Job)
	var wg sync.WaitGroup

	for i := 0; i < *workersFlag; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobChan {
				var err error
				if job.IsURL {
					err = processURLJob(job.URL, outDir, *qualityFlag, *losslessFlag)
				} else {
					err = processFileJob(job.InputPath, outDir, *qualityFlag, *losslessFlag)
				}

				if err != nil {
					logger.Println(err)
					mu.Lock()
					failed++
					mu.Unlock()
				} else {
					mu.Lock()
					converted++
					mu.Unlock()
				}
				bar.Add(1)
			}
		}()
	}

	for _, job := range jobs {
		jobChan <- job
	}
	close(jobChan)
	wg.Wait()

	fmt.Printf("\nDone. Converted: %d, Failed: %d. Output: %s\n", converted, failed, outDir)
	fmt.Println("Error log:", filepath.Join(outDir, "webp-errors.log"))
}

func processFileJob(inputPath, outDir string, quality int, lossless bool) error {
	imgFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("open file %s: %v", inputPath, err)
	}
	defer imgFile.Close()

	img, format, err := image.Decode(imgFile)
	if err != nil {
		return fmt.Errorf("decode %s: %v", inputPath, err)
	}

	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	outPath := uniquePath(filepath.Join(outDir, base+".webp"))
	return encodeWebPImage(img, outPath, quality, lossless, format)
}

func processURLJob(u, outDir string, quality int, lossless bool) error {
	resp, err := http.Get(u)
	if err != nil {
		return fmt.Errorf("download %s: %v", u, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("invalid response %s: %s", u, resp.Status)
	}

	img, format, err := image.Decode(resp.Body)
	if err != nil {
		return fmt.Errorf("decode %s: %v", u, err)
	}

	name := path.Base(u)
	name = sanitizeFileName(strings.TrimSuffix(name, filepath.Ext(name)))
	outPath := uniquePath(filepath.Join(outDir, name+".webp"))
	return encodeWebPImage(img, outPath, quality, lossless, format)
}

func encodeWebPImage(img image.Image, outputPath string, quality int, lossless bool, format string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	// PNG → lossless by default
	opts := &webp.Options{Quality: float32(quality)}
	if lossless || strings.EqualFold(format, "png") {
		opts.Lossless = true
	}

	if err := webp.Encode(out, img, opts); err != nil {
		return fmt.Errorf("encode %s: %v", outputPath, err)
	}
	return nil
}

func collectLocalFiles(folder string, recursive bool) ([]string, error) {
	abs, err := filepath.Abs(folder)
	if err != nil {
		return nil, err
	}

	var files []string
	if recursive {
		err = filepath.WalkDir(abs, func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && isImageExt(p) {
				files = append(files, p)
			}
			return nil
		})
	} else {
		entries, err := os.ReadDir(abs)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if !e.IsDir() && isImageExt(e.Name()) {
				files = append(files, filepath.Join(abs, e.Name()))
			}
		}
	}
	return files, err
}

func createOutputFolder() (string, error) {
	dl, err := downloadsDir()
	if err != nil {
		return "", err
	}
	folder := filepath.Join(dl, "webp-"+time.Now().Format("20060102-150405"))
	return folder, os.MkdirAll(folder, 0o755)
}

func downloadsDir() (string, error) {
	if runtime.GOOS == "windows" {
		if home := os.Getenv("USERPROFILE"); home != "" {
			return filepath.Join(home, "Downloads"), nil
		}
	} else {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			return filepath.Join(home, "Downloads"), nil
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return cwd, nil
}

func isImageExt(p string) bool {
	ext := strings.ToLower(filepath.Ext(p))
	return ext == ".png" || ext == ".jpg" || ext == ".jpeg"
}

func uniquePath(p string) string {
	if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
		return p
	}
	ext := filepath.Ext(p)
	base := strings.TrimSuffix(filepath.Base(p), ext)
	dir := filepath.Dir(p)
	for i := 1; i < 1_000_000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s-%d%s", base, i, ext))
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate
		}
	}
	return filepath.Join(dir, fmt.Sprintf("%s-%d%s", base, time.Now().Unix(), ext))
}

var invalidNameRe = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "file"
	}
	name = invalidNameRe.ReplaceAllString(name, "_")
	name = strings.TrimRight(name, ". ")
	if name == "" {
		return "file"
	}
	return name
}
