<div align="center">
  <p align="center">
  <img src="logo.svg" alt="MumbleBeats Logo" width="600">
</p>
  <p><strong>El Bot de Música de Nueva Generación para Mumble</strong></p>
  <p>
    <img alt="Go" src="https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go">
    <img alt="React" src="https://img.shields.io/badge/React-Vite-61DAFB?style=flat-square&logo=react">
    <img alt="Platform" src="https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey?style=flat-square">
    <img alt="License" src="https://img.shields.io/badge/license-MIT-green?style=flat-square">
  </p>
</div>

MumbleBeats es un bot de música moderno y auto-contenido para servidores **Mumble**, escrito en Go con un panel de control web en React. Un solo archivo ejecutable — sin instaladores, sin dependencias manuales.

---

## Características

- **Audio Estéreo en Tiempo Real** — Opus a 48kHz vía FFmpeg → Gumble, latencia de ~10ms
- **Dashboard Web Reactivo** — Control total desde el navegador, sincronizado en tiempo real con WebSockets
- **Filtros DSP en Vivo** — Nightcore, Bass Boost, Echo, sin interrumpir la canción (hot-swap FFmpeg)
- **Radio en Directo** — Integración con `radio-browser.info` via `!radio <nombre>`
- **Control de Volumen Global** — Slider en la web y comando `!volume` en Mumble (0–200%)
- **Panel de Configuración Web** — Cambia servidor, usuario, admins... sin tocar ningún archivo
- **Auto-Setup** — Descarga `yt-dlp` y `ffmpeg` automáticamente en la primera ejecución
- **Sin Terminal Visible** — En Windows el bot corre completamente en segundo plano

---

## Instalación (¡Menos de 1 minuto!)

1. Ve a **[Releases](https://github.com/Osas34091/MumbleBeats/releases/latest)** y descarga el ejecutable para tu sistema:
   - `mumblebeats-windows-amd64.exe` → Windows
   - `mumblebeats-linux-amd64` → Linux
   - `mumblebeats-macos-arm64` → macOS Apple Silicon (M1/M2/M3)
   - `mumblebeats-macos-amd64` → macOS Intel

2. **Colócalo en una carpeta vacía** y ejecútalo (doble clic en Windows, `./mumblebeats-...` en terminal en Linux/macOS).

3. Abre tu navegador en **`http://localhost:8080`** y configura el servidor Mumble desde la pestaña **⚙️ Configuración**.

> El bot descargará `yt-dlp` y `ffmpeg` automáticamente la primera vez (~50MB).

---

## Comandos de Mumble

| Comando | Alias | Descripción |
|---------|-------|-------------|
| `!play <búsqueda/url>` | `!p` | Añade una canción de YouTube |
| `!queue` | `!q` | Muestra la cola actual |
| `!now` | `!np` | Canción actual con progreso |
| `!skip` | `!s` | Salta la canción actual |
| `!pause` / `!resume` | — | Pausa / Reanuda |
| `!stop` | — | Para la música y limpia la cola |
| `!volume <0-200>` | `!v`, `!vol` | Ajusta el volumen global |
| `!filter <nombre>` | — | Filtros: `nightcore`, `bassboost`, `echo`, `off` |
| `!speed <0.5-2.0>` | — | Cambia la velocidad de reproducción |
| `!seek <segundos>` | — | Salta a un punto de la canción |
| `!radio <búsqueda>` | — | Reproduce una radio en directo |
| `!playlist <nombre>` | — | Carga una playlist guardada |
| `!playlocal <archivo>` | — | Reproduce un archivo de la carpeta `music/` |
| `!help` | `!h` | Muestra todos los comandos |

> Los comandos de control (skip, pause, stop...) solo los pueden usar los usuarios en la lista `admins` de la configuración.

---

## Compilar desde el código fuente

```bash
# Requisitos: Go 1.22+, Node.js 18+

git clone https://github.com/Osas34091/MumbleBeats.git
cd MumbleBeats

# 1. Build del frontend
cd web && npm ci && npm run build && cd ..

# 2. Build del bot
# Windows (sin consola):
go build -ldflags "-H=windowsgui -s -w" -o mumblebeats.exe ./cmd/bot

# Linux / macOS:
go build -ldflags "-s -w" -o mumblebeats ./cmd/bot
```

---

## Licencia

Este proyecto es libre y de código abierto bajo la licencia MIT. ¡Siéntete libre de contribuir!
