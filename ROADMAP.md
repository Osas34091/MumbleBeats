# MumbleBeats Roadmap

This document outlines the planned features, enhancements, and upcoming versions for **MumbleBeats** to continuously improve the user experience and functionality.

## Upcoming Versions

### v1.1
- **Token Authentication for Web API:** Add a secure token-based authentication system to the web API so the dashboard can be exposed to the internet safely, rather than just running on `localhost`.

### v1.2
- **System Tray Icon (Windows):** Add a background tray icon in the Windows taskbar. This will allow the user to easily see if the bot is running, right-click to open the dashboard, or stop the bot entirely, improving the UX for background processes.

### v1.3
- **Multi-channel Support:** Allow changing the Mumble channel directly from the Web Dashboard without needing to configure it manually and restart the bot.

## Future Enhancements (Ideas & Backlog)
- **Advanced Playlist Management:** Visual drag-and-drop playlist creation in the Web UI.
- **Docker Support:** Official `docker-compose.yml` and Dockerhub images for easy server deployments on Linux.
- **Spotify Integration:** Ability to search and queue Spotify tracks (which will be matched and played via YouTube under the hood).
- **User Roles:** Distinct roles on the Web UI (Admin, DJ, Viewer) with different permissions for skipping/stopping music.
- **Audio Equalizer (EQ):** A visual equalizer on the dashboard to adjust specific frequency bands (using FFmpeg filters).
