import React, { useEffect, useState } from 'react';

const API = import.meta.env.VITE_API_URL || '';

const fmt = (n) =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(n);

const EMPTY_FORM = { name: '', type: 'bank', balance: '', currency: 'TWD' };

export default function Accounts() {
  const [accounts, setAccounts]           = useState([]);
  const [loading, setLoading]             = useState(true);
  const [error, setError]                 = useState('');
  const [showModal, setShowModal]         = useState(false);
  const [form, setForm]                   = useState(EMPTY_FORM);
  const [saving, setSaving]               = useState(false);
  // Transaction history drawer
  const [selectedAccount, setSelectedAccount] = useState(null); // account object
  const [txHistory, setTxHistory]         = useState([]);
  const [txLoading, setTxLoading]         = useState(false);

  const load = () => {
    setLoading(true);
    fetch(`${API}/api/accounts`)
      .then(r => r.json())
      .then(setAccounts)
      .catch(() => setError('Failed to load accounts.'))
      .finally(() => setLoading(false));
  };

  useEffect(load, []);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      const res = await fetch(`${API}/api/accounts`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ...form, balance: parseFloat(form.balance) || 0 }),
      });
      if (!res.ok) throw new Error();
      setShowModal(false);
      setForm(EMPTY_FORM);
      load();
    } catch {
      alert('Failed to create account.');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Delete this account?')) return;
    await fetch(`${API}/api/accounts/${id}`, { method: 'DELETE' });
    if (selectedAccount && selectedAccount.id === id) setSelectedAccount(null);
    load();
  };

  const handleSelectAccount = (acct) => {
    if (selectedAccount && selectedAccount.id === acct.id) {
      setSelectedAccount(null);
      setTxHistory([]);
      return;
    }
    setSelectedAccount(acct);
    setTxLoading(true);
    fetch(`${API}/api/transactions?account_id=${acct.id}`)
      .then(r => r.json())
      .then(data => setTxHistory((data || []).sort((a, b) => new Date(b.date) - new Date(a.date))))
      .catch(() => setTxHistory([]))
      .finally(() => setTxLoading(false));
  };

  const totalBalance = accounts.reduce((s, a) => s + a.balance, 0);

  if (loading) return <div className="loading">Loading accounts…</div>;
  if (error)   return <div className="error">{error}</div>;

  return (
    <div>
      <h1 className="page-title">Accounts</h1>

      <div className="table-container">
        <div className="table-header">
          <h2>{accounts.length} account{accounts.length !== 1 ? 's' : ''}</h2>
          <button className="btn btn-primary" onClick={() => setShowModal(true)}>
            + Add Account
          </button>
        </div>
        {accounts.length === 0 ? (
          <div className="empty">No accounts yet. Add one above.</div>
        ) : (
          <table>
            <thead>
              <tr>
                <th style={{ width: '1.5rem' }}></th>
                <th>Name</th>
                <th>Type</th>
                <th>Currency</th>
                <th>Balance</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {accounts.map(a => {
                const isSelected = selectedAccount && selectedAccount.id === a.id;
                return (
                  <React.Fragment key={a.id}>
                    <tr
                      style={{ cursor: 'pointer', background: isSelected ? '#f7fafc' : '' }}
                      onClick={() => handleSelectAccount(a)}
                    >
                      <td style={{ textAlign: 'center', color: '#718096', fontSize: '0.75rem' }}>
                        {isSelected ? '▼' : '▶'}
                      </td>
                      <td><strong>{a.name}</strong></td>
                      <td><span className={`badge badge-${a.type}`}>{a.type}</span></td>
                      <td>{a.currency}</td>
                      <td className={a.balance >= 0 ? 'amount-positive' : 'amount-negative'}>
                        {fmt(a.balance)}
                      </td>
                      <td onClick={e => e.stopPropagation()}>
                        <button className="btn btn-danger" onClick={() => handleDelete(a.id)}>
                          Delete
                        </button>
                      </td>
                    </tr>
                    {isSelected && (
                      <tr style={{ background: '#f7fafc' }}>
                        <td></td>
                        <td colSpan={5} style={{ paddingBottom: '1rem' }}>
                          <div style={{ borderTop: '1px solid #e2e8f0', paddingTop: '0.75rem' }}>
                            <strong style={{ fontSize: '0.9rem', color: '#4a5568' }}>
                              Transaction History – {a.name}
                            </strong>
                            {txLoading ? (
                              <div style={{ color: '#718096', marginTop: '0.5rem' }}>Loading…</div>
                            ) : txHistory.length === 0 ? (
                              <div style={{ color: '#a0aec0', marginTop: '0.5rem', fontSize: '0.85rem' }}>
                                No transactions found for this account.
                              </div>
                            ) : (
                              <table style={{ marginTop: '0.5rem', width: '100%', fontSize: '0.84rem' }}>
                                <thead>
                                  <tr>
                                    <th>Date</th>
                                    <th>Description</th>
                                    <th>Category</th>
                                    <th>Type</th>
                                    <th>Symbol</th>
                                    <th>Amount</th>
                                  </tr>
                                </thead>
                                <tbody>
                                  {txHistory.map(t => (
                                    <tr key={t.id}>
                                      <td>{t.date}</td>
                                      <td>{t.description}</td>
                                      <td>{t.category}</td>
                                      <td>
                                        {t.subtype
                                          ? <span className={`badge ${t.subtype === 'stock_buy' ? 'badge-income' : 'badge-expense'}`}>
                                              {t.subtype === 'stock_buy' ? 'BUY' : 'SELL'}
                                            </span>
                                          : <span className={`badge badge-${t.type}`}>{t.type}</span>
                                        }
                                      </td>
                                      <td>
                                        {t.symbol
                                          ? <span className="badge badge-checking">{t.symbol}</span>
                                          : <span style={{ color: '#a0aec0' }}>—</span>}
                                      </td>
                                      <td className={t.amount >= 0 ? 'amount-positive' : 'amount-negative'}>
                                        {fmt(t.amount)}
                                      </td>
                                    </tr>
                                  ))}
                                </tbody>
                              </table>
                            )}
                          </div>
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                );
              })}
              <tr>
                <td></td>
                <td colSpan={3}><strong>Total</strong></td>
                <td className={totalBalance >= 0 ? 'amount-positive' : 'amount-negative'}>
                  <strong>{fmt(totalBalance)}</strong>
                </td>
                <td></td>
              </tr>
            </tbody>
          </table>
        )}
      </div>

      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <h2>Add Account</h2>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label>Account Name</label>
                <input
                  required
                  value={form.name}
                  onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                  placeholder="e.g. Chase Checking"
                />
              </div>
              <div className="form-group">
                <label>Type</label>
                <select
                  value={form.type}
                  onChange={e => setForm(f => ({ ...f, type: e.target.value }))}
                >
                  <option value="bank">🏦 Bank（銀行）</option>
                  <option value="brokerage">📈 Brokerage（證券）</option>
                  <option value="credit">💳 Credit（信用）</option>
                </select>
              </div>
              <div className="form-group">
                <label>Balance (TWD)</label>
                <input
                  type="number"
                  step="0.01"
                  required
                  value={form.balance}
                  onChange={e => setForm(f => ({ ...f, balance: e.target.value }))}
                  placeholder="0.00"
                />
              </div>
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
