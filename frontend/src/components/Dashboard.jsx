import React, { useEffect, useState } from 'react';

const API = import.meta.env.VITE_API_URL || '';

const fmt = (n) =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(n);

const SUMMARY_CARDS = [
  { key: 'net_worth',        label: 'Net Worth',         variant: 'positive', cls: '' },
  { key: 'total_assets',     label: 'Total Assets',      variant: 'positive', cls: 'info' },
  { key: 'total_liabilities',label: 'Liabilities',       variant: 'negative', cls: 'danger' },
  { key: 'portfolio_value',  label: 'Portfolio Value',   variant: 'positive', cls: 'purple' },
  { key: 'monthly_income',   label: 'Monthly Income',    variant: 'positive', cls: '' },
  { key: 'monthly_expenses', label: 'Monthly Expenses',  variant: 'negative', cls: 'danger' },
  { key: 'monthly_savings',  label: 'Monthly Savings',   variant: 'positive', cls: 'warning' },
];

export default function Dashboard() {
  const [summary, setSummary]           = useState(null);
  const [transactions, setTransactions] = useState([]);
  const [accounts, setAccounts]         = useState([]);
  const [loading, setLoading]           = useState(true);
  const [error, setError]               = useState('');

  useEffect(() => {
    const all = [
      fetch(`${API}/api/summary`).then(r => r.json()),
      fetch(`${API}/api/transactions`).then(r => r.json()),
      fetch(`${API}/api/accounts`).then(r => r.json()),
    ];
    Promise.all(all)
      .then(([s, t, a]) => { setSummary(s); setTransactions(t || []); setAccounts(a || []); })
      .catch(() => setError('Failed to load dashboard data. Is the backend running?'))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div className="loading">Loading dashboard…</div>;
  if (error)   return <div className="error">{error}</div>;

  const accountName = (id) => {
    const a = accounts.find(a => a.id === id);
    return a ? a.name : id;
  };

  const recent = [...transactions]
    .sort((a, b) => new Date(b.date) - new Date(a.date))
    .slice(0, 8);

  return (
    <div>
      <h1 className="page-title">Dashboard</h1>

      {/* Summary cards */}
      <div className="summary-grid">
        {SUMMARY_CARDS.map(({ key, label, variant, cls }) => (
          <div key={key} className={`summary-card ${cls}`}>
            <div className="card-title">{label}</div>
            <div className={`card-value ${variant}`}>{fmt(summary[key])}</div>
          </div>
        ))}
      </div>

      {/* Account balances */}
      <div className="table-container" style={{ marginBottom: '1.5rem' }}>
        <div className="table-header"><h2>Account Balances</h2></div>
        <table>
          <thead>
            <tr>
              <th>Account</th>
              <th>Type</th>
              <th>Balance</th>
            </tr>
          </thead>
          <tbody>
            {accounts.map(a => (
              <tr key={a.id}>
                <td>{a.name}</td>
                <td><span className={`badge badge-${a.type}`}>{a.type}</span></td>
                <td className={a.balance >= 0 ? 'amount-positive' : 'amount-negative'}>
                  {fmt(a.balance)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Recent transactions */}
      <div className="table-container">
        <div className="table-header"><h2>Recent Transactions</h2></div>
        <table>
          <thead>
            <tr>
              <th>Date</th>
              <th>Description</th>
              <th>Account</th>
              <th>Category</th>
              <th>Amount</th>
            </tr>
          </thead>
          <tbody>
            {recent.map(t => (
              <tr key={t.id}>
                <td>{t.date}</td>
                <td>{t.description}</td>
                <td>{accountName(t.account_id)}</td>
                <td>
                  <span className={`badge badge-${t.type}`}>{t.category}</span>
                </td>
                <td className={t.amount >= 0 ? 'amount-positive' : 'amount-negative'}>
                  {fmt(t.amount)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
