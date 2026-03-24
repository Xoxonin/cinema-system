import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

interface UserProfile {
  id: number;
  username: string;
  email: string;
  role: string;
}

export function Profile({ token, setAuthToken }: { token: string | null, setAuthToken: (t: string | null) => void }) {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    if (!token) {
      navigate('/login');
      return;
    }
    fetch('/api/users/profile', {
      headers: { 'Authorization': `Bearer ${token}` }
    })
      .then(res => {
        if (!res.ok) throw new Error('Failed to fetch profile. Token may be invalid.');
        return res.json();
      })
      .then(data => setProfile(data))
      .catch(err => {
        setError(err.message);
        setAuthToken(null);
        navigate('/login');
      });
  }, [token, navigate, setAuthToken]);

  const handleLogout = () => {
    setAuthToken(null);
    navigate('/');
  };

  if (error) return <div className="text-center text-red-500 mt-12">{error}</div>;
  if (!profile) return <div className="text-center text-gray-400 mt-12">Loading profile...</div>;

  return (
    <div className="max-w-md mx-auto bg-gray-800 p-8 rounded-xl shadow-md border border-gray-700 mt-12">
      <h2 className="text-3xl font-bold mb-6 text-white text-center">My Profile</h2>
      <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 mb-8 text-gray-300">
        <p className="mb-2"><span className="font-bold text-gray-400">Username:</span> {profile.username}</p>
        <p className="mb-2"><span className="font-bold text-gray-400">Email:</span> {profile.email}</p>
        <p><span className="font-bold text-gray-400">Role:</span> {profile.role}</p>
      </div>
      <button 
        onClick={handleLogout}
        className="w-full bg-red-500/20 hover:bg-red-500 text-red-400 hover:text-white font-bold py-3 rounded transition-colors border border-red-500 hover:border-transparent outline-none focus:ring-2 focus:ring-red-400"
      >
        Logout
      </button>
    </div>
  );
}
