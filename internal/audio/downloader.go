package audio

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

// EnsureDependencies checks if yt-dlp and ffmpeg exist.
// If they don't, it downloads the correct binaries for the current OS.
func EnsureDependencies(ctx context.Context) error {
	err := ensureYtDlp(ctx)
	if err != nil {
		return fmt.Errorf("error downloading yt-dlp: %w", err)
	}

	err = ensureFFmpeg(ctx)
	if err != nil {
		return fmt.Errorf("error downloading ffmpeg: %w", err)
	}

	return nil
}

func getExeName(base string) string {
	if runtime.GOOS == "windows" {
		return base + ".exe"
	}
	return base
}

func ensureYtDlp(ctx context.Context) error {
	exeName := getExeName("yt-dlp")
	if _, err := os.Stat(exeName); err == nil {
		fmt.Printf("%s ya está instalado.\n", exeName)
		return nil // File exists
	}

	fmt.Printf("Descargando %s...\n", exeName)

	var downloadURL string
	switch runtime.GOOS {
	case "windows":
		downloadURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp.exe"
	case "linux":
		downloadURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_linux"
	case "darwin":
		downloadURL = "https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp_macos"
	default:
		return fmt.Errorf("unsupported os for yt-dlp: %s", runtime.GOOS)
	}

	return downloadFile(ctx, downloadURL, exeName)
}

func ensureFFmpeg(ctx context.Context) error {
	exeName := getExeName("ffmpeg")
	if _, err := os.Stat(exeName); err == nil {
		fmt.Printf("%s ya está instalado.\n", exeName)
		return nil
	}

	fmt.Printf("Descargando %s...\n", exeName)

	// URL simple para ffmpeg estático (usando un build genérico para win64)
	var downloadURL string
	switch runtime.GOOS {
	case "windows":
		downloadURL = "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip"
		return downloadAndExtractFFmpegWin(ctx, downloadURL, exeName)
	case "linux":
		downloadURL = "https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz"
		// TODO: Implementar extracción de tar.xz para linux
		return fmt.Errorf("auto-download para linux no implementado aún")
	default:
		return fmt.Errorf("unsupported os para auto-descarga de ffmpeg")
	}
}

func downloadFile(ctx context.Context, url string, filepath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	if runtime.GOOS != "windows" {
		err = os.Chmod(filepath, 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

func downloadAndExtractFFmpegWin(ctx context.Context, url string, exeName string) error {
	zipPath := "ffmpeg_temp.zip"
	err := downloadFile(ctx, url, zipPath)
	if err != nil {
		return err
	}
	defer os.Remove(zipPath)

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if filepath.Base(f.Name) == "ffmpeg.exe" {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			out, err := os.Create(exeName)
			if err != nil {
				return err
			}
			defer out.Close()

			_, err = io.Copy(out, rc)
			if err != nil {
				return err
			}
			fmt.Println("ffmpeg extraído exitosamente.")
			return nil
		}
	}

	return fmt.Errorf("no se encontró ffmpeg.exe dentro del zip")
}
