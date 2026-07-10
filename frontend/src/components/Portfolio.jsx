import React, { useEffect, useState, useCallback } from 'react';

const API = import.meta.env.VITE_API_URL || '';

const fmt = (n) =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(n);

const fmtPct = (n) => `${n >= 0 ? '+' : ''}${n.toFixed(2)}%`;

const DEFAULT_PRICE_CFG = { source: 'yahoo', interval_seconds: 300, api_key: '', enabled: false };

function fmtLastUpdated(iso) {
  if (!iso) return '—';
  try {
    return new Date(iso).toLocaleString();
  } catch {
    return iso;
  }
}

export default function Portfolio() {
  const [tab, setTab]             = useState('holdings'); // 'holdings' | 'settings'
  const [holdings, setHoldings]   = useState([]);
  const [accounts, setAccounts]   = useState([]);
  const [accountFilter, setAccountFilter] = useState('');
  const [loading, setLoading]     = useState(true);
  const [error, setError]         = useState('');
  // inline price editing per symbol
  const [editingPrice, setEditingPrice] = useState(null);
  const [newPrice, setNewPrice]         = useState('');
  // price poller config
  const [priceCfg, setPriceCfg]         = useState(DEFAULT_PRICE_CFG);
  const [priceCfgForm, setPriceCfgForm] = useState(DEFAULT_PRICE_CFG);
  const [savingCfg, setSavingCfg]       = useState(false);
  const [refreshing, setRefreshing]     = useState(false);
  const [refreshResult, setRefreshResult] = useState(null);

  const loadHoldings = useCallback(() => {
    const url = accountFilter
      ? `${API}/api/portfolio/holdings?account_id=${accountFilter}`
      : `${API}/api/portfolio/holdings`;
    return fetch(url).then(r => r.json()).then(data => setHoldings(data || []));
  }, [accountFilter]);

  const loadAccounts = useCallback(() =>
    fetch(`${API}/api/accounts`).then(r => r.json()).then(data => setAccounts(data || [])), []);

  const loadPriceCfg = useCallback(() =>
    fetch(`${API}/api/price-config`)
      .then(r => r.json())
      .then(cfg => { setPriceCfg(cfg); setPriceCfgForm(cfg); }), []);

  const loadAll = useCallback(() => {
    setLoading(true);
    Promise.all([loadHoldings(), loadPriceCfg(), loadAccounts()])
      .catch(() => setError('Failed to load portfolio.'))
      .finally(() => setLoading(false));
  }, [loadHoldings, loadPriceCfg, loadAccounts]);

  useEffect(loadAll, [loadAll]);

  useEffect(() => {
    loadHoldings();
  }, [accountFilter, loadHoldings]);

  const handleSavePriceConfig = async (e) => {
    e.preventDefault();
    setSavingCfg(true);
    try {
      const payload = {
        ...priceCfgForm,
        interval_seconds: parseInt(priceCfgForm.interval_seconds, 10) || 300,
      };
      const res = await fetch(`${API}/api/price-config`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (!res.ok) throw new Error();
      const saved = await res.json();
      setPriceCfg(saved);
      setPriceCfgForm(saved);
      alert('Price settings saved.');
    } catch {
      alert('Failed to save price settings.');
    } finally {
      setSavingCfg(false);
    }
  };

  const handleManualRefresh = async () => {
    setRefreshing(true);
    setRefreshResult(null);
    try {
      const res = await fetch(`${API}/api/price-refresh`, { method: 'POST' });
      const data = await res.json();
      setRefreshResult(data);
      loadHoldings();
    } catch {
      alert('Refresh failed.');
    } finally {
      setRefreshing(false);
    }
  };

  const handleUpdatePrice = async (symbol) => {
    const price = parseFloat(newPrice);
    if (!price || price <= 0) return;
    await fetch(`${API}/api/portfolio/holdings/${encodeURIComponent(symbol)}/price`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ price }),
    });
    setEditingPrice(null);
    setNewPrice('');
    loadHoldings();
  };

  const totalValue    = holdings.reduce((s, h) => s + h.value, 0);
  const totalCost     = holdings.reduce((s, h) => s + h.total_cost, 0);
  const totalGainLoss = totalValue - totalCost;

  if (loading) return <div className="loading">Loading portfolio…</div>;
  if (error)   return <div className="error">{error}</div>;

  return (
    <div>
      <h1 className="page-title">Portfolio</h1>

      {/* Summary bar */}
      <div className="summary-grid" style={{ marginBottom: '1.5rem' }}>
        <div className="summary-card info">
          <div className="card-title">Total Value</div>
          <div className="card-value positive">{fmt(totalValue)}</div>
        </div>
        <div className="summary-card">
          <div className="card-title">Total Cost</div>
          <div className="card-value">{fmt(totalCost)}</div>
        </div>
        <div className={`summary-card ${totalGainLoss >= 0 ? '' : 'danger'}`}>
          <div className="card-title">Total Gain / Loss</div>
          <div className={`card-value ${totalGainLoss >= 0 ? 'positive' : 'negative'}`}>
            {fmt(totalGainLoss)}
          </div>
        </div>
        <div className="summary-card purple">
          <div className="card-title">Holdings</div>
          <div className="card-value">{holdings.length}</div>
        </div>
      </div>

      {/* Account filter */}
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '1rem' }}>
        <label style={{ fontWeight: 600 }}>Account Filter:</label>
        <select
          value={accountFilter}
          onChange={e => setAccountFilter(e.target.value)}
          style={{ padding: '0.4rem 0.6rem', borderRadius: '4px', border: '1px solid #cbd5e0' }}
        >
          <option value="">All accounts</option>
          {accounts.map(a => (
            <option key={a.id} value={a.id}>{a.name}</option>
          ))}
        </select>
      </div>

      {/* Tab switcher */}
      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
        <button
          className={`btn ${tab === 'holdings' ? 'btn-primary' : 'btn-secondary'}`}
          onClick={() => setTab('holdings')}
        >📊 Holdings</button>
        <button
          className={`btn ${tab === 'settings' ? 'btn-primary' : 'btn-secondary'}`}
          onClick={() => setTab('settings')}
        >⚙️ Price Settings</button>
      </div>

      {/* ── Holdings tab ── */}
      {tab === 'holdings' && (
        <div className="table-container">
          <div className="table-header">
            <h2>Current Holdings</h2>
            <span style={{ fontSize: '0.88rem', color: '#718096' }}>
              Add investment transactions in the 💳 Transactions tab
            </span>
          </div>
          {holdings.length === 0 ? (
            <div className="empty">
              No holdings yet. Add a <strong>Stock Buy</strong> transaction in the Transactions tab to get started.
            </div>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>Symbol</th>
                  <th>Name</th>
                  <th>Shares</th>
                  <th>Avg Cost</th>
                  <th>Current Price</th>
                  <th>Value</th>
                  <th>Gain / Loss</th>
                  <th style={{ color: '#718096', fontWeight: 'normal', fontSize: '0.82rem' }}>Price Updated</th>
                </tr>
              </thead>
              <tbody>
                {holdings.map(h => (
                  <tr key={h.symbol}>
                    <td><strong>{h.symbol}</strong></td>
                    <td>{h.name}</td>
                    <td>{h.quantity.toFixed(4).replace(/\.?0+$/, '')}</td>
                    <td>{fmt(h.avg_cost)}</td>
                    <td>
                      {editingPrice === h.symbol ? (
                        <span style={{ display: 'inline-flex', gap: '0.3rem', alignItems: 'center' }}>
                          <input
                            type="number" step="0.01" min="0"
                            value={newPrice}
                            onChange={e => setNewPrice(e.target.value)}
                            style={{ width: '90px', padding: '0.25rem 0.4rem', border: '1px solid #cbd5e0', borderRadius: '4px', fontSize: '0.88rem' }}
                            autoFocus
                          />
                          <button className="btn btn-primary" style={{ padding: '0.25rem 0.5rem', fontSize: '0.78rem' }} onClick={() => handleUpdatePrice(h.symbol)}>✓</button>
                          <button className="btn btn-secondary" style={{ padding: '0.25rem 0.5rem', fontSize: '0.78rem' }} onClick={() => { setEditingPrice(null); setNewPrice(''); }}>✕</button>
                        </span>
                      ) : (
                        <span
                          style={{ cursor: 'pointer', borderBottom: '1px dashed #a0aec0' }}
                          title="Click to update price manually"
                          onClick={() => { setEditingPrice(h.symbol); setNewPrice(String(h.current_price)); }}
                        >
                          {fmt(h.current_price)}
                        </span>
                      )}
                    </td>
                    <td className="amount-positive">{fmt(h.value)}</td>
                    <td className={h.gain_loss >= 0 ? 'gain' : 'loss'}>
                      {fmt(h.gain_loss)}{' '}
                      <span style={{ fontSize: '0.82rem' }}>({fmtPct(h.gain_loss_pct)})</span>
                    </td>
                    <td style={{ color: '#a0aec0', fontSize: '0.78rem', whiteSpace: 'nowrap' }}>
                      {fmtLastUpdated(h.last_updated)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {/* ── Price Settings tab ── */}
      {tab === 'settings' && (
        <div className="table-container" style={{ maxWidth: '600px' }}>
          <div className="table-header">
            <h2>⚙️ Price Polling Settings</h2>
          </div>

          <div style={{ padding: '1.25rem' }}>
            <div style={{
              background: priceCfg.enabled ? '#f0fff4' : '#fffaf0',
              border: `1px solid ${priceCfg.enabled ? '#9ae6b4' : '#fbd38d'}`,
              borderRadius: '6px',
              padding: '0.75rem 1rem',
              marginBottom: '1.5rem',
              fontSize: '0.9rem',
            }}>
              <strong>Status:</strong>{' '}
              {priceCfg.enabled
                ? `🟢 Polling ${priceCfg.source === 'yahoo' ? 'Yahoo Finance' : 'Alpha Vantage'} every ${priceCfg.interval_seconds}s`
                : '🟡 Polling disabled — prices updated manually or via Refresh button'}
            </div>

            <form onSubmit={handleSavePriceConfig}>
              <div className="form-group">
                <label><strong>Price Source</strong></label>
                <select
                  value={priceCfgForm.source}
                  onChange={e => setPriceCfgForm(f => ({ ...f, source: e.target.value }))}
                >
                  <option value="yahoo">Yahoo Finance (no API key needed)</option>
                  <option value="alphavantage">Alpha Vantage (API key required, free tier available)</option>
                </select>
                <small style={{ color: '#718096', marginTop: '0.3rem', display: 'block' }}>
                  Yahoo Finance is recommended for most users. Alpha Vantage free tier allows 25 requests/day total — with multiple holdings, polling may exhaust the daily quota quickly.
                </small>
              </div>

              {priceCfgForm.source === 'alphavantage' && (
                <div className="form-group">
                  <label><strong>Alpha Vantage API Key</strong></label>
                  <input
                    type="password"
                    value={priceCfgForm.api_key}
                    onChange={e => setPriceCfgForm(f => ({ ...f, api_key: e.target.value }))}
                    placeholder="Enter your API key from alphavantage.co"
                  />
                  <small style={{ color: '#718096', marginTop: '0.3rem', display: 'block' }}>
                    Get a free API key at <a href="https://www.alphavantage.co/support/#api-key" target="_blank" rel="noopener noreferrer">alphavantage.co</a>
                  </small>
                </div>
              )}

              <div className="form-group">
                <label><strong>Poll Interval (seconds)</strong></label>
                <input
                  type="number" min="60" max="86400" step="30"
                  value={priceCfgForm.interval_seconds}
                  onChange={e => setPriceCfgForm(f => ({ ...f, interval_seconds: e.target.value }))}
                />
                <small style={{ color: '#718096', marginTop: '0.3rem', display: 'block' }}>
                  Minimum 60 seconds. 300 = 5 min, 3600 = 1 hour. Yahoo Finance rate-limits aggressive polling.
                </small>
              </div>

              <div className="form-group" style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                <input
                  type="checkbox"
                  id="polling-enabled"
                  checked={priceCfgForm.enabled}
                  onChange={e => setPriceCfgForm(f => ({ ...f, enabled: e.target.checked }))}
                  style={{ width: '18px', height: '18px', cursor: 'pointer' }}
                />
                <label htmlFor="polling-enabled" style={{ margin: 0, cursor: 'pointer' }}>
                  <strong>Enable automatic background polling</strong>
                </label>
              </div>

              <div style={{ display: 'flex', gap: '0.75rem', marginTop: '1rem', flexWrap: 'wrap' }}>
                <button type="submit" className="btn btn-primary" disabled={savingCfg}>
                  {savingCfg ? 'Saving…' : '💾 Save Settings'}
                </button>
                <button
                  type="button"
                  className="btn btn-secondary"
                  onClick={handleManualRefresh}
                  disabled={refreshing}
                >
                  {refreshing ? 'Refreshing…' : '🔄 Refresh Now'}
                </button>
              </div>
            </form>

            {refreshResult && (
              <div style={{
                marginTop: '1rem',
                background: '#ebf8ff',
                border: '1px solid #90cdf4',
                borderRadius: '6px',
                padding: '0.75rem 1rem',
                fontSize: '0.88rem',
              }}>
                <strong>Last refresh:</strong> updated {refreshResult.count} symbol{refreshResult.count !== 1 ? 's' : ''}.
                {refreshResult.updated && Object.keys(refreshResult.updated).length > 0 && (
                  <ul style={{ margin: '0.5rem 0 0', paddingLeft: '1.2rem' }}>
                    {Object.entries(refreshResult.updated).map(([sym, price]) => (
                      <li key={sym}>{sym}: ${Number(price).toFixed(2)}</li>
                    ))}
                  </ul>
                )}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
