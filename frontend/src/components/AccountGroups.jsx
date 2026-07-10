import React, { useEffect, useState, useCallback } from 'react';

const API = import.meta.env.VITE_API_URL || '';

const fmt = (n) =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' }).format(n);

const fmtPct = (n) => `${n >= 0 ? '+' : ''}${n.toFixed(2)}%`;

const EMPTY_GROUP = { name: '', description: '', account_ids: [] };

export default function AccountGroups() {
  const [groups, setGroups] = useState([]);
  const [accounts, setAccounts] = useState([]);
  const [stats, setStats] = useState({});
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [form, setForm] = useState(EMPTY_GROUP);
  const [saving, setSaving] = useState(false);
  const [editingMembers, setEditingMembers] = useState(null);
  const [memberSelection, setMemberSelection] = useState([]);

  const loadGroups = useCallback(async () => {
    const [gRes, aRes] = await Promise.all([
      fetch(`${API}/api/account-groups`),
      fetch(`${API}/api/accounts`),
    ]);
    const gs = await gRes.json();
    const as = await aRes.json();
    setGroups(gs || []);
    setAccounts(as || []);
    const statsMap = {};
    await Promise.all((gs || []).map(async g => {
      try {
        const r = await fetch(`${API}/api/account-groups/${g.id}/stats`);
        if (r.ok) statsMap[g.id] = await r.json();
      } catch {}
    }));
    setStats(statsMap);
  }, []);

  useEffect(() => {
    setLoading(true);
    loadGroups().catch(() => {}).finally(() => setLoading(false));
  }, [loadGroups]);

  const handleCreate = async (e) => {
    e.preventDefault();
    setSaving(true);
    try {
      const res = await fetch(`${API}/api/account-groups`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: form.name, description: form.description }),
      });
      if (!res.ok) throw new Error();
      setShowModal(false);
      setForm(EMPTY_GROUP);
      loadGroups();
    } catch {
      alert('Failed to create group.');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id) => {
    if (!window.confirm('Delete this group?')) return;
    await fetch(`${API}/api/account-groups/${id}`, { method: 'DELETE' });
    loadGroups();
  };

  const openMembersEditor = (group) => {
    setEditingMembers(group.id);
    setMemberSelection(group.account_ids || []);
  };

  const toggleMember = (accountId) => {
    setMemberSelection(prev =>
      prev.includes(accountId) ? prev.filter(id => id !== accountId) : [...prev, accountId]
    );
  };

  const saveMembers = async () => {
    try {
      await fetch(`${API}/api/account-groups/${editingMembers}/members`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ account_ids: memberSelection }),
      });
      setEditingMembers(null);
      loadGroups();
    } catch {
      alert('Failed to save members.');
    }
  };

  if (loading) return <div className="loading">Loading groups…</div>;

  return (
    <div>
      <h1 className="page-title">Account Groups</h1>

      <div className="table-container">
        <div className="table-header">
          <h2>Groups</h2>
          <button className="btn btn-primary" onClick={() => setShowModal(true)}>+ New Group</button>
        </div>

        {groups.length === 0 ? (
          <div className="empty">No groups yet. Create one to aggregate accounts.</div>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Name</th>
                <th>Description</th>
                <th>Accounts</th>
                <th>Total Balance</th>
                <th>Portfolio Value</th>
                <th>Return</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {groups.map(g => {
                const s = stats[g.id];
                return (
                  <tr key={g.id}>
                    <td><strong>{g.name}</strong></td>
                    <td style={{ color: '#718096' }}>{g.description || '—'}</td>
                    <td>
                      <span style={{ fontSize: '0.82rem', color: '#4a5568' }}>
                        {g.account_ids && g.account_ids.length > 0
                          ? g.account_ids.map(id => {
                              const a = accounts.find(a => String(a.id) === String(id));
                              return a ? a.name : id;
                            }).join(', ')
                          : 'None'}
                      </span>
                    </td>
                    <td>{s ? fmt(s.total_balance) : '—'}</td>
                    <td>{s ? fmt(s.portfolio_value) : '—'}</td>
                    <td>
                      {s ? (
                        <span style={{ color: s.ror >= 0 ? '#38a169' : '#e53e3e', fontWeight: 600 }}>
                          {fmtPct(s.ror)}
                        </span>
                      ) : '—'}
                    </td>
                    <td style={{ display: 'flex', gap: '0.4rem', flexWrap: 'nowrap' }}>
                      <button className="btn btn-secondary" style={{ fontSize: '0.8rem', padding: '0.25rem 0.6rem' }}
                        onClick={() => openMembersEditor(g)}>
                        Edit Members
                      </button>
                      <button className="btn btn-danger" style={{ fontSize: '0.8rem', padding: '0.25rem 0.6rem' }}
                        onClick={() => handleDelete(g.id)}>
                        Delete
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>

      {/* Create Group Modal */}
      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal" onClick={e => e.stopPropagation()} style={{ maxWidth: '420px' }}>
            <h2>New Account Group</h2>
            <form onSubmit={handleCreate}>
              <div className="form-group">
                <label>Name</label>
                <input
                  required
                  value={form.name}
                  onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                  placeholder="e.g. Retirement Accounts"
                />
              </div>
              <div className="form-group">
                <label>Description (optional)</label>
                <input
                  value={form.description}
                  onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
                  placeholder="e.g. Long-term holdings"
                />
              </div>
              <div className="form-actions">
                <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)}>Cancel</button>
                <button type="submit" className="btn btn-primary" disabled={saving}>
                  {saving ? 'Creating…' : 'Create Group'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Edit Members Modal */}
      {editingMembers && (
        <div className="modal-overlay" onClick={() => setEditingMembers(null)}>
          <div className="modal" onClick={e => e.stopPropagation()} style={{ maxWidth: '420px' }}>
            <h2>Edit Group Members</h2>
            <p style={{ color: '#718096', marginBottom: '1rem', fontSize: '0.9rem' }}>
              Select accounts to include in this group:
            </p>
            {accounts.length === 0 ? (
              <div className="empty">No accounts available.</div>
            ) : (
              <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', marginBottom: '1rem' }}>
                {accounts.map(a => (
                  <label key={a.id} style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', cursor: 'pointer' }}>
                    <input
                      type="checkbox"
                      checked={memberSelection.includes(String(a.id))}
                      onChange={() => toggleMember(String(a.id))}
                      style={{ width: '16px', height: '16px' }}
                    />
                    <span>{a.name} <span style={{ color: '#a0aec0', fontSize: '0.82rem' }}>({a.type})</span></span>
                  </label>
                ))}
              </div>
            )}
            <div className="form-actions">
              <button className="btn btn-secondary" onClick={() => setEditingMembers(null)}>Cancel</button>
              <button className="btn btn-primary" onClick={saveMembers}>Save Members</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
