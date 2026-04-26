import React, { useEffect, useState } from 'react';

const API = import.meta.env.VITE_API_URL || '';

const fmt = (n) =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(n);

const fmtPct = (n) => `${n >= 0 ? '+' : ''}${n.toFixed(2)}%`;

const EMPTY_FORM = {
  symbol: '',
  name: '',
  quantity: '',
  purchase_price: '',
  current_price: '',
};

export default function Portfolio() {
  const [assets, setAssets]     = useState([]);
  const [loading, setLoading]   = useState(true);
  const [error, setError]       = useState('');
  const [showModal, setShowModal] = useState(false);
  const [form, setForm]         = useState(EMPTY_FORM);
  const [saving, setSaving]     = useState(false);

  const load = () => {
    setLoading(true);
    fetch(`${API}/api/portfolio`)
      .then(r => r.json())
      .then(setAssets)
      .catch(() => setError('Failed to load portfolio.'))
      .finally(() => setLoading(false));
  };

  useEffect(load, []);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      const res = await fetch(`${API}/api/portfolio`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          ...form,
          quantity:       parseFloat(form.quantity),
          purchase_price: parseFloat(form.purchase_price),
          current_price:  parseFloat(form.current_price),
        }),
      });
      if (!res.ok) throw new Error();
      setShowModal(false);
      setForm(EMPTY_FORM);
      load();
    } catch {
      alert('Failed to add asset.');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Remove this asset from portfolio?')) return;
    await fetch(`${API}/api/portfolio/${id}`, { method: 'DELETE' });
    load();
  };

  const totalValue   = assets.reduce((s, a) => s + a.value, 0);
  const totalCost    = assets.reduce((s, a) => s + a.quantity * a.purchase_price, 0);
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
          <div className="card-value">{assets.length}</div>
        </div>
      </div>

      <div className="table-container">
        <div className="table-header">
          <h2>Holdings</h2>
          <button className="btn btn-primary" onClick={() => setShowModal(true)}>
            + Add Holding
          </button>
        </div>
        {assets.length === 0 ? (
          <div className="empty">No holdings yet. Add one above.</div>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Symbol</th>
                <th>Name</th>
                <th>Qty</th>
                <th>Purchase Price</th>
                <th>Current Price</th>
                <th>Value</th>
                <th>Gain / Loss</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {assets.map(a => (
                <tr key={a.id}>
                  <td><strong>{a.symbol}</strong></td>
                  <td>{a.name}</td>
                  <td>{a.quantity}</td>
                  <td>{fmt(a.purchase_price)}</td>
                  <td>{fmt(a.current_price)}</td>
                  <td className="amount-positive">{fmt(a.value)}</td>
                  <td className={a.gain_loss >= 0 ? 'gain' : 'loss'}>
                    {fmt(a.gain_loss)}{' '}
                    <span style={{ fontSize: '0.82rem' }}>
                      ({fmtPct(a.gain_loss_pct)})
                    </span>
                  </td>
                  <td>
                    <button className="btn btn-danger" onClick={() => handleDelete(a.id)}>
                      Remove
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <h2>Add Holding</h2>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label>Symbol</label>
                <input
                  required
                  value={form.symbol}
                  onChange={e => setForm(f => ({ ...f, symbol: e.target.value.toUpperCase() }))}
                  placeholder="e.g. AAPL"
                />
              </div>
              <div className="form-group">
                <label>Company Name</label>
                <input
                  required
                  value={form.name}
                  onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                  placeholder="e.g. Apple Inc."
                />
              </div>
              <div className="form-group">
                <label>Quantity</label>
                <input
                  type="number"
                  step="0.0001"
                  min="0"
                  required
                  value={form.quantity}
                  onChange={e => setForm(f => ({ ...f, quantity: e.target.value }))}
                  placeholder="0"
                />
              </div>
              <div className="form-group">
                <label>Purchase Price (USD)</label>
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  required
                  value={form.purchase_price}
                  onChange={e => setForm(f => ({ ...f, purchase_price: e.target.value }))}
                  placeholder="0.00"
                />
              </div>
              <div className="form-group">
                <label>Current Price (USD)</label>
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  required
                  value={form.current_price}
                  onChange={e => setForm(f => ({ ...f, current_price: e.target.value }))}
                  placeholder="0.00"
                />
              </div>
              <div className="form-actions">
                <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary" disabled={saving}>
                  {saving ? 'Saving…' : 'Save'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
