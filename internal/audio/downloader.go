package audio

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/schollz/progressbar/v3"
)

// ResolveExecutable busca el ejecutable en el PATH o en la carpeta local de forma segura
func ResolveExecutable(base string) string {
	exeName := getExeName(base)

	// Si está en el PATH, LookPath devuelve la ruta absoluta o el nombre.
	if path, err := exec.LookPath(exeName); err == nil {
		if filepath.IsAbs(path) {
			return path
		}
	}

	// Si está en la carpeta local
	if _, err := os.Stat(exeName); err == nil {
		if abs, err := filepath.Abs(exeName); err == nil {
			return abs
		}
	}

	return exeName
}

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
	
	// 1. Verificar en PATH
	if _, err := exec.LookPath(exeName); err == nil {
		fmt.Printf("✅ yt-dlp detectado en el sistema.\n")
		return nil
	}
	
	// 2. Verificar en carpeta actual
	if info, err := os.Stat(exeName); err == nil {
		if info.Size() > 5000000 { // Debe ser mayor a 5MB
			fmt.Printf("✅ yt-dlp detectado en la carpeta actual.\n")
			return nil
		}
		// Si es menor a 5MB, probablemente es un archivo corrupto de una descarga fallida
		os.Remove(exeName)
	}

	fmt.Printf("⏳ Descargando %s (Requerido para leer audio de internet)...\n", exeName)

	var downloadURL string
	switch runtime.GOOS {
	case "windows":
		downloadURL = "https://github.com/yt-dlp/yt-dlp-nightly-builds/releases/latest/download/yt-dlp.exe"
	case "linux":
		downloadURL = "https://github.com/yt-dlp/yt-dlp-nightly-builds/releases/latest/download/yt-dlp_linux"
	case "darwin":
		downloadURL = "https://github.com/yt-dlp/yt-dlp-nightly-builds/releases/latest/download/yt-dlp_macos"
	default:
		return fmt.Errorf("unsupported os for yt-dlp: %s", runtime.GOOS)
	}

	return downloadFile(ctx, downloadURL, exeName)
}

func ensureFFmpeg(ctx context.Context) error {
	exeName := getExeName("ffmpeg")
	
	// 1. Verificar en PATH (muy común que los usuarios ya lo tengan)
	if _, err := exec.LookPath(exeName); err == nil {
		fmt.Printf("✅ ffmpeg detectado en el sistema.\n")
		return nil
	}

	// 2. Verificar en carpeta actual
	if _, err := os.Stat(exeName); err == nil {
		fmt.Printf("✅ ffmpeg detectado en la carpeta actual.\n")
		return nil
	}

	fmt.Printf("⏳ Descargando %s (Esto puede tardar unos minutos)...\n", exeName)

	var downloadURL string
	switch runtime.GOOS {
	case "windows":
		downloadURL = "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip"
		return downloadAndExtractFFmpegWin(ctx, downloadURL, exeName)
	case "linux":
		downloadURL = "https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz"
		return downloadAndExtractFFmpegLinux(ctx, downloadURL, exeName)
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

	// Crear barra de progreso
	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"Descargando",
	)

	// Escribir al archivo y a la barra al mismo tiempo
	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
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

	fmt.Println("\nExtraendo ffmpeg.exe (por favor espera)...")

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
			fmt.Println("✅ ffmpeg extraído exitosamente.")
			return nil
		}
	}

	return fmt.Errorf("no se encontró ffmpeg.exe dentro del zip")
}

func downloadAndExtractFFmpegLinux(ctx context.Context, url string, exeName string) error {
	tarPath := "ffmpeg_temp.tar.xz"
	err := downloadFile(ctx, url, tarPath)
	if err != nil {
		return err
	}
	defer os.Remove(tarPath)

	fmt.Println("\nExtrayendo ffmpeg (por favor espera)...")

	// Use tar command to extract
	cmd := exec.CommandContext(ctx, "tar", "-xf", tarPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error extrayendo tar.xz: %v", err)
	}

	// The extracted folder has a dynamic name like ffmpeg-*-amd64-static
	// Find the ffmpeg binary inside
	files, err := filepath.Glob("ffmpeg-*-static/ffmpeg")
	if err != nil || len(files) == 0 {
		return fmt.Errorf("no se encontró el binario ffmpeg extraído")
	}

	// Move to current directory
	if err := os.Rename(files[0], exeName); err != nil {
		return fmt.Errorf("error moviendo binario: %v", err)
	}

	// Clean up extracted folder
	dirToClean := filepath.Dir(files[0])
	os.RemoveAll(dirToClean)

	fmt.Println("✅ ffmpeg extraído exitosamente.")
	return nil
}
