/**
 * Settings page for entering API credentials
 */

import { useEffect, useState } from 'react';
import { saveSecrets, loadSecrets, clearSecrets } from '../lib/store';

export default function Settings() {
  const [apiKey, setApiKey] = useState('');
  const [budgetKey, setBudgetKey] = useState('');
  const [auth, setAuth] = useState('');
  const [status, setStatus] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadSecrets()
      .then(secrets => {
        if (secrets) {
          setApiKey(secrets.apiKey);
          setBudgetKey(secrets.budgetKey);
          setAuth(secrets.auth || '');
        }
      })
      .finally(() => setLoading(false));
  }, []);

  const handleSave = async () => {
    try {
      await saveSecrets({ apiKey, budgetKey, auth: auth || undefined });
      setStatus('✅ Credentials saved securely');
      setTimeout(() => setStatus(''), 3000);
    } catch (error) {
      setStatus('❌ Failed to save credentials');
      console.error(error);
    }
  };

  const handleClear = async () => {
    if (!confirm('Are you sure you want to clear all credentials?')) {
      return;
    }

    try {
      await clearSecrets();
      setApiKey('');
      setBudgetKey('');
      setAuth('');
      setStatus('✅ Credentials cleared');
      setTimeout(() => setStatus(''), 3000);
    } catch (error) {
      setStatus('❌ Failed to clear credentials');
      console.error(error);
    }
  };

  if (loading) {
    return <div className="card">Loading...</div>;
  }

  return (
    <div className="card">
      <h2>API Credentials</h2>
      <p className="text-muted">
        Enter your Agent Payment API credentials. They are stored encrypted locally
        in your browser and never sent to any server except Agent Payment.
      </p>

      <div className="form-group">
        <label htmlFor="api-key">
          API Key <span className="required">*</span>
        </label>
        <input
          id="api-key"
          type="password"
          value={apiKey}
          onChange={(e) => setApiKey(e.target.value)}
          placeholder="x-api-key"
          className="input"
        />
      </div>

      <div className="form-group">
        <label htmlFor="budget-key">
          Budget Key <span className="required">*</span>
        </label>
        <input
          id="budget-key"
          type="password"
          value={budgetKey}
          onChange={(e) => setBudgetKey(e.target.value)}
          placeholder="x-budget-key"
          className="input"
        />
      </div>

      <div className="form-group">
        <label htmlFor="auth">
          Authorization (optional)
        </label>
        <input
          id="auth"
          type="password"
          value={auth}
          onChange={(e) => setAuth(e.target.value)}
          placeholder="Bearer ..."
          className="input"
        />
        <small className="text-muted">
          Only needed if your API requires additional authorization
        </small>
      </div>

      <div className="button-row">
        <button
          onClick={handleSave}
          className="button button-primary"
          disabled={!apiKey || !budgetKey}
        >
          Save Credentials
        </button>
        <button
          onClick={handleClear}
          className="button button-secondary"
        >
          Clear All
        </button>
      </div>

      {status && (
        <div className={`status ${status.includes('✅') ? 'success' : 'error'}`}>
          {status}
        </div>
      )}
    </div>
  );
}
