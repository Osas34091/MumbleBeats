import React, { useState } from 'react';
import { Lock, Music, User } from 'lucide-react';
import { useAuth } from './AuthContext';

function Login() {
  const { login } = useAuth();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    setError(null);

    const formData = new FormData(e.target);
    const data = Object.fromEntries(formData.entries());

    if (!data.username || !data.password) {
      setError("Por favor, ingresa tus credenciales.");
      setIsLoading(false);
      return;
    }

    try {
      const success = await login(data.username, data.password);
      if (!success) {
        setError("Usuario o contraseña incorrectos.");
      }
    } catch (err) {
      setError("Fallo en la conexión con el servidor.");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col items-center justify-center p-4">
      
      <div className="glass-panel w-full max-w-md p-8 rounded-3xl relative overflow-hidden">
        <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-primary-500 to-sky-400"></div>
        
        <div className="flex flex-col items-center mb-8">
          <div className="bg-primary-600/20 p-4 rounded-2xl text-primary-400 mb-4 border border-primary-500/20">
            <Music size={32} />
          </div>
          <h1 className="text-2xl font-bold text-white">Iniciar Sesión</h1>
          <p className="text-sm text-slate-400 mt-1">Accede al panel de MumbleBeats</p>
        </div>

        {error && (
          <div className="bg-rose-500/20 text-rose-300 px-4 py-3 rounded-xl border border-rose-500/30 mb-6 text-sm text-center">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="flex flex-col gap-5">
          
          <div className="relative group">
            <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-500 group-focus-within:text-primary-400 transition-colors">
              <User size={18} />
            </div>
            <input 
              name="username" 
              type="text" 
              required 
              placeholder="Usuario" 
              className="w-full bg-slate-800/50 border border-slate-700/50 rounded-xl py-3 pl-11 pr-4 text-white placeholder:text-slate-600 focus:outline-none focus:ring-2 focus:ring-primary-500 transition-all" 
            />
          </div>

          <div className="relative group">
            <div className="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-slate-500 group-focus-within:text-primary-400 transition-colors">
              <Lock size={18} />
            </div>
            <input 
              name="password" 
              type="password" 
              required 
              placeholder="Contraseña" 
              className="w-full bg-slate-800/50 border border-slate-700/50 rounded-xl py-3 pl-11 pr-4 text-white placeholder:text-slate-600 focus:outline-none focus:ring-2 focus:ring-primary-500 transition-all" 
            />
          </div>

          <button 
            type="submit" 
            disabled={isLoading} 
            className="mt-2 flex items-center justify-center gap-2 w-full py-3.5 bg-primary-600 hover:bg-primary-500 text-white font-bold rounded-xl transition-all shadow-lg shadow-primary-500/30 hover:shadow-primary-500/50 disabled:opacity-50"
          >
            {isLoading ? 'Iniciando sesión...' : 'Entrar'}
          </button>
        </form>
      </div>
      
    </div>
  );
}

export default Login;
