<div align="center">
  <p align="center">
  <img src="logo.svg" alt="MumbleBeats Logo" width="600">
</p>
  <p><strong>The Next-Generation Music Bot for Mumble</strong></p>
  <p>
    <a href="https://github.com/Osas34091/MumbleBeats/releases/latest"><img alt="Latest Release" src="https://img.shields.io/github/v/release/Osas34091/MumbleBeats?style=flat-square&color=7c3aed"></a>
    <a href="https://github.com/Osas34091/MumbleBeats/actions"><img alt="Build Status" src="https://img.shields.io/github/actions/workflow/status/Osas34091/MumbleBeats/release.yml?style=flat-square"></a>
    <img alt="Go" src="https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go">
    <img alt="React" src="https://img.shields.io/badge/React-Vite-61DAFB?style=flat-square&logo=react">
    <img alt="Platform" src="https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey?style=flat-square">
    <img alt="License" src="https://img.shields.io/badge/license-MIT-green?style=flat-square">
  </p>
</div>

MumbleBeats is a modern, self-contained music bot for **Mumble** servers, written in Go with a React web dashboard. A single executable file — no installers, no manual dependencies.

---

## Features

- **Real-time Stereo Audio** — Opus at 48kHz via FFmpeg → Gumble, ~10ms latency
- **Reactive Web Dashboard** — Full control from the browser, synced in real-time with WebSockets
- **Live DSP Filters** — Nightcore, Bass Boost, Echo, without interrupting the song (FFmpeg hot-swap)
- **Live Radio** — Integration with `radio-browser.info` via `!radio <name>`
- **Global Volume Control** — Web slider and `!volume` command in Mumble (0–200%)
- **Web Configuration Panel** — Change server, username, admins... without touching any files
- **Auto-Setup** — Automatically downloads `yt-dlp` and `ffmpeg` on first run
<img width="600" alt="image" src="https://github.com/user-attachments/assets/3b25ffda-27ae-47bd-a953-1cb900c5ff36" />
<img width="600" alt="image" src="https://github.com/user-attachments/assets/5bbb31be-9148-4410-b5e2-d8cf8a37842a" />


---

## Installation (Under 1 minute!)

1. Go to **[Releases](https://github.com/Osas34091/MumbleBeats/releases/latest)** and download the executable for your system:
   - `mumblebeats-windows-amd64.exe` → Windows
   - `mumblebeats-linux-amd64` → Linux
   - `mumblebeats-macos-arm64` → macOS Apple Silicon (M1/M2/M3)

2. **Run it** anywhere (double click on Windows, `./mumblebeats-...` in terminal on Linux/macOS).

It will automatically create a `MumbleBeats` folder.

4. Open your browser at **`http://localhost:8080`** and configure the Mumble server from the **⚙️ Settings** tab.

> The bot will automatically download `yt-dlp` and `ffmpeg` on the first run (~50MB).
<img width="286" height="500" alt="image" src="https://github.com/user-attachments/assets/a1b8f835-a0c3-422a-b960-b96ee0907685" />


---

## Mumble Commands

| Command | Alias | Description |
|---------|-------|-------------|
| `!play <search/url>` | `!p` | Add a YouTube song |
| `!queue` | `!q` | Show the current queue |
| `!now` | `!np` | Current song with progress |
| `!skip` | `!s` | Skip the current song |
| `!pause` / `!resume` | — | Pause / Resume |
| `!stop` | — | Stop the music and clear the queue |
| `!volume <0-200>` | `!v`, `!vol` | Adjust global volume |
| `!filter <name>` | — | Filters: `nightcore`, `bassboost`, `echo`, `off` |
| `!speed <0.5-2.0>` | — | Change playback speed |
| `!seek <seconds>` | — | Skip to a specific point in the song |
| `!radio <search>` | — | Play a live radio station |
| `!playlist <name>` | — | Load a saved playlist |
| `!playlocal <file>` | — | Play a file from the `music/` folder |
| `!help` | `!h` | Show all commands |

> Control commands (skip, pause, stop...) can only be used by users in the `admins` list in the configuration.

---

## Build from Source

```bash
# Requirements: Go 1.22+, Node.js 18+

git clone https://github.com/Osas34091/MumbleBeats.git
cd MumbleBeats

# 1. Build frontend
cd web && npm ci && npm run build && cd ..

# 2. Build bot
# Windows (without console):
go build -ldflags "-H=windowsgui -s -w" -o mumblebeats.exe ./cmd/bot

# Linux / macOS:
go build -ldflags "-s -w" -o mumblebeats ./cmd/bot
```

---

## License

This project is free and open source under the MIT license. Feel free to contribute!
