import { useState, useEffect } from 'react'
import { Play, Pause, SkipForward, Square, Trash2, Search, Music, ListMusic, User, Volume2, Volume, Volume1, VolumeX, Settings, Save } from 'lucide-react'

function App() {
  const [queue, setQueue] = useState([])
  const [searchQuery, setSearchQuery] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const [isSearching, setIsSearching] = useState(false)
  const [isOnline, setIsOnline] = useState(true)

  const [nowPlaying, setNowPlaying] = useState(null)
  const [playbackState, setPlaybackState] = useState({ position: 0, is_paused: false, speed: 1.0 })
  const [playlists, setPlaylists] = useState([])
  const [playlistName, setPlaylistName] = useState('')
  const [activeTab, setActiveTab] = useState('player') // 'player' | 'settings'
  const [config, setConfig] = useState({})

  const fetchConfig = async () => {
    try {
      const res = await fetch('/api/config')
      if (res.ok) setConfig(await res.json())
    } catch(err) {}
  }

  useEffect(() => {
    const firstRun = localStorage.getItem('mumblebeats_first_run');
    if (!firstRun) {
      setActiveTab('settings');
      localStorage.setItem('mumblebeats_first_run', 'false');
    }
  }, []);

  const fetchPlaylists = async () => {
    try {
      const res = await fetch('/api/playlists')
      if (res.ok) setPlaylists(await res.json() || [])
    } catch(err) {}
  }

  const handleSavePlaylist = async (e) => {
    e.preventDefault()
    if (!playlistName.trim()) return
    setIsLoading(true)
    try {
      await fetch('/api/playlists/save', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: playlistName })
      })
      setPlaylistName('')
      fetchPlaylists()
    } catch(err) {}
    setIsLoading(false)
  }

  const handleLoadPlaylist = async (name) => {
    setIsLoading(true)
    try {
      await fetch('/api/playlists/load', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name })
      })
      fetchQueue()
    } catch(err) {}
    setIsLoading(false)
  }

  const fetchQueue = async () => {
    try {
      const res = await fetch('/api/queue')
      if (res.ok) {
        const data = await res.json()
        if (Array.isArray(data)) {
            // Old format fallback
            setQueue(data || [])
            setNowPlaying(data.length > 0 ? data[0] : null)
        } else {
            // New format
            setQueue(data.queue || [])
            setNowPlaying(data.now_playing || null)
            setPlaybackState({
                position: data.position || 0,
                is_paused: data.is_paused || false,
                speed: data.speed || 1.0
            })
        }
        setIsOnline(true)
      } else {
        setIsOnline(false)
      }
    } catch (err) {
      console.error("Error fetching queue:", err)
      setIsOnline(false)
    }
  }

  useEffect(() => {
    fetchQueue()
    fetchPlaylists()
    fetchConfig()
    
    // Configurar WebSocket
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/ws`;
    
    let ws;
    let reconnectTimer;
    
    const connectWS = () => {
      ws = new WebSocket(wsUrl);
      
      ws.onopen = () => {
        console.log('WS Conectado');
        setIsOnline(true);
      };
      
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.type === 'STATE_UPDATE') {
            setQueue(data.queue || []);
            setNowPlaying(data.now_playing || null);
            setPlaybackState({
              position: data.position || 0,
              is_paused: data.is_paused || false,
              speed: data.speed || 1.0
            });
            setIsOnline(true);
          }
        } catch (err) {
          console.error("Error parseando WS message", err);
        }
      };
      
      ws.onclose = () => {
        console.log('WS Desconectado, reconectando...');
        setIsOnline(false);
        reconnectTimer = setTimeout(connectWS, 3000);
      };
      
      ws.onerror = (err) => {
        console.error('WS Error:', err);
        ws.close();
      };
    };
    
    connectWS();
    
    return () => {
      clearTimeout(reconnectTimer);
      if (ws) ws.close();
    }
  }, [])

  const handlePlay = async (e) => {
    e.preventDefault()
    if (!searchQuery.trim()) return
    
    setIsSearching(true)
    try {
      await fetch('/api/play', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query: searchQuery })
      })
      setSearchQuery('')
      fetchQueue()
    } catch (err) {
      console.error("Error playing:", err)
    }
    setIsSearching(false)
  }

  const handleCommand = async (endpoint) => {
    setIsLoading(true)
    try {
      await fetch(`/api/${endpoint}`, { method: 'POST' })
      fetchQueue()
    } catch (err) {
      console.error(`Error executing ${endpoint}:`, err)
    }
    setIsLoading(false)
  }

  const handleRemove = async (id) => {
    setIsLoading(true)
    try {
      await fetch(`/api/queue/${id}`, { method: 'DELETE' })
      fetchQueue()
    } catch (err) {
      console.error(`Error removing track:`, err)
    }
    setIsLoading(false)
  }

  const currentTrack = nowPlaying
  const upcomingTracks = queue.filter(t => t.id !== currentTrack?.id)

  return (
    <div className="max-w-6xl mx-auto p-4 md:p-8">
      {/* Header */}
      <header className="flex items-center justify-between mb-8 glass-panel p-4 rounded-2xl">
        <div className="flex items-center gap-3">
          <div className="bg-primary-600 p-2 rounded-xl text-white shadow-lg shadow-primary-500/30">
            <Music size={24} />
          </div>
          <div>
            <h1 className="text-xl font-bold bg-gradient-to-r from-white to-slate-400 bg-clip-text text-transparent">
              MumbleBeats
            </h1>
            <p className="text-xs text-slate-400 font-medium tracking-wider uppercase">Dashboard de Control</p>
          </div>
        </div>
        
        <div className="flex gap-4 items-center">
        {isOnline ? (
          <div className="flex items-center gap-2 text-sm text-emerald-400 font-medium px-3 py-1.5 bg-emerald-400/10 rounded-full border border-emerald-400/20">
            <span className="relative flex h-2.5 w-2.5">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
              <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-emerald-500"></span>
            </span>
            Online
          </div>
        ) : (
          <div className="flex items-center gap-2 text-sm text-rose-400 font-medium px-3 py-1.5 bg-rose-400/10 rounded-full border border-rose-400/20">
            <span className="relative flex h-2.5 w-2.5">
              <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-rose-500"></span>
            </span>
            Offline
          </div>
        )}
        <button
          onClick={() => {
            if(confirm("¿Seguro que deseas apagar y desconectar el bot por completo?")) {
              fetch('/api/shutdown', { method: 'POST' });
              setIsOnline(false);
            }
          }}
          className="text-xs bg-rose-500/20 text-rose-400 hover:bg-rose-500/40 px-3 py-1.5 rounded-full border border-rose-500/30 transition-all font-bold"
          title="Desconecta el bot y cierra la consola"
        >
          Apagar Bot
        </button>
        </div>
      </header>

      {/* Navigation Tabs */}
      <div className="flex gap-4 mb-6 border-b border-slate-800 pb-2">
        <button 
          onClick={() => setActiveTab('player')}
          className={`flex items-center gap-2 px-4 py-2 font-medium transition-all ${activeTab === 'player' ? 'text-primary-400 border-b-2 border-primary-500' : 'text-slate-400 hover:text-slate-200'}`}
        >
          <Music size={18} /> Reproductor
        </button>
        <button 
          onClick={() => setActiveTab('settings')}
          className={`flex items-center gap-2 px-4 py-2 font-medium transition-all ${activeTab === 'settings' ? 'text-primary-400 border-b-2 border-primary-500' : 'text-slate-400 hover:text-slate-200'}`}
        >
          <Settings size={18} /> Configuración
        </button>
      </div>

      {activeTab === 'player' ? (
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left Column: Player & Search */}
        <div className="lg:col-span-7 flex flex-col gap-6">
          
          {/* Now Playing Widget */}
          <div className="glass-panel rounded-3xl p-6 relative overflow-hidden group">
            <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-primary-500 to-sky-400"></div>
            
            <h2 className="text-sm font-bold text-slate-400 tracking-widest uppercase mb-6 flex items-center gap-2">
              <Volume2 size={16} className={currentTrack ? "text-primary-400 animate-pulse" : ""} />
              Reproduciendo Ahora
            </h2>
            
            {currentTrack ? (
              <div className="flex flex-col md:flex-row gap-6 items-center md:items-start">
                <div className="w-48 h-48 rounded-2xl overflow-hidden shadow-2xl shrink-0 border border-slate-700/50 relative">
                  {currentTrack.thumbnail ? (
                    <img src={currentTrack.thumbnail} alt="Portada" className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500" />
                  ) : (
                    <div className="w-full h-full bg-slate-800 flex items-center justify-center text-slate-500">
                      <Music size={48} />
                    </div>
                  )}
                  <div className="absolute inset-0 ring-1 ring-inset ring-white/10 rounded-2xl pointer-events-none"></div>
                </div>
                
                <div className="flex-1 w-full text-center md:text-left">
                  <span className="inline-block px-3 py-1 text-xs font-semibold rounded-full bg-primary-500/20 text-primary-300 border border-primary-500/30 mb-3">
                    {currentTrack.type === 'youtube' ? 'YouTube' : currentTrack.type}
                  </span>
                  <h3 className="text-2xl font-bold text-white mb-2 leading-tight line-clamp-2">
                    {currentTrack.title}
                  </h3>
                  <div className="flex items-center justify-center md:justify-start gap-2 text-slate-400 text-sm mb-6">
                    <User size={14} />
                    <span>Añadido por <strong className="text-slate-200">{currentTrack.added_by}</strong></span>
                  </div>
                  
                  {/* Controls */}
                  <div className="flex flex-col gap-4 w-full mt-4">
                    <div className="flex items-center justify-center md:justify-start gap-3">
                      <button 
                        onClick={() => handleCommand('stop')}
                        disabled={isLoading}
                        className="p-3 rounded-full bg-slate-800 hover:bg-rose-500/20 text-slate-300 hover:text-rose-400 border border-slate-700 transition-all hover:scale-105 disabled:opacity-50"
                        title="Detener Reproducción"
                      >
                        <Square size={20} fill="currentColor" />
                      </button>
                      
                      <button 
                        onClick={() => {
                          const newPos = Math.max(0, playbackState.position - 15);
                          fetch('/api/seek', { method: 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify({seconds: newPos})});
                        }}
                        disabled={isLoading || playbackState.position < 15}
                        className="p-2 rounded-full bg-slate-800 hover:bg-slate-700 text-slate-300 border border-slate-700 transition-all hover:scale-105 disabled:opacity-50"
                        title="Atrasar 15s"
                      >
                        -15s
                      </button>

                      <button 
                        onClick={() => handleCommand(playbackState.is_paused ? 'resume' : 'pause')}
                        disabled={isLoading}
                        className="p-4 rounded-full bg-primary-600 hover:bg-primary-500 text-white shadow-lg shadow-primary-500/30 transition-all hover:scale-105 hover:shadow-primary-500/50 disabled:opacity-50"
                        title={playbackState.is_paused ? "Reanudar" : "Pausar"}
                      >
                        {playbackState.is_paused ? (
                          <Play size={24} fill="currentColor" />
                        ) : (
                          <Pause size={24} fill="currentColor" />
                        )}
                      </button>

                      <button 
                        onClick={() => {
                          const newPos = playbackState.position + 15;
                          fetch('/api/seek', { method: 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify({seconds: newPos})});
                        }}
                        disabled={isLoading}
                        className="p-2 rounded-full bg-slate-800 hover:bg-slate-700 text-slate-300 border border-slate-700 transition-all hover:scale-105 disabled:opacity-50"
                        title="Adelantar 15s"
                      >
                        +15s
                      </button>

                      <button 
                        onClick={() => handleCommand('skip')}
                        disabled={isLoading}
                        className="p-3 rounded-full bg-slate-800 hover:bg-slate-700 text-slate-300 border border-slate-700 transition-all hover:scale-105 disabled:opacity-50"
                        title="Saltar Canción"
                      >
                        <SkipForward size={20} fill="currentColor" />
                      </button>
                    </div>

                    {/* Advanced Controls */}
                    <div className="flex flex-wrap items-center justify-center md:justify-start gap-4 text-sm mt-2">
                      <div className="flex items-center gap-2">
                        <span className="text-slate-400 font-medium">Velocidad:</span>
                        <select 
                          className="bg-slate-800 border border-slate-700 rounded-lg text-slate-200 px-2 py-1 outline-none focus:ring-1 focus:ring-primary-500"
                          value={playbackState.speed}
                          onChange={(e) => {
                            fetch('/api/speed', { method: 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify({speed: parseFloat(e.target.value)})});
                          }}
                        >
                          <option value="0.5">0.5x</option>
                          <option value="0.75">0.75x</option>
                          <option value="1">1.0x</option>
                          <option value="1.25">1.25x</option>
                          <option value="1.5">1.5x</option>
                          <option value="2">2.0x</option>
                        </select>
                      </div>
                      
                      <div className="flex items-center gap-2">
                        <span className="text-slate-400 font-medium">Filtro DSP:</span>
                        <select 
                          className="bg-slate-800 border border-slate-700 rounded-lg text-slate-200 px-2 py-1 outline-none focus:ring-1 focus:ring-primary-500"
                          onChange={(e) => {
                            fetch('/api/filter', { method: 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify({filter: e.target.value})});
                          }}
                        >
                          <option value="off">Ninguno (Normal)</option>
                          <option value="nightcore">Nightcore</option>
                          <option value="bassboost">Bass Boost</option>
                          <option value="echo">Echo</option>
                        </select>
                      </div>
                      
                      <div className="flex items-center gap-3 bg-slate-800/50 px-3 py-1.5 rounded-xl border border-slate-700/50">
                        <button
                          onClick={() => {
                            const newVol = playbackState.volume === 0 ? 1 : 0;
                            setPlaybackState(prev => ({...prev, volume: newVol}));
                            fetch('/api/volume', { method: 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify({volume: newVol})});
                          }}
                          className="text-slate-400 hover:text-primary-400 transition-colors"
                        >
                          {playbackState.volume === 0 ? <VolumeX size={18} /> : playbackState.volume < 0.5 ? <Volume1 size={18} /> : <Volume2 size={18} />}
                        </button>
                        <div className="relative flex items-center group/vol">
                          <input 
                            type="range" 
                            min="0" max="2" step="0.05"
                            value={playbackState.volume}
                            onChange={(e) => {
                              const newVol = parseFloat(e.target.value);
                              setPlaybackState(prev => ({...prev, volume: newVol}));
                              fetch('/api/volume', { method: 'POST', headers: {'Content-Type':'application/json'}, body: JSON.stringify({volume: newVol})});
                            }}
                            className="w-24 h-1.5 bg-slate-700 rounded-lg appearance-none cursor-pointer [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-3 [&::-webkit-slider-thumb]:h-3 [&::-webkit-slider-thumb]:bg-primary-500 [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:shadow-[0_0_10px_rgba(14,165,233,0.5)] group-hover/vol:[&::-webkit-slider-thumb]:scale-125 transition-all"
                          />
                          <div 
                            className="absolute left-0 h-1.5 bg-primary-500 rounded-l-lg pointer-events-none" 
                            style={{width: `${(playbackState.volume / 2) * 100}%`}}
                          />
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            ) : (
              <div className="py-12 flex flex-col items-center justify-center text-slate-500 gap-4">
                <Music size={48} className="opacity-50 mb-2" />
                <p className="text-lg font-medium">El bot está en silencio.</p>
                <p className="text-sm">Busca una canción para empezar.</p>
              </div>
            )}
          </div>

          {/* Search Box */}
          <div className="glass-panel rounded-2xl p-2 relative">
            <form onSubmit={handlePlay} className="flex relative items-center w-full">
              <div className="absolute left-4 text-slate-400">
                <Search size={20} />
              </div>
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Busca en YouTube o pega un enlace..."
                className="w-full bg-transparent border-none py-4 pl-12 pr-32 text-white placeholder:text-slate-500 focus:outline-none focus:ring-0 text-lg"
                disabled={isSearching}
              />
              <button
                type="submit"
                disabled={isSearching || !searchQuery.trim()}
                className="absolute right-2 px-6 py-2.5 bg-white text-slate-900 font-bold rounded-xl hover:bg-slate-200 transition-colors disabled:opacity-50 flex items-center gap-2"
              >
                {isSearching ? 'Buscando...' : (
                  <>
                    <Play size={16} fill="currentColor" /> Play
                  </>
                )}
              </button>
            </form>
          </div>

          {/* Playlists */}
          <div className="glass-panel rounded-2xl p-6 relative">
            <h2 className="text-sm font-bold text-slate-400 tracking-widest uppercase mb-4 flex items-center gap-2">
              <ListMusic size={16} /> Playlists Guardadas
            </h2>
            <form onSubmit={handleSavePlaylist} className="flex gap-2 mb-4">
              <input
                type="text"
                value={playlistName}
                onChange={(e) => setPlaylistName(e.target.value)}
                placeholder="Nombre para guardar la cola actual..."
                className="flex-1 bg-slate-800/50 border border-slate-700/50 rounded-xl py-2 px-4 text-white placeholder:text-slate-500 focus:outline-none focus:ring-1 focus:ring-primary-500 text-sm"
                disabled={isLoading}
              />
              <button
                type="submit"
                disabled={isLoading || !playlistName.trim()}
                className="px-4 py-2 bg-primary-600 text-white font-semibold rounded-xl hover:bg-primary-500 transition-colors disabled:opacity-50 text-sm whitespace-nowrap"
              >
                Guardar
              </button>
            </form>
            <div className="flex flex-wrap gap-2">
              {playlists.length === 0 ? (
                <div className="text-sm text-slate-500 py-2">No tienes playlists guardadas.</div>
              ) : (
                playlists.map(name => (
                  <button
                    key={name}
                    onClick={() => handleLoadPlaylist(name)}
                    disabled={isLoading}
                    className="px-3 py-1.5 bg-slate-800 hover:bg-slate-700 border border-slate-700 rounded-lg text-sm text-slate-300 transition-colors flex items-center gap-2 disabled:opacity-50"
                  >
                    <ListMusic size={14} /> {name}
                  </button>
                ))
              )}
            </div>
          </div>
          
          {/* Upload MP3 */}
          <div className="glass-panel rounded-2xl p-6 relative mt-6 border border-primary-500/30 shadow-[0_0_15px_rgba(14,165,233,0.1)]">
            <h2 className="text-sm font-bold text-slate-300 tracking-widest uppercase mb-4 flex items-center gap-2">
              <div className="bg-primary-500/20 p-1.5 rounded-lg text-primary-400">
                <ListMusic size={16} />
              </div>
              Subir MP3 Local
            </h2>
            <div className="flex gap-2 mb-2">
              <input
                type="file"
                accept=".mp3,.m4a,.wav"
                id="file-upload"
                className="hidden"
                disabled={isLoading}
                onChange={async (e) => {
                  const file = e.target.files?.[0];
                  if (!file) return;
                  
                  // Optimistic state
                  const prevLabel = document.getElementById('upload-label-text').innerText;
                  document.getElementById('upload-label-text').innerText = 'Subiendo ' + file.name + '...';
                  document.getElementById('upload-label').classList.add('animate-pulse');
                  
                  const formData = new FormData();
                  formData.append('file', file);
                  try {
                    await fetch('/api/upload', {
                      method: 'POST',
                      body: formData
                    });
                    e.target.value = '';
                    document.getElementById('upload-label-text').innerText = '¡Subido con éxito!';
                    setTimeout(() => {
                      document.getElementById('upload-label-text').innerText = 'Haz clic para seleccionar otro archivo (.mp3, .m4a)';
                    }, 3000);
                  } catch (err) {
                    console.error('Error uploading:', err);
                    document.getElementById('upload-label-text').innerText = 'Error al subir';
                  } finally {
                    document.getElementById('upload-label').classList.remove('animate-pulse');
                  }
                }}
              />
              <label 
                id="upload-label"
                htmlFor="file-upload" 
                className={`flex-1 bg-primary-900/20 border-2 border-dashed border-primary-500/40 hover:border-primary-400 rounded-xl py-6 px-4 text-center cursor-pointer hover:bg-primary-900/40 transition-all text-sm group ${isLoading ? 'opacity-50 pointer-events-none' : ''}`}
              >
                <div className="flex flex-col items-center justify-center gap-2">
                  <div className="p-3 rounded-full bg-primary-500/20 text-primary-400 group-hover:scale-110 group-hover:bg-primary-500 group-hover:text-white transition-all">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="17 8 12 3 7 8"></polyline><line x1="12" y1="3" x2="12" y2="15"></line></svg>
                  </div>
                  <span id="upload-label-text" className="text-primary-200 font-medium">Haz clic aquí para seleccionar un archivo (.mp3, .m4a)</span>
                </div>
              </label>
            </div>
            <p className="text-xs text-slate-400 text-center mt-3">Una vez subido, búscalo por su nombre en la barra de búsqueda de arriba.</p>
          </div>
          
        </div>

        {/* Right Column: Queue */}
        <div className="lg:col-span-5 flex flex-col gap-6 h-full">
          <div className="glass-panel rounded-3xl p-6 flex-1 flex flex-col min-h-[500px]">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-sm font-bold text-slate-400 tracking-widest uppercase flex items-center gap-2">
                <ListMusic size={16} />
                Cola de Reproducción ({upcomingTracks.length})
              </h2>
              {upcomingTracks.length > 0 && (
                <button
                  onClick={() => handleCommand('clear')}
                  disabled={isLoading}
                  className="text-xs flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-slate-800/50 text-slate-400 hover:text-rose-400 hover:bg-rose-500/10 transition-colors border border-transparent hover:border-rose-500/20"
                >
                  <Trash2 size={14} />
                  Limpiar
                </button>
              )}
            </div>

            <div className="flex-1 overflow-y-auto pr-2 custom-scrollbar space-y-3">
              {upcomingTracks.length === 0 ? (
                <div className="h-full flex flex-col items-center justify-center text-slate-500 gap-3">
                  <ListMusic size={32} className="opacity-50" />
                  <p className="text-sm font-medium">No hay más canciones en la cola.</p>
                </div>
              ) : (
                upcomingTracks.map((track, idx) => (
                  <div key={track.id} className="group flex items-center gap-4 p-3 rounded-xl hover:bg-slate-800/50 transition-colors border border-transparent hover:border-slate-700/50">
                    <div className="text-slate-500 font-bold w-4 text-right text-sm group-hover:hidden">
                      {idx + 1}
                    </div>
                    <div className="hidden group-hover:flex flex-col gap-1 w-4 items-center justify-center">
                      <button 
                        onClick={() => handleCommand(`queue/${track.id}/up`)}
                        className="text-slate-400 hover:text-white"
                        title="Subir"
                        disabled={idx === 0}
                      >
                        <span className="text-[10px]">▲</span>
                      </button>
                      <button 
                        onClick={() => handleCommand(`queue/${track.id}/down`)}
                        className="text-slate-400 hover:text-white"
                        title="Bajar"
                        disabled={idx === upcomingTracks.length - 1}
                      >
                        <span className="text-[10px]">▼</span>
                      </button>
                    </div>
                    <div className="w-12 h-12 rounded-lg overflow-hidden shrink-0 bg-slate-800 border border-slate-700/50 relative">
                      {track.thumbnail ? (
                        <img src={track.thumbnail} alt="" className="w-full h-full object-cover" />
                      ) : (
                        <Music size={20} className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-slate-600" />
                      )}
                    </div>
                    <div className="flex-1 min-w-0">
                      <h4 className="text-sm font-medium text-slate-200 truncate group-hover:text-white transition-colors">
                        {track.title}
                      </h4>
                      <p className="text-xs text-slate-500 truncate mt-0.5">
                        Añadido por {track.added_by}
                      </p>
                    </div>
                    <button 
                      onClick={() => handleRemove(track.id)}
                      className="hidden group-hover:flex w-8 h-8 items-center justify-center text-slate-500 hover:text-rose-400 rounded-lg hover:bg-rose-500/10 ml-auto"
                      title="Eliminar de la cola"
                    >
                      <Trash2 size={16} />
                    </button>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      </div>
      ) : (
      <div className="glass-panel p-6 rounded-2xl max-w-2xl mx-auto">
        <h2 className="text-xl font-bold text-white mb-6 flex items-center gap-2">
          <Settings size={24} className="text-primary-400" />
          Ajustes del Servidor
        </h2>
        <form onSubmit={async (e) => {
          e.preventDefault()
          setIsLoading(true)
          try {
            const req = {
              server_address: e.target.server_address.value,
              server_port: e.target.server_port.value,
              username: e.target.username.value,
              password: e.target.password.value,
              channel: e.target.channel.value,
              insecure: e.target.insecure.checked,
              admins: e.target.admins.value.split(',').map(s => s.trim()).filter(Boolean)
            }
            await fetch('/api/config', { method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(req)})
            alert("Configuración guardada.")
            setActiveTab('player')
          } catch(err) {
            alert("Error al guardar.")
          }
          setIsLoading(false)
        }} className="flex flex-col gap-4">
          
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-slate-400 mb-1">Dirección del Servidor</label>
              <input name="server_address" defaultValue={config.server_address || ''} className="w-full bg-slate-800/50 border border-slate-700/50 rounded-xl py-2 px-4 text-white placeholder:text-slate-500 focus:outline-none focus:ring-1 focus:ring-primary-500" />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-400 mb-1">Puerto</label>
              <input name="server_port" defaultValue={config.server_port || ''} className="w-full bg-slate-800/50 border border-slate-700/50 rounded-xl py-2 px-4 text-white placeholder:text-slate-500 focus:outline-none focus:ring-1 focus:ring-primary-500" />
            </div>
          </div>
          
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-slate-400 mb-1">Nombre de Usuario</label>
              <input name="username" defaultValue={config.username || ''} className="w-full bg-slate-800/50 border border-slate-700/50 rounded-xl py-2 px-4 text-white placeholder:text-slate-500 focus:outline-none focus:ring-1 focus:ring-primary-500" />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-400 mb-1">Contraseña</label>
              <input name="password" type="password" placeholder="Dejar en blanco para no cambiar" className="w-full bg-slate-800/50 border border-slate-700/50 rounded-xl py-2 px-4 text-white placeholder:text-slate-500 focus:outline-none focus:ring-1 focus:ring-primary-500" />
            </div>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-slate-400 mb-1">Canal (Por Defecto)</label>
            <input name="channel" defaultValue={config.channel || ''} className="w-full bg-slate-800/50 border border-slate-700/50 rounded-xl py-2 px-4 text-white placeholder:text-slate-500 focus:outline-none focus:ring-1 focus:ring-primary-500" />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-slate-400 mb-1">Admins (separados por coma)</label>
            <input name="admins" defaultValue={(config.admins || []).join(', ')} className="w-full bg-slate-800/50 border border-slate-700/50 rounded-xl py-2 px-4 text-white placeholder:text-slate-500 focus:outline-none focus:ring-1 focus:ring-primary-500" />
          </div>
          
          <div className="flex items-center gap-2 mt-2">
            <input name="insecure" id="insecure" type="checkbox" defaultChecked={config.insecure} className="w-4 h-4 rounded bg-slate-800 border-slate-700 text-primary-500 focus:ring-primary-500" />
            <label htmlFor="insecure" className="text-sm font-medium text-slate-300 cursor-pointer">
              Permitir conexiones inseguras (Insecure TLS) - Ignora errores de certificado
            </label>
          </div>
          
          <button type="submit" disabled={isLoading} className="mt-4 flex items-center justify-center gap-2 w-full py-3 bg-primary-600 hover:bg-primary-500 text-white font-bold rounded-xl transition-all shadow-lg shadow-primary-500/30 disabled:opacity-50">
            <Save size={20} /> Guardar Cambios
          </button>
        </form>
      </div>
      )}
      
      <style dangerouslySetInnerHTML={{__html: `
        .custom-scrollbar::-webkit-scrollbar {
          width: 6px;
        }
        .custom-scrollbar::-webkit-scrollbar-track {
          background: rgba(30, 41, 59, 0.5);
          border-radius: 8px;
        }
        .custom-scrollbar::-webkit-scrollbar-thumb {
          background: rgba(71, 85, 105, 0.8);
          border-radius: 8px;
        }
        .custom-scrollbar::-webkit-scrollbar-thumb:hover {
          background: rgba(100, 116, 139, 1);
        }
      `}} />
    </div>
  )
}

export default App
