<div align="center">
  <h1>🎵 MumbleBeats</h1>
  <p><strong>El Bot de Música de Nueva Generación para Mumble (Baja Latencia + Estéreo Real)</strong></p>
</div>

MumbleBeats es un bot de música moderno, modular y reactivo para servidores Mumble, escrito completamente en Go y React. Transforma tu servidor en un Spotify colaborativo con filtros de audio en vivo, reproducción estéreo y un dashboard de control web espectacular.

---

## ✨ Características Principales

* **🎧 Estéreo Real y Baja Latencia**: Procesamiento nativo en estéreo a través del códec Opus con una tasa perfecta de sincronización de frames (20ms) gracias al motor de inyección directa desde FFmpeg a Gumble.
* **⚡ Dashboard en Tiempo Real (Web UI)**: Controla la música desde tu navegador con una interfaz hermosa, reactiva y fluida sincronizada vía WebSockets.
* **🎛️ Filtros DSP y Control Pro**: ¿Quieres escuchar música en modo *Nightcore*? ¿Aumentar el *Bass*? ¿Cambiar la velocidad a 1.25x? Hazlo todo en vivo sin detener la canción, gracias al "hot-swapping" de FFmpeg.
* **📻 Radio Inteligente**: Integración nativa con `radio-browser.info`. Usa `!radio <nombre>` para buscar estaciones de todo el mundo.
* **🤖 Auto-Setup Inteligente**: No necesitas instalar `yt-dlp` ni dependencias pesadas manualmente. El bot detecta tu sistema operativo (Windows, Mac o Linux) y los descarga automáticamente en su primera ejecución.

## 🚀 Instalación y Uso (¡Facilísimo!)

Gracias a la compilación en Go, **no necesitas instalar Node.js, ni Python, ni compilar nada**.

1. Ve a la sección de [Releases](https://github.com/tu-usuario/MumbleBeats/releases) de este repositorio.
2. Descarga el ejecutable para tu sistema operativo (`.exe` para Windows, o el binario de Linux/Mac).
3. Colócalo en una carpeta y **haz doble clic** (o ejecútalo en la terminal).

El bot descargará automáticamente sus dependencias de audio la primera vez, se conectará al servidor de Mumble por defecto y levantará el Panel de Control Web.

> **Nota:** Por defecto, el panel web estará disponible en `http://localhost:8080`.

## 💬 Comandos Disponibles en Mumble

MumbleBeats incluye un sistema de comandos súper intuitivo. Sólo escribe en el chat de tu canal:

* `!play <búsqueda/url>` (o `!p`): Añade una canción de YouTube. Si pones una playlist de YT entera, ¡las añade todas!
* `!queue` (o `!q`): Muestra la cola actual.
* `!skip` / `!pause` / `!resume` / `!stop` / `!clear`: Controles básicos.
* `!seek <segundos>`: Salta a un segundo específico de la canción (Ej: `!seek 60`).
* `!speed <valor>`: Cambia la velocidad (Ej: `!speed 1.25`).
* `!filter <nombre>`: Aplica filtros de audio (`nightcore`, `bassboost`, `echo`, `off`).
* `!lyrics` (o `!letra`): Muestra la letra de la canción que está sonando.
* `!radio <búsqueda/url>`: Busca y reproduce estaciones de radio en vivo.
* `!playlist <nombre>`: Carga una playlist que hayas guardado previamente en el Dashboard.
* `!help`: Muestra la lista completa de comandos en vivo.

## 🛠️ Para Desarrolladores

Si quieres compilarlo desde cero o modificar el código fuente:

```bash
# 1. Clona el repositorio
git clone https://github.com/tu-usuario/MumbleBeats.git
cd MumbleBeats

# 2. Construye la aplicación Web (Requiere Node.js)
cd web
npm install
npm run build
cd ..

# 3. Compila el Bot (Requiere Go)
go build -o bot.exe ./cmd/bot

# 4. Ejecuta el Bot
./bot.exe
```

## 📜 Licencia
Este proyecto es libre y de código abierto. ¡Siéntete libre de contribuir!
