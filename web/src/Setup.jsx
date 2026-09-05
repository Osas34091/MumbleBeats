import React, { useState } from 'react';
import { Settings, Save, Server, Shield, Music } from 'lucide-react';
import { useAuth } from './AuthContext';
import { useTranslation } from 'react-i18next';

function Setup() {
  const { t, i18n } = useTranslation();
  const { checkAuth } = useAuth();
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsLoading(true);
    setError(null);

    const formData = new FormData(e.target);
    const data = Object.fromEntries(formData.entries());

    if (!data.username || !data.password) {
      setError(t('setup.req_user_pass'));
      setIsLoading(false);
      return;
    }

    try {
      const res = await fetch('/api/setup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
      });
      
      const resData = await res.json();
      
      if (!res.ok) {
        setError(resData.error || "Error al configurar el bot");
      } else {
        // Setup successful, refresh auth state
        await checkAuth();
      }
    } catch (err) {
      setError(t('setup.err_conn'));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-slate-950 flex flex-col items-center justify-center p-4">
      
      <div className="mb-8 flex flex-col items-center gap-3">
        <div className="bg-primary-600 p-4 rounded-2xl text-white shadow-lg shadow-primary-500/30 animate-bounce">
          <Music size={40} />
        </div>
        <h1 className="text-3xl font-bold bg-gradient-to-r from-white to-slate-400 bg-clip-text text-transparent">
          MumbleBeats
        </h1>
        <p className="text-slate-400 font-medium">{t('setup.title')}</p>
      </div>

      <div className="glass-panel w-full max-w-2xl p-8 rounded-3xl relative overflow-hidden">
        <div className="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-primary-500 to-sky-400"></div>
        
        {error && (
          <div className="bg-rose-500/20 text-rose-300 px-4 py-3 rounded-xl border border-rose-500/30 mb-6 text-sm">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="flex flex-col gap-8">
          
          {/* Admin Account Section */}
          <section>
            <h2 className="text-lg font-bold text-slate-200 mb-4 flex items-center justify-between border-b border-slate-800 pb-2">
              <span className="flex items-center gap-2">
                <Shield size={20} className="text-primary-400" />
                {t('setup.admin_account')}
              </span>
              <select 
                name="language"
                value={i18n.language} 
                onChange={(e) => i18n.changeLanguage(e.target.value)} 
                className="bg-slate-800/50 border border-slate-700/50 rounded-lg py-1 px-2 text-sm text-white focus:outline-none focus:ring-1 focus:ring-primary-500"
              >
                <option value="en">EN</option>
                <option value="es">ES</option>
              </select>
            </h2>
            <p className="text-xs text-slate-500 mb-4">
              {t('setup.admin_desc')}
            </p>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-slate-400 mb-1">{t('dashboard.username')}</label>
                <input name="username" type="text" required placeholder="Ej: admin" className="w-full bg-slate-800/50 border border-slate-700/50 rounded-xl py-3 px-4 text-white placeholder:text-slate-600 focus:outline-none focus:ring-2 focus:ring-primary-500 transition-all" />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-400 mb-1">{t('dashboard.password')}</label>
                <input name="password" type="password" required placeholder="..." className="w-full bg-slate-800/50 border border-slate-700/50 rounded-xl py-3 px-4 text-white placeholder:text-slate-600 focus:outline-none focus:ring-2 focus:ring-primary-500 transition-all" />
              </div>
            </div>
          </section>

          {/* Mumble Server Section */}
          <section>
            <h2 className="text-lg font-bold text-slate-200 mb-4 flex items-center gap-2 border-b border-slate-800 pb-2">
              <Server size={20} className="text-primary-400" />
              {t('setup.mumble_conn')}
            </h2>
            <p className="text-xs text-slate-500 mb-4">
              {t('setup.mumble_desc')}
            </p>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-slate-400 mb-1">{t('setup.server_address')}</label>
                <input name="mumble_address" type="text" defaultValue="localhost" required className="w-full bg-slate-800/50 border border-slate-700/50 rounded-xl py-3 px-4 text-white focus:outline-none focus:ring-2 focus:ring-primary-500 transition-all" />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-400 mb-1">{t('dashboard.port')}</label>
                <input name="mumble_port" type="text" defaultValue="64738" required className="w-full bg-slate-800/50 border border-slate-700/50 rounded-xl py-3 px-4 text-white focus:outline-none focus:ring-2 focus:ring-primary-500 transition-all" />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-400 mb-1">{t('setup.bot_name')}</label>
                <input name="mumble_username" type="text" defaultValue="MumbleBeats" required className="w-full bg-slate-800/50 border border-slate-700/50 rounded-xl py-3 px-4 text-white focus:outline-none focus:ring-2 focus:ring-primary-500 transition-all" />
              </div>
              <div>
                <label className="block text-sm font-medium text-slate-400 mb-1">{t('dashboard.password')}</label>
                <input name="mumble_password" type="password" className="w-full bg-slate-800/50 border border-slate-700/50 rounded-xl py-3 px-4 text-white focus:outline-none focus:ring-2 focus:ring-primary-500 transition-all" />
              </div>
              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-slate-400 mb-1">{t('setup.channel_opt')}</label>
                <input name="mumble_channel" type="text" defaultValue="Root" className="w-full bg-slate-800/50 border border-slate-700/50 rounded-xl py-3 px-4 text-white focus:outline-none focus:ring-2 focus:ring-primary-500 transition-all" />
              </div>
            </div>
          </section>

          <button 
            type="submit" 
            disabled={isLoading} 
            className="mt-4 flex items-center justify-center gap-2 w-full py-4 bg-primary-600 hover:bg-primary-500 text-white font-bold rounded-xl transition-all shadow-xl shadow-primary-500/30 hover:shadow-primary-500/50 disabled:opacity-50 text-lg"
          >
            {isLoading ? t('setup.setting_up') : (
              <><Save size={24} /> {t('setup.submit')}</>
            )}
          </button>
        </form>
      </div>
      <p className="mt-8 text-sm text-slate-600">
        MumbleBeats v1.1.0 - Panel de Administración
      </p>
    </div>
  );
}

export default Setup;
