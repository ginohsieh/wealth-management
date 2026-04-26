import React, { useEffect, useState } from 'react';

const API = import.meta.env.VITE_API_URL || '';

const fmt = (n) =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(n);

const fmtPct = (n) => `${n >= 0 ? '+' : ''}${n.toFixed(2)}%`;

const today = () => new Date().toISOString().split('T')[0];

const MARKETS = [
  { value: 'US',    label: 'US (commission-free, no tx tax)' },
  { value: 'TW',    label: 'TW (fee 0.1425%, sell tax 0.3%)' },
  { value: 'HK',    label: 'HK (fee 0.25%, stamp duty 0.1%)' },
  { value: 'OTHER', label: 'Other / Manual' },
];

const EMPTY_TRADE = {
  symbol: '', name: '', type: 'buy', date: today(),
  quantity: '', price: '', market: 'US', fee: '', tax: '', notes: '',
};

function calcFeeAndTax(market, type, quantity, price) {
  const value = (parseFloat(quantity) || 0) * (parseFloat(price) || 0);
  if (value <= 0) return { fee: 0, tax: 0 };
  switch (market) {
    case 'TW':
      return { fee: value * 0.001425, tax: type === 'sell' ? value * 0.003 : 0 };
    case 'HK':
      return { fee: value * 0.0025, tax: value * 0.001 };
    default:
      return { fee: 0, tax: 0 };
  }
}

export default function Portfolio() {
  const [tab, setTab]             = useState('holdings');  // 'holdings' | 'trades'
  const [holdings, setHoldings]   = useState([]);
  const [trades, setTrades]       = useState([]);
  const [loading, setLoading]     = useState(true);
  const [error, setError]         = useState('');
  const [showModal, setShowModal] = useState(false);
  const [form, setForm]           = useState(EMPTY_TRADE);
  const [saving, setSaving]       = useState(false);
  // inline price editing per symbol
  const [editingPrice, setEditingPrice] = useState(null);
  const [newPrice, setNewPrice]         = useState('');

  const loadHoldings = () =>
    fetch(`${API}/api/portfolio/holdings`).then(r => r.json()).then(setHoldings);
  const loadTrades = () =>
    fetch(`${API}/api/portfolio/trades`).then(r => r.json()).then(data => setTrades(data || []));

  const loadAll = () => {
    setLoading(true);
    Promise.all([loadHoldings(), loadTrades()])
      .catch(() => setError('Failed to load portfolio.'))
      .finally(() => setLoading(false));
  };

  useEffect(loadAll, []);

  // Auto-compute fee & tax when market/type/qty/price changes (only if user hasn't overridden)
  const computed = calcFeeAndTax(form.market, form.type, form.quantity, form.price);
  const displayFee = form.fee !== '' ? parseFloat(form.fee) : computed.fee;
  const displayTax = form.tax !== '' ? parseFloat(form.tax) : computed.tax;

  const handleFormChange = (field, value) =>
    setForm(f => ({ ...f, [field]: value }));

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      const payload = {
        ...form,
        quantity: parseFloat(form.quantity),
        price:    parseFloat(form.price),
        fee:      form.fee !== '' ? parseFloat(form.fee) : 0,
        tax:      form.tax !== '' ? parseFloat(form.tax) : 0,
      };
      const res = await fetch(`${API}/api/portfolio/trades`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (!res.ok) throw new Error();
      setShowModal(false);
      setForm(EMPTY_TRADE);
      loadAll();
    } catch {
      alert('Failed to record trade.');
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteTrade = async (id) => {
    if (!window.confirm('Delete this trade? Holdings will be recalculated.')) return;
    await fetch(`${API}/api/portfolio/trades/${id}`, { method: 'DELETE' });
    loadAll();
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

      {/* Tab switcher */}
      <div style={{ display: 'flex', gap: '0.5rem', marginBottom: '1rem' }}>
        <button
          className={`btn ${tab === 'holdings' ? 'btn-primary' : 'btn-secondary'}`}
          onClick={() => setTab('holdings')}
        >📊 Holdings</button>
        <button
          className={`btn ${tab === 'trades' ? 'btn-primary' : 'btn-secondary'}`}
          onClick={() => setTab('trades')}
        >📋 Trade History</button>
      </div>

      {/* ── Holdings tab ── */}
      {tab === 'holdings' && (
        <div className="table-container">
          <div className="table-header">
            <h2>Current Holdings</h2>
            <button className="btn btn-primary" onClick={() => setShowModal(true)}>
              + Add Trade
            </button>
          </div>
          {holdings.length === 0 ? (
            <div className="empty">No holdings yet. Add a buy trade to get started.</div>
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
                            type="number"
                            step="0.01"
                            min="0"
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
                          title="Click to update price"
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
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {/* ── Trade History tab ── */}
      {tab === 'trades' && (
        <div className="table-container">
          <div className="table-header">
            <h2>{trades.length} trade{trades.length !== 1 ? 's' : ''}</h2>
            <button className="btn btn-primary" onClick={() => setShowModal(true)}>
              + Add Trade
            </button>
          </div>
          {trades.length === 0 ? (
            <div className="empty">No trades yet.</div>
          ) : (
            <table>
              <thead>
                <tr>
                  <th>Date</th>
                  <th>Type</th>
                  <th>Symbol</th>
                  <th>Shares</th>
                  <th>Price</th>
                  <th>Fee</th>
                  <th>Tax</th>
                  <th>Total</th>
                  <th>Market</th>
                  <th>Notes</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {trades.map(t => {
                  const total = t.quantity * t.price + (t.type === 'buy' ? t.fee : -(t.fee + t.tax));
                  return (
                    <tr key={t.id}>
                      <td>{t.date}</td>
                      <td>
                        <span className={`badge ${t.type === 'buy' ? 'badge-income' : 'badge-expense'}`}>
                          {t.type.toUpperCase()}
                        </span>
                      </td>
                      <td><strong>{t.symbol}</strong></td>
                      <td>{t.quantity}</td>
                      <td>{fmt(t.price)}</td>
                      <td style={{ color: '#718096' }}>{fmt(t.fee)}</td>
                      <td style={{ color: '#718096' }}>{fmt(t.tax)}</td>
                      <td className={t.type === 'buy' ? 'amount-negative' : 'amount-positive'}>
                        {t.type === 'buy' ? '-' : '+'}{fmt(Math.abs(total))}
                      </td>
                      <td><span className="badge badge-checking">{t.market}</span></td>
                      <td style={{ color: '#718096', maxWidth: '120px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{t.notes}</td>
                      <td>
                        <button className="btn btn-danger" onClick={() => handleDeleteTrade(t.id)}>Delete</button>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          )}
        </div>
      )}

      {/* ── Add Trade Modal ── */}
      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal" onClick={e => e.stopPropagation()} style={{ maxWidth: '520px' }}>
            <h2>Record Trade</h2>
            <form onSubmit={handleSubmit}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0 1rem' }}>
                <div className="form-group">
                  <label>Symbol</label>
                  <input
                    required
                    value={form.symbol}
                    onChange={e => handleFormChange('symbol', e.target.value.toUpperCase())}
                    placeholder="e.g. AAPL"
                  />
                </div>
                <div className="form-group">
                  <label>Type</label>
                  <select
                    value={form.type}
                    onChange={e => handleFormChange('type', e.target.value)}
                  >
                    <option value="buy">Buy</option>
                    <option value="sell">Sell</option>
                  </select>
                </div>
              </div>

              <div className="form-group">
                <label>Company Name</label>
                <input
                  required
                  value={form.name}
                  onChange={e => handleFormChange('name', e.target.value)}
                  placeholder="e.g. Apple Inc."
                />
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0 1rem' }}>
                <div className="form-group">
                  <label>Date</label>
                  <input
                    type="date"
                    required
                    value={form.date}
                    onChange={e => handleFormChange('date', e.target.value)}
                  />
                </div>
                <div className="form-group">
                  <label>Market</label>
                  <select
                    value={form.market}
                    onChange={e => handleFormChange('market', e.target.value)}
                  >
                    {MARKETS.map(m => (
                      <option key={m.value} value={m.value}>{m.label}</option>
                    ))}
                  </select>
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0 1rem' }}>
                <div className="form-group">
                  <label>Quantity (shares)</label>
                  <input
                    type="number"
                    step="0.0001"
                    min="0"
                    required
                    value={form.quantity}
                    onChange={e => handleFormChange('quantity', e.target.value)}
                    placeholder="0"
                  />
                </div>
                <div className="form-group">
                  <label>Price per Share</label>
                  <input
                    type="number"
                    step="0.0001"
                    min="0"
                    required
                    value={form.price}
                    onChange={e => handleFormChange('price', e.target.value)}
                    placeholder="0.00"
                  />
                </div>
              </div>

              {/* Fee / Tax preview */}
              <div style={{
                background: '#f7fafc',
                border: '1px solid #e8ecf0',
                borderRadius: '6px',
                padding: '0.75rem 1rem',
                marginBottom: '1rem',
                fontSize: '0.86rem',
                color: '#4a5568',
              }}>
                <strong>Auto-calculated from market rules</strong>
                <div style={{ marginTop: '0.4rem', display: 'flex', gap: '2rem' }}>
                  <span>Fee: <strong>{fmt(computed.fee)}</strong></span>
                  <span>Tax: <strong>{fmt(computed.tax)}</strong></span>
                  <span>Trade value: <strong>{fmt((parseFloat(form.quantity)||0)*(parseFloat(form.price)||0))}</strong></span>
                </div>
                <div style={{ marginTop: '0.3rem', color: '#718096', fontSize: '0.82rem' }}>
                  Leave override fields blank to use these values.
                </div>
              </div>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0 1rem' }}>
                <div className="form-group">
                  <label>Fee Override (optional)</label>
                  <input
                    type="number"
                    step="0.01"
                    min="0"
                    value={form.fee}
                    onChange={e => handleFormChange('fee', e.target.value)}
                    placeholder={computed.fee.toFixed(2)}
                  />
                </div>
                <div className="form-group">
                  <label>Tax Override (optional)</label>
                  <input
                    type="number"
                    step="0.01"
                    min="0"
                    value={form.tax}
                    onChange={e => handleFormChange('tax', e.target.value)}
                    placeholder={computed.tax.toFixed(2)}
                  />
                </div>
              </div>

              {/* Total cost summary */}
              {(form.quantity && form.price) && (
                <div style={{
                  background: form.type === 'buy' ? '#f0fff4' : '#fff5f5',
                  border: `1px solid ${form.type === 'buy' ? '#9ae6b4' : '#feb2b2'}`,
                  borderRadius: '6px',
                  padding: '0.6rem 1rem',
                  marginBottom: '1rem',
                  fontSize: '0.9rem',
                  fontWeight: 600,
                  color: form.type === 'buy' ? '#276749' : '#c53030',
                }}>
                  {form.type === 'buy'
                    ? `Total outlay: ${fmt((parseFloat(form.quantity)||0)*(parseFloat(form.price)||0) + displayFee)}`
                    : `Net proceeds: ${fmt((parseFloat(form.quantity)||0)*(parseFloat(form.price)||0) - displayFee - displayTax)}`
                  }
                </div>
              )}

              <div className="form-group">
                <label>Notes (optional)</label>
                <input
                  value={form.notes}
                  onChange={e => handleFormChange('notes', e.target.value)}
                  placeholder="e.g. Q1 rebalance"
                />
              </div>

              <div className="form-actions">
                <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary" disabled={saving}>
                  {saving ? 'Saving…' : `Record ${form.type === 'buy' ? 'Buy' : 'Sell'}`}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}

