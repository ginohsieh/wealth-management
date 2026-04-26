import React, { useState } from 'react';
import Dashboard from './components/Dashboard.jsx';
import Accounts from './components/Accounts.jsx';
import Transactions from './components/Transactions.jsx';
import Portfolio from './components/Portfolio.jsx';
import NetWorthHistory from './components/NetWorthHistory.jsx';

const TABS = [
  { id: 'dashboard', label: '📊 Dashboard' },
  { id: 'accounts', label: '🏦 Accounts' },
  { id: 'transactions', label: '💳 Transactions' },
  { id: 'portfolio', label: '📈 Portfolio' },
  { id: 'history', label: '📅 History' },
];

function App() {
  const [activeTab, setActiveTab] = useState('dashboard');

  const renderContent = () => {
    switch (activeTab) {
      case 'dashboard':    return <Dashboard />;
      case 'accounts':     return <Accounts />;
      case 'transactions': return <Transactions />;
      case 'portfolio':    return <Portfolio />;
      case 'history':      return <NetWorthHistory />;
      default:             return <Dashboard />;
    }
  };

  return (
    <div className="app">
      <header className="app-header">
        <div className="header-brand">
          <span className="header-icon">💰</span>
          <h1>Wealth Management</h1>
        </div>
        <nav className="app-nav">
          {TABS.map(tab => (
            <button
              key={tab.id}
              className={`nav-btn ${activeTab === tab.id ? 'active' : ''}`}
              onClick={() => setActiveTab(tab.id)}
            >
              {tab.label}
            </button>
          ))}
        </nav>
      </header>
      <main className="app-main">{renderContent()}</main>
    </div>
  );
}

export default App;
