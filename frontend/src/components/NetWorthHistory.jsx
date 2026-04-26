import React, { useEffect, useState } from 'react';

const API = import.meta.env.VITE_API_URL || '';

const fmt = (n) =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(n);

const today = () => new Date().toISOString().split('T')[0];

export default function NetWorthHistory() {
  const [snapshots, setSnapshots] = useState([]);
  const [loading, setLoading]     = useState(true);
  const [error, setError]         = useState('');
  const [recording, setRecording] = useState(false);
  const [recordDate, setRecordDate] = useState(today());

  const load = () => {
    setLoading(true);
    fetch(`${API}/api/snapshots`)
      .then(r => r.json())
      .then(data => setSnapshots(data || []))
      .catch(() => setError('Failed to load snapshots.'))
      .finally(() => setLoading(false));
  };

  useEffect(load, []);

  const handleRecord = async () => {
    setRecording(true);
    try {
      const res = await fetch(`${API}/api/snapshots`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ date: recordDate }),
      });
      if (!res.ok) throw new Error();
      load();
    } catch {
      alert('Failed to record snapshot.');
    } finally {
      setRecording(false);
    }
  };

  if (loading) return <div className="loading">Loading history…</div>;
  if (error)   return <div className="error">{error}</div>;

  const latestNetWorth  = snapshots.length > 0 ? snapshots[0].net_worth : null;
  const latestPortfolio = snapshots.length > 0 ? snapshots[0].portfolio_value : null;

  return (
    <div>
      <h1 className="page-title">Net Worth History</h1>

      {/* Summary cards */}
      {snapshots.length > 0 && (
        <div className="summary-grid" style={{ marginBottom: '1.5rem' }}>
          <div className="summary-card info">
            <div className="card-title">Latest Net Worth</div>
            <div className="card-value positive">{fmt(latestNetWorth)}</div>
          </div>
          <div className="summary-card purple">
            <div className="card-title">Latest Portfolio Value</div>
            <div className="card-value positive">{fmt(latestPortfolio)}</div>
          </div>
          <div className="summary-card">
            <div className="card-title">Days Recorded</div>
            <div className="card-value">{snapshots.length}</div>
          </div>
        </div>
      )}

      <div className="table-container">
        <div className="table-header">
          <h2>Daily Snapshots</h2>
          <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
            <input
              type="date"
              value={recordDate}
              onChange={e => setRecordDate(e.target.value)}
              style={{
                padding: '0.45rem 0.7rem',
                border: '1px solid #cbd5e0',
                borderRadius: '6px',
                fontSize: '0.88rem',
              }}
            />
            <button
              className="btn btn-primary"
              onClick={handleRecord}
              disabled={recording}
            >
              {recording ? 'Recording…' : '📸 Record Snapshot'}
            </button>
          </div>
        </div>

        {snapshots.length === 0 ? (
          <div className="empty">No snapshots yet. Click "Record Snapshot" to capture today's state.</div>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Date</th>
                <th>Portfolio Value</th>
                <th>Total Assets</th>
                <th>Liabilities</th>
                <th>Net Worth</th>
                <th>Change</th>
              </tr>
            </thead>
            <tbody>
              {snapshots.map((s, idx) => {
                const prev = snapshots[idx + 1];
                const change = prev != null ? s.net_worth - prev.net_worth : null;
                return (
                  <tr key={s.id}>
                    <td><strong>{s.date}</strong></td>
                    <td className="amount-positive">{fmt(s.portfolio_value)}</td>
                    <td className="amount-positive">{fmt(s.total_assets)}</td>
                    <td className="amount-negative">{fmt(s.total_liabilities)}</td>
                    <td className={s.net_worth >= 0 ? 'amount-positive' : 'amount-negative'}>
                      {fmt(s.net_worth)}
                    </td>
                    <td>
                      {change != null ? (
                        <span className={change >= 0 ? 'gain' : 'loss'}>
                          {change >= 0 ? '+' : ''}{fmt(change)}
                        </span>
                      ) : (
                        <span style={{ color: '#a0aec0' }}>—</span>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
