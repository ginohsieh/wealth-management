import React, { useEffect, useState } from 'react';

const API = import.meta.env.VITE_API_URL || '';

const fmt = (n) =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(n);

const today = () => new Date().toISOString().split('T')[0];

const EMPTY_FORM = {
  account_id: '',
  date: today(),
  description: '',
  amount: '',
  category: '',
  type: 'expense',
};

export default function Transactions() {
  const [transactions, setTransactions] = useState([]);
  const [accounts, setAccounts]         = useState([]);
  const [filterAccount, setFilterAccount] = useState('');
  const [loading, setLoading]           = useState(true);
  const [error, setError]               = useState('');
  const [showModal, setShowModal]       = useState(false);
  const [form, setForm]                 = useState(EMPTY_FORM);
  const [saving, setSaving]             = useState(false);

  const loadTransactions = (accountId = filterAccount) => {
    const qs = accountId ? `?account_id=${accountId}` : '';
    fetch(`${API}/api/transactions${qs}`)
      .then(r => r.json())
      .then(data => setTransactions(data || []))
      .catch(() => setError('Failed to load transactions.'));
  };

  useEffect(() => {
    setLoading(true);
    Promise.all([
      fetch(`${API}/api/accounts`).then(r => r.json()),
    ])
      .then(([accts]) => {
        setAccounts(accts);
        loadTransactions('');
      })
      .catch(() => setError('Failed to load data.'))
      .finally(() => setLoading(false));
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const handleFilterChange = (e) => {
    setFilterAccount(e.target.value);
    loadTransactions(e.target.value);
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      const payload = {
        ...form,
        amount: form.type === 'expense'
          ? -Math.abs(parseFloat(form.amount))
          : Math.abs(parseFloat(form.amount)),
      };
      const res = await fetch(`${API}/api/transactions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (!res.ok) throw new Error();
      setShowModal(false);
      setForm(EMPTY_FORM);
      loadTransactions(filterAccount);
    } catch {
      alert('Failed to create transaction.');
    } finally {
      setSaving(false);
    }
  };

  const accountName = (id) => {
    const a = accounts.find(a => a.id === id);
    return a ? a.name : id;
  };

  const sorted = [...transactions].sort((a, b) => new Date(b.date) - new Date(a.date));

  if (loading) return <div className="loading">Loading transactions…</div>;
  if (error)   return <div className="error">{error}</div>;

  return (
    <div>
      <h1 className="page-title">Transactions</h1>

      <div className="table-container">
        <div className="table-header">
          <h2>{sorted.length} transaction{sorted.length !== 1 ? 's' : ''}</h2>
          <button className="btn btn-primary" onClick={() => setShowModal(true)}>
            + Add Transaction
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

        {sorted.length === 0 ? (
          <div className="empty">No transactions found.</div>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Date</th>
                <th>Description</th>
                <th>Account</th>
                <th>Category</th>
                <th>Type</th>
                <th>Amount</th>
              </tr>
            </thead>
            <tbody>
              {sorted.map(t => (
                <tr key={t.id}>
                  <td>{t.date}</td>
                  <td>{t.description}</td>
                  <td>{accountName(t.account_id)}</td>
                  <td>{t.category}</td>
                  <td><span className={`badge badge-${t.type}`}>{t.type}</span></td>
                  <td className={t.amount >= 0 ? 'amount-positive' : 'amount-negative'}>
                    {fmt(t.amount)}
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
            <h2>Add Transaction</h2>
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
                <label>Date</label>
                <input
                  type="date"
                  required
                  value={form.date}
                  onChange={e => setForm(f => ({ ...f, date: e.target.value }))}
                />
              </div>
              <div className="form-group">
                <label>Description</label>
                <input
                  required
                  value={form.description}
                  onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
                  placeholder="e.g. Grocery Store"
                />
              </div>
              <div className="form-group">
                <label>Type</label>
                <select
                  value={form.type}
                  onChange={e => setForm(f => ({ ...f, type: e.target.value }))}
                >
                  <option value="income">Income</option>
                  <option value="expense">Expense</option>
                </select>
              </div>
              <div className="form-group">
                <label>Category</label>
                <input
                  required
                  value={form.category}
                  onChange={e => setForm(f => ({ ...f, category: e.target.value }))}
                  placeholder="e.g. Food, Housing, Income…"
                />
              </div>
              <div className="form-group">
                <label>Amount (USD)</label>
                <input
                  type="number"
                  step="0.01"
                  min="0"
                  required
                  value={form.amount}
                  onChange={e => setForm(f => ({ ...f, amount: e.target.value }))}
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
