/**
 * Tools page for browsing and testing available tools
 */

import { useEffect, useState } from 'react';
import { fetchTools, purchaseTool, ToolRecord } from '../lib/api';
import { loadSecrets } from '../lib/store';

export default function Tools() {
  const [tools, setTools] = useState<ToolRecord[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    loadCredentialsAndFetchTools();
  }, []);

  const loadCredentialsAndFetchTools = async () => {
    try {
      console.log('[Tools] Loading credentials...');
      const credentials = await loadSecrets();

      if (!credentials) {
        setError('Please enter your API credentials in Settings first');
        setLoading(false);
        return;
      }

      console.log('[Tools] Credentials loaded, fetching tools...');
      const response = await fetchTools(credentials);
      console.log('[Tools] Received response:', response);
      setTools(response.tools);
      setLoading(false);
    } catch (err) {
      console.error('[Tools] Error:', err);
      setError(err instanceof Error ? err.message : 'Failed to fetch tools');
      setLoading(false);
    }
  };

  if (loading) {
    return <div className="card">Loading tools...</div>;
  }

  if (error) {
    return (
      <div className="card error">
        <h2>Error</h2>
        <p>{error}</p>
      </div>
    );
  }

  if (tools.length === 0) {
    return (
      <div className="card">
        <h2>No Tools Available</h2>
        <p>No tools found. Please check your API credentials in Settings.</p>
      </div>
    );
  }

  return (
    <div>
      <h2>Available Tools ({tools.length})</h2>
      <p className="text-muted">
        These tools will be available in your desktop client after installation.
      </p>

      <div className="tools-grid">
        {tools.map((tool) => (
          <ToolCard key={tool.function.name} tool={tool} />
        ))}
      </div>
    </div>
  );
}

function ToolCard({ tool }: { tool: ToolRecord }) {
  const [expanded, setExpanded] = useState(false);
  const [executing, setExecuting] = useState(false);
  const [result, setResult] = useState<any>(null);
  const [params, setParams] = useState<Record<string, string>>({});

  const schema = tool.function.parameters.properties || {};
  const required = tool.function.parameters.required || [];

  // Extract human-readable name from description (before "—")
  const extractToolName = (description: string): string => {
    const match = description.match(/^(.+?)\s*—/);
    return match ? match[1].trim() : tool.function.name;
  };

  // Get description without the name prefix
  const getCleanDescription = (description: string): string => {
    const match = description.match(/—\s*(.+)/);
    return match ? match[1].trim() : description;
  };

  const toolName = extractToolName(tool.function.description);
  const toolDescription = getCleanDescription(tool.function.description);
  const productId = tool.function.name;

  const handleExecute = async () => {
    setExecuting(true);
    setResult(null);

    try {
      const credentials = await loadSecrets();
      if (!credentials) {
        setResult({ error: 'No credentials found' });
        return;
      }

      const response = await purchaseTool(
        credentials,
        tool.function.name,
        params
      );
      setResult(response);
    } catch (err) {
      setResult({ error: err instanceof Error ? err.message : 'Unknown error' });
    } finally {
      setExecuting(false);
    }
  };

  return (
    <div className="card tool-card">
      <h3>{toolName}</h3>
      <p className="tool-description">{toolDescription}</p>

      <div className="tool-meta">
        {tool['x-pricing'] && (
          <div className="pricing">
            💰 {tool['x-pricing'].price_per_unit || tool['x-pricing'].cost || 0} {tool['x-pricing'].currency || 'USDC'}
            {tool['x-pricing'].metering === 'prepaid' && ' (prepaid)'}
          </div>
        )}
        {tool['x-prepaid-balance'] && (
          <div className="prepaid-balance">
            ✅ Prepaid: {tool['x-prepaid-balance'].uses_remaining} / {tool['x-prepaid-balance'].uses_purchased} remaining
          </div>
        )}
        <div className="product-id">
          ID: <code>{productId}</code>
        </div>
      </div>

      <button
        onClick={() => setExpanded(!expanded)}
        className="button button-small"
      >
        {expanded ? 'Hide Details' : 'Show Details'}
      </button>

      {expanded && (
        <div className="tool-details">
          <h4>Parameters</h4>
          {Object.entries(schema).map(([key, def]: [string, any]) => (
            <div key={key} className="form-group">
              <label>
                {key}
                {required.includes(key) && <span className="required">*</span>}
              </label>
              <input
                type="text"
                placeholder={def.description || key}
                onChange={(e) => setParams({ ...params, [key]: e.target.value })}
                className="input input-small"
              />
              {def.description && (
                <small className="text-muted">{def.description}</small>
              )}
            </div>
          ))}

          <button
            onClick={handleExecute}
            disabled={executing}
            className="button button-primary"
          >
            {executing ? 'Executing...' : 'Test Tool'}
          </button>

          {result && (
            <div className="result">
              <h4>Result</h4>
              <pre>{JSON.stringify(result, null, 2)}</pre>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
