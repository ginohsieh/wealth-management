import React, { useEffect, useState } from 'react';

const API = import.meta.env.VITE_API_URL || '';

const ASSET_TYPES = [
  { value: 'cash',        label: '💵 Cash（現金）' },
  { value: 'stock',       label: '📈 Stock（股票）' },
  { value: 'credit_line', label: '💳 Credit Line（信用額度）' },
];

const EMPTY_FORM = {
  account_id: '',
  type: 'cash',
  name: '',
  symbol: '',
  currency: 'TWD',
};

export default function Assets() {
  const [assets, setAssets]     = useState([]);
  const [accounts, setAccounts] = useState([]);
  const [filterAccount, setFilterAccount] = useState('');
  const [loading, setLoading]   = useState(true);
  const [error, setError]       = useState('');
  const [showModal, setShowModal] = useState(false);
  const [form, setForm]         = useState(EMPTY_FORM);
  const [saving, setSaving]     = useState(false);

  const loadAssets = (accountId = filterAccount) => {
    const qs = accountId ? `?account_id=${accountId}` : '';
    fetch(`${API}/api/assets${qs}`)
      .then(r => r.json())
      .then(data => setAssets(data || []))
      .catch(() => setError('Failed to load assets.'));
  };

  useEffect(() => {
    setLoading(true);
    Promise.all([
      fetch(`${API}/api/accounts`).then(r => r.json()),
    ])
      .then(([accts]) => {
        setAccounts(accts || []);
        loadAssets('');
      })
      .catch(() => setError('Failed to load data.'))
      .finally(() => setLoading(false));
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const handleFilterChange = (e) => {
    setFilterAccount(e.target.value);
    loadAssets(e.target.value);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      const res = await fetch(`${API}/api/assets`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(form),
      });
      if (!res.ok) throw new Error();
      setShowModal(false);
      setForm(EMPTY_FORM);
      loadAssets(filterAccount);
    } catch {
      alert('Failed to create asset.');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Delete this asset?')) return;
    await fetch(`${API}/api/assets/${id}`, { method: 'DELETE' });
    loadAssets(filterAccount);
  };

  const accountName = (id) => {
    const a = accounts.find(a => a.id === String(id));
    return a ? a.name : id;
  };

  const typeLabel = (t) => ASSET_TYPES.find(x => x.value === t)?.label ?? t;

  if (loading) return <div className="loading">Loading assets…</div>;
  if (error)   return <div className="error">{error}</div>;

  return (
    <div>
      <h1 className="page-title">Assets</h1>

      <div className="table-container">
        <div className="table-header">
          <h2>{assets.length} asset{assets.length !== 1 ? 's' : ''}</h2>
          <button className="btn btn-primary" onClick={() => setShowModal(true)}>
            + Add Asset
          </button>
        </div>

        {/* Filter bar */}
        <div className="filter-bar">
          <label>Filter by account:</label>
          <select value={filterAccount} onChange={handleFilterChange}>
            <option value="">All accounts</option>
            {accounts.map(a => (
              <option key={a.id} value={a.id}>{a.name}</option>
            ))}
          </select>
        </div>

        {assets.length === 0 ? (
          <div className="empty">No assets yet. Add one above.</div>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Account</th>
                <th>Type</th>
                <th>Name</th>
                <th>Symbol</th>
                <th>Currency</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {assets.map(a => (
                <tr key={a.id}>
                  <td>{accountName(a.account_id)}</td>
                  <td><span className={`badge badge-${a.type}`}>{typeLabel(a.type)}</span></td>
                  <td><strong>{a.name}</strong></td>
                  <td>{a.symbol || '—'}</td>
                  <td>{a.currency}</td>
                  <td>
                    <button className="btn btn-danger" onClick={() => handleDelete(a.id)}>
                      Delete
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
            <h2>Add Asset</h2>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label>Account</label>
                <select
                  required
                  value={form.account_id}
                  onChange={e => setForm(f => ({ ...f, account_id: e.target.value }))}
                >
                  <option value="">Select account…</option>
                  {accounts.map(a => (
                    <option key={a.id} value={a.id}>{a.name}</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>Type</label>
                <select
                  value={form.type}
                  onChange={e => setForm(f => ({ ...f, type: e.target.value }))}
                >
                  {ASSET_TYPES.map(t => (
                    <option key={t.value} value={t.value}>{t.label}</option>
                  ))}
                </select>
              </div>
              <div className="form-group">
                <label>Name</label>
                <input
                  required
                  value={form.name}
                  onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                  placeholder="e.g. 活存、AAPL"
                />
              </div>
              {form.type === 'stock' && (
                <div className="form-group">
                  <label>Symbol</label>
                  <input
                    value={form.symbol}
                    onChange={e => setForm(f => ({ ...f, symbol: e.target.value.toUpperCase() }))}
                    placeholder="e.g. AAPL, 2330.TW"
                  />
                </div>
              )}
              <div className="form-group">
                <label>Currency</label>
                <select
                  value={form.currency}
                  onChange={e => setForm(f => ({ ...f, currency: e.target.value }))}
                >
                  <option value="TWD">TWD</option>
                  <option value="USD">USD</option>
                  <option value="EUR">EUR</option>
                  <option value="GBP">GBP</option>
                  <option value="JPY">JPY</option>
                </select>
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
