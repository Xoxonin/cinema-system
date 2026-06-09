import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';

export function Login({ setAuthToken }: { setAuthToken: (t: string) => void }) {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const navigate = useNavigate();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const res = await fetch('/api/users/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password, username: "" })
      });
      if (res.ok) {
        const data = await res.json();
        setAuthToken(data.token);
        navigate('/');
      } else {
        const text = await res.text();
        setError(text);
      }
    } catch (err: any) {
      setError(err.message);
    }
  };

  return (
    <div className="max-w-md mx-auto bg-gray-800 p-8 rounded-xl shadow-md border border-gray-700 mt-12">
      <h2 className="text-3xl font-bold mb-6 text-center text-white">Login</h2>
      {error && <div className="bg-red-500/20 text-red-500 p-3 rounded mb-4">{error}</div>}
      <form onSubmit={handleLogin} className="flex flex-col space-y-4">
        <input 
          type="email" placeholder="Email" required 
          className="p-3 rounded bg-gray-900 border border-gray-700 text-white focus:outline-none focus:ring-2 focus:ring-yellow-500"
          value={email} onChange={e => setEmail(e.target.value)} 
        />
        <input 
          type="password" placeholder="Password" required 
          className="p-3 rounded bg-gray-900 border border-gray-700 text-white focus:outline-none focus:ring-2 focus:ring-yellow-500"
          value={password} onChange={e => setPassword(e.target.value)} 
        />
        <button type="submit" className="bg-gradient-to-r from-yellow-500 to-orange-500 hover:from-yellow-400 hover:to-orange-400 text-white font-bold py-3 rounded shadow hover:shadow-lg transition-all">
          Sign In
        </button>
      </form>
      <p className="mt-6 text-center text-gray-400">
        Don't have an account? <Link to="/register" className="text-orange-400 hover:text-orange-300">Register</Link>
      </p>
    </div>
  );
}
