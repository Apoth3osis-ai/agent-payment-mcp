/**
 * Main application component with routing
 */

import { useState } from 'react';
import Header from './components/Header';
import Settings from './routes/Settings';
import Tools from './routes/Tools';
import Install from './routes/Install';

type Tab = 'settings' | 'tools' | 'install';

export default function App() {
  const [activeTab, setActiveTab] = useState<Tab>('settings');

  return (
    <div className="app">
      <Header />

      <nav className="tabs">
        <button
          className={`tab ${activeTab === 'settings' ? 'active' : ''}`}
          onClick={() => setActiveTab('settings')}
        >
          Settings
        </button>
        <button
          className={`tab ${activeTab === 'tools' ? 'active' : ''}`}
          onClick={() => setActiveTab('tools')}
        >
          Tools
        </button>
        <button
          className={`tab ${activeTab === 'install' ? 'active' : ''}`}
          onClick={() => setActiveTab('install')}
        >
          Install
        </button>
      </nav>

      <main className="content">
        {activeTab === 'settings' && <Settings />}
        {activeTab === 'tools' && <Tools />}
        {activeTab === 'install' && <Install />}
      </main>

      <footer className="footer">
        <p>
          Agent Payment MCP &copy; {new Date().getFullYear()}
          {' | '}
          <a href="https://agentpmt.com" target="_blank" rel="noopener noreferrer">
            Website
          </a>
          {' | '}
          <a href="https://github.com/your-repo" target="_blank" rel="noopener noreferrer">
            GitHub
          </a>
        </p>
      </footer>
    </div>
  );
}
