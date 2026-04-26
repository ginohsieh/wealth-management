import React, { useEffect, useState } from 'react';

const API = import.meta.env.VITE_API_URL || '';

const fmt = (n) =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(n);

const today = () => new Date().toISOString().split('T')[0];

const MARKETS = [
  { value: 'US',    label: 'US (commission-free, no tx tax)' },
  { value: 'TW',    label: 'TW (fee 0.1425%, sell tax 0.3%)' },
  { value: 'HK',    label: 'HK (fee 0.25%, stamp duty 0.1%)' },
  { value: 'OTHER', label: 'Other / Manual' },
];

const STOCK_SUBTYPES = ['stock_buy', 'stock_sell'];

function calcFeeAndTax(market, subtype, quantity, price) {
  const value = (parseFloat(quantity) || 0) * (parseFloat(price) || 0);
  if (value <= 0) return { fee: 0, tax: 0 };
  switch (market) {
    case 'TW':
      return { fee: value * 0.001425, tax: subtype === 'stock_sell' ? value * 0.003 : 0 };
    case 'HK':
      return { fee: value * 0.0025, tax: value * 0.001 };
    default:
      return { fee: 0, tax: 0 };
  }
}

const EMPTY_FORM = {
  account_id: '',
  bank_account_id: '',
  date: today(),
  description: '',
  amount: '',
  category: '',
  type: 'expense',
  subtype: '',
  symbol: '',
  quantity: '',
  price: '',
  market: 'US',
  fee: '',
  tax: '',
};

export default function Transactions() {
  const [transactions, setTransactions] = useState([]);
  const [accounts, setAccounts]         = useState([]);
  const [settlements, setSettlements]   = useState([]);
  const [filterAccount, setFilterAccount] = useState('');
  const [loading, setLoading]           = useState(true);
  const [error, setError]               = useState('');
  const [showModal, setShowModal]       = useState(false);
  const [form, setForm]                 = useState(EMPTY_FORM);
  const [saving, setSaving]             = useState(false);
  const [resolvedName, setResolvedName] = useState('');
  const [resolvingName, setResolvingName] = useState(false);

  const isStockTx = STOCK_SUBTYPES.includes(form.subtype);
  const computed  = isStockTx
    ? calcFeeAndTax(form.market, form.subtype, form.quantity, form.price)
    : { fee: 0, tax: 0 };
  const displayFee = form.fee !== '' ? parseFloat(form.fee) : computed.fee;
  const displayTax = form.tax !== '' ? parseFloat(form.tax) : computed.tax;

  const loadTransactions = (accountId = filterAccount) => {
    const qs = accountId ? `?account_id=${accountId}` : '';
    fetch(`${API}/api/transactions${qs}`)
      .then(r => r.json())
      .then(data => setTransactions(data || []))
      .catch(() => setError('Failed to load transactions.'));
  };

  const loadSettlements = () => {
    fetch(`${API}/api/settlements?status=pending`)
      .then(r => r.json())
      .then(data => setSettlements(data || []))
      .catch(() => {});
  };

  useEffect(() => {
    setLoading(true);
    Promise.all([
      fetch(`${API}/api/accounts`).then(r => r.json()),
    ])
      .then(([accts]) => {
        setAccounts(accts || []);
        loadTransactions('');
        loadSettlements();
        loadAllSettlements();
      })
      .catch(() => setError('Failed to load data.'))
      .finally(() => setLoading(false));
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  const handleFilterChange = (e) => {
    setFilterAccount(e.target.value);
    loadTransactions(e.target.value);
  };

  const handleSymbolBlur = async (symbol) => {
    if (!symbol) { setResolvedName(''); return; }
    setResolvingName(true);
    try {
      const res = await fetch(`${API}/api/symbol/${encodeURIComponent(symbol)}/name`);
      if (res.ok) {
        const data = await res.json();
        setResolvedName(data.name || '');
      }
    } catch { /* silent */ } finally {
      setResolvingName(false);
    }
  };

  const handleSubtypeChange = (subtype) => {
    setForm(f => {
      const next = { ...f, subtype };
      if (subtype === 'stock_buy')  next.type = 'expense';
      if (subtype === 'stock_sell') next.type = 'income';
      if (subtype === '')           next.market = 'US';
      return next;
    });
    setResolvedName('');
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      let payload = { ...form };

      if (isStockTx) {
        payload.quantity = parseFloat(form.quantity);
        payload.price    = parseFloat(form.price);
        payload.fee      = form.fee !== '' ? parseFloat(form.fee) : 0;
        payload.tax      = form.tax !== '' ? parseFloat(form.tax) : 0;
        const tradeValue = payload.quantity * payload.price;
        if (form.subtype === 'stock_buy') {
          payload.amount = -(tradeValue + (form.fee !== '' ? payload.fee : computed.fee));
        } else {
          payload.amount = tradeValue - (form.fee !== '' ? payload.fee : computed.fee)
                                      - (form.tax !== '' ? payload.tax : computed.tax);
        }
        if (!payload.category) payload.category = 'Investment';
        if (!payload.bank_account_id) delete payload.bank_account_id;
      } else {
        payload.amount = form.type === 'expense'
          ? -Math.abs(parseFloat(form.amount))
          : Math.abs(parseFloat(form.amount));
        payload.subtype  = '';
        payload.symbol   = '';
        delete payload.quantity;
        delete payload.price;
        delete payload.fee;
        delete payload.tax;
        delete payload.bank_account_id;
      }

      const res = await fetch(`${API}/api/transactions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
      if (!res.ok) throw new Error();
      setShowModal(false);
      setForm(EMPTY_FORM);
      setResolvedName('');
      loadTransactions(filterAccount);
      loadSettlements();
      loadAllSettlements();
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

  const handleSettle = async (id) => {
    try {
      const res = await fetch(`${API}/api/settlements/${id}/settle`, { method: 'POST' });
      if (!res.ok) throw new Error();
      loadSettlements();
      loadAllSettlements();
    } catch {
      alert('Failed to settle transaction.');
    }
  };

  const [expandedTx, setExpandedTx] = useState(null); // transaction id that is expanded

  // Build a lookup: transaction_id (string) → settlement record (all statuses)
  const [allSettlements, setAllSettlements] = useState([]);

  const loadAllSettlements = () => {
    fetch(`${API}/api/settlements`)
      .then(r => r.json())
      .then(data => setAllSettlements(data || []))
      .catch(() => {});
  };

  const settlementForTx = (txId) =>
    allSettlements.find(s => s.transaction_id === txId);

  const sorted = [...transactions].sort((a, b) => new Date(b.date) - new Date(a.date));

  if (loading) return <div className="loading">Loading transactions…</div>;
  if (error)   return <div className="error">{error}</div>;

  return (
    <div>
      <h1 className="page-title">Transactions</h1>

      {/* Pending settlements panel */}
      {settlements.length > 0 && (
        <div className="table-container" style={{ marginBottom: '1.5rem', borderLeft: '4px solid #ed8936' }}>
          <div className="table-header">
            <h2 style={{ color: '#c05621' }}>⏳ Pending Settlements ({settlements.length})</h2>
          </div>
          <table>
            <thead>
              <tr>
                <th>Trade Date</th>
                <th>Settlement Date</th>
                <th>Type</th>
                <th>Market</th>
                <th>Securities Account</th>
                <th>Bank Account</th>
                <th>Amount</th>
                <th>Action</th>
              </tr>
            </thead>
            <tbody>
              {settlements.map(s => (
                <tr key={s.id}>
                  <td>{s.trade_date}</td>
                  <td>{s.settlement_date}</td>
                  <td>
                    <span className={`badge ${s.trade_type === 'buy' ? 'badge-expense' : 'badge-income'}`}>
                      {s.trade_type.toUpperCase()}
                    </span>
                  </td>
                  <td>{s.market}</td>
                  <td>{accountName(s.securities_account_id)}</td>
                  <td>{s.bank_account_id ? accountName(s.bank_account_id) : <span style={{ color: '#a0aec0' }}>—</span>}</td>
                  <td className={s.trade_type === 'buy' ? 'amount-negative' : 'amount-positive'}>
                    {s.trade_type === 'buy' ? '-' : '+'}{fmt(s.amount)}
                  </td>
                  <td>
                    <button
                      className="btn btn-secondary"
                      style={{ fontSize: '0.78rem', padding: '0.2rem 0.6rem' }}
                      onClick={() => handleSettle(s.id)}
                    >
                      Settle Now
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <div className="table-container">
        <div className="table-header">
          <h2>{sorted.length} transaction{sorted.length !== 1 ? 's' : ''}</h2>
          <button className="btn btn-primary" onClick={() => setShowModal(true)}>
            + Add Transaction
          </button>
        </div>

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
                <th style={{ width: '1.5rem' }}></th>
                <th>Date</th>
                <th>Description</th>
                <th>Account</th>
                <th>Category</th>
                <th>Type</th>
                <th>Symbol</th>
                <th>Amount</th>
              </tr>
            </thead>
            <tbody>
              {sorted.map(t => {
                const isStock = t.subtype === 'stock_buy' || t.subtype === 'stock_sell';
                const expanded = expandedTx === t.id;
                const settlement = isStock ? settlementForTx(t.id) : null;
                return (
                  <React.Fragment key={t.id}>
                    <tr
                      style={{ cursor: isStock ? 'pointer' : 'default', background: expanded ? '#f7fafc' : '' }}
                      onClick={() => isStock && setExpandedTx(expanded ? null : t.id)}
                    >
                      <td style={{ textAlign: 'center', color: '#718096', fontSize: '0.75rem' }}>
                        {isStock ? (expanded ? '▼' : '▶') : ''}
                      </td>
                      <td>{t.date}</td>
                      <td>{t.description}</td>
                      <td>{accountName(t.account_id)}</td>
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
                    {expanded && (
                      <tr style={{ background: '#f7fafc' }}>
                        <td></td>
                        <td colSpan={7} style={{ paddingTop: 0, paddingBottom: '0.75rem' }}>
                          {settlement ? (
                            <div style={{
                              border: `1px solid ${settlement.status === 'settled' ? '#9ae6b4' : '#fbd38d'}`,
                              borderRadius: '6px',
                              padding: '0.6rem 1rem',
                              fontSize: '0.84rem',
                              background: settlement.status === 'settled' ? '#f0fff4' : '#fffbeb',
                              display: 'flex',
                              gap: '1.5rem',
                              flexWrap: 'wrap',
                              alignItems: 'center',
                            }}>
                              <span>
                                <strong>Settlement</strong>&nbsp;
                                <span className={`badge ${settlement.status === 'settled' ? 'badge-income' : 'badge-expense'}`}>
                                  {settlement.status}
                                </span>
                              </span>
                              <span>Trade: <strong>{settlement.trade_date}</strong></span>
                              <span>Settles: <strong>{settlement.settlement_date}</strong></span>
                              <span>Market: <strong>{settlement.market}</strong></span>
                              <span>
                                Cash {settlement.trade_type === 'buy' ? 'debit' : 'credit'}:&nbsp;
                                <strong className={settlement.trade_type === 'buy' ? 'amount-negative' : 'amount-positive'}>
                                  {settlement.trade_type === 'buy' ? '-' : '+'}{fmt(settlement.amount)}
                                </strong>
                              </span>
                              <span>
                                Bank: <strong>{settlement.bank_account_id ? accountName(settlement.bank_account_id) : '—'}</strong>
                              </span>
                              {settlement.status === 'pending' && (
                                <button
                                  className="btn btn-secondary"
                                  style={{ fontSize: '0.78rem', padding: '0.2rem 0.6rem' }}
                                  onClick={(e) => { e.stopPropagation(); handleSettle(settlement.id); }}
                                >
                                  Settle Now
                                </button>
                              )}
                            </div>
                          ) : (
                            <span style={{ color: '#a0aec0', fontSize: '0.84rem' }}>No settlement record found.</span>
                          )}
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                );
              })}
            </tbody>
          </table>
        )}
      </div>

      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal" onClick={e => e.stopPropagation()} style={{ maxWidth: '520px' }}>
            <h2>Add Transaction</h2>
            <form onSubmit={handleSubmit}>

              <div className="form-group">
                <label>Kind</label>
                <select
                  value={form.subtype}
                  onChange={e => handleSubtypeChange(e.target.value)}
                >
                  <option value="">Regular (income / expense)</option>
                  <option value="stock_buy">📈 Stock Buy</option>
                  <option value="stock_sell">📉 Stock Sell</option>
                  <option value="dividend">💰 Dividend</option>
                </select>
              </div>

              {isStockTx && (
                <>
                  <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '0 1rem' }}>
                    <div className="form-group">
                      <label>Symbol</label>
                      <input
                        required
                        value={form.symbol}
                        onChange={e => { setForm(f => ({ ...f, symbol: e.target.value.toUpperCase() })); setResolvedName(''); }}
                        onBlur={e => handleSymbolBlur(e.target.value)}
                        placeholder="e.g. AAPL"
                      />
                      {resolvingName && <small style={{ color: '#718096' }}>Looking up…</small>}
                      {resolvedName && !resolvingName && (
                        <small style={{ color: '#38a169' }}>✓ {resolvedName}</small>
                      )}
                    </div>
                    <div className="form-group">
                      <label>Market</label>
                      <select
                        value={form.market}
                        onChange={e => setForm(f => ({ ...f, market: e.target.value }))}
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
                        type="number" step="0.0001" min="0" required
                        value={form.quantity}
                        onChange={e => setForm(f => ({ ...f, quantity: e.target.value }))}
                        placeholder="0"
                      />
                    </div>
                    <div className="form-group">
                      <label>Price per Share</label>
                      <input
                        type="number" step="0.0001" min="0" required
                        value={form.price}
                        onChange={e => setForm(f => ({ ...f, price: e.target.value }))}
                        placeholder="0.00"
                      />
                    </div>
                  </div>

                  <div style={{
                    background: '#f7fafc', border: '1px solid #e8ecf0',
                    borderRadius: '6px', padding: '0.75rem 1rem', marginBottom: '1rem',
                    fontSize: '0.86rem', color: '#4a5568',
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
                        type="number" step="0.01" min="0"
                        value={form.fee}
                        onChange={e => setForm(f => ({ ...f, fee: e.target.value }))}
                        placeholder={computed.fee.toFixed(2)}
                      />
                    </div>
                    <div className="form-group">
                      <label>Tax Override (optional)</label>
                      <input
                        type="number" step="0.01" min="0"
                        value={form.tax}
                        onChange={e => setForm(f => ({ ...f, tax: e.target.value }))}
                        placeholder={computed.tax.toFixed(2)}
                      />
                    </div>
                  </div>

                  {(form.quantity && form.price) && (
                    <div style={{
                      background: form.subtype === 'stock_buy' ? '#f0fff4' : '#fff5f5',
                      border: `1px solid ${form.subtype === 'stock_buy' ? '#9ae6b4' : '#feb2b2'}`,
                      borderRadius: '6px', padding: '0.6rem 1rem', marginBottom: '1rem',
                      fontSize: '0.9rem', fontWeight: 600,
                      color: form.subtype === 'stock_buy' ? '#276749' : '#c53030',
                    }}>
                      {form.subtype === 'stock_buy'
                        ? `Total outlay: ${fmt((parseFloat(form.quantity)||0)*(parseFloat(form.price)||0) + displayFee)}`
                        : `Net proceeds: ${fmt((parseFloat(form.quantity)||0)*(parseFloat(form.price)||0) - displayFee - displayTax)}`}
                    </div>
                  )}
                </>
              )}

              {!isStockTx && (
                <>
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
                    <label>Amount (USD)</label>
                    <input
                      type="number" step="0.01" min="0" required
                      value={form.amount}
                      onChange={e => setForm(f => ({ ...f, amount: e.target.value }))}
                      placeholder="0.00"
                    />
                  </div>
                </>
              )}

              <div className="form-group">
                <label>{isStockTx ? 'Securities Account' : 'Account'}</label>
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
              {isStockTx && (
                <div className="form-group">
                  <label>Bank Account for Settlement (optional)</label>
                  <select
                    value={form.bank_account_id}
                    onChange={e => setForm(f => ({ ...f, bank_account_id: e.target.value }))}
                  >
                    <option value="">— select bank account —</option>
                    {accounts.map(a => (
                      <option key={a.id} value={a.id}>{a.name}</option>
                    ))}
                  </select>
                  <small style={{ color: '#718096' }}>
                    Cash will be debited/credited on settlement date (T+{form.market === 'US' ? 1 : 2}).
                  </small>
                </div>
              )}
              <div className="form-group">
                <label>Date</label>
                <input
                  type="date" required
                  value={form.date}
                  onChange={e => setForm(f => ({ ...f, date: e.target.value }))}
                />
              </div>
              <div className="form-group">
                <label>Description</label>
                <input
                  required={!isStockTx}
                  value={form.description}
                  onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
                  placeholder={isStockTx ? 'Auto-filled if blank' : 'e.g. Grocery Store'}
                />
              </div>
              {!isStockTx && (
                <div className="form-group">
                  <label>Category</label>
                  <input
                    required
                    value={form.category}
                    onChange={e => setForm(f => ({ ...f, category: e.target.value }))}
                    placeholder="e.g. Food, Housing, Income…"
                  />
                </div>
              )}

              <div className="form-actions">
                <button type="button" className="btn btn-secondary"
                  onClick={() => { setShowModal(false); setResolvedName(''); }}>
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
