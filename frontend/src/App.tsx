import { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Link, useNavigate } from 'react-router-dom';
import { Home } from './Home';
import { Login } from './Login';
import { Register } from './Register';
import { Profile } from './Profile';
import './index.css';

function Navbar({ token, setAuthToken }: { token: string | null, setAuthToken: (t: string | null) => void }) {
  const navigate = useNavigate();
  
  const handleLogout = () => {
    setAuthToken(null);
    navigate('/');
  };

  return (
    <header className="max-w-6xl mx-auto flex justify-between items-center bg-gray-800 p-6 rounded-xl shadow-lg mb-12 border border-gray-700">
      <Link to="/" className="text-3xl font-extrabold tracking-tight bg-gradient-to-r from-purple-500 to-pink-500 bg-clip-text text-transparent">
        Nexus Cinema
      </Link>
      <nav className="flex space-x-6 text-lg font-medium items-center">
        <Link to="/" className="hover:text-purple-400 transition-colors">Movies</Link>
        {token ? (
          <>
            <Link to="#" className="hover:text-pink-400 transition-colors">My Tickets</Link>
            <Link to="/profile" className="hover:text-pink-400 transition-colors">Profile</Link>
            <button onClick={handleLogout} className="px-4 py-2 bg-red-500/20 text-red-400 hover:bg-red-500 hover:text-white rounded-lg transition-colors border border-red-500 hover:border-transparent">Logout</button>
          </>
        ) : (
          <>
            <Link to="/login" className="hover:text-purple-400 transition-colors">Login</Link>
            <Link to="/register" className="px-4 py-2 bg-gradient-to-r from-purple-600 to-pink-600 hover:from-purple-500 hover:to-pink-500 text-white rounded-lg transition-colors shadow">Register</Link>
          </>
        )}
      </nav>
    </header>
  );
}

function App() {
  const [token, setToken] = useState<string | null>(localStorage.getItem('token'));

  const setAuthToken = (t: string | null) => {
    if (t) {
      localStorage.setItem('token', t);
    } else {
      localStorage.removeItem('token');
    }
    setToken(t);
  };

  return (
    <Router>
      <div className="w-full min-h-screen text-white bg-gray-900 p-8 font-sans">
        <Navbar token={token} setAuthToken={setAuthToken} />
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/login" element={<Login setAuthToken={setAuthToken} />} />
          <Route path="/register" element={<Register />} />
          <Route path="/profile" element={<Profile token={token} setAuthToken={setAuthToken} />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
