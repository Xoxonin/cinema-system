import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';

interface Showtime {
  id: number;
  movie_id: number;
  room_id: number;
  start_time: string;
  end_time: string;
  type: string;
}

export function ShowtimeSelection() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [showtimes, setShowtimes] = useState<Showtime[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch(`/api/showtimes?movie_id=${id}`)
      .then(res => res.json())
      .then(data => {
        setShowtimes(data || []);
        setLoading(false);
      })
      .catch(err => {
        console.error("Failed to fetch showtimes", err);
        setLoading(false);
      });
  }, [id]);

  return (
    <main className="max-w-4xl mx-auto">
      <h1 className="text-3xl font-extrabold mb-8 text-center bg-gradient-to-r from-yellow-400 to-orange-400 bg-clip-text text-transparent">
        Select Showtime
      </h1>
      
      {loading ? (
        <div className="text-center text-gray-400 py-12 text-xl animate-pulse">Loading showtimes...</div>
      ) : showtimes.length === 0 ? (
        <div className="text-center text-gray-400 py-12 text-xl bg-gray-800 rounded-xl border border-gray-700">No showtimes available for this movie.</div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {showtimes.map(st => {
            const dateObj = new Date(st.start_time);
            return (
              <div key={st.id} className="bg-gray-800 p-6 rounded-2xl shadow-lg border border-gray-700 hover:border-yellow-500 transition-all duration-300 flex flex-col items-center group cursor-pointer" onClick={() => navigate(`/book/${st.id}/seats`)}>
                <div className="text-yellow-400 font-bold mb-2">
                  {dateObj.toLocaleDateString(undefined, { weekday: 'short', month: 'short', day: 'numeric' })}
                </div>
                <div className="text-4xl font-black text-white mb-4 group-hover:scale-110 transition-transform">
                  {dateObj.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                </div>
                <span className="bg-orange-500/20 text-orange-300 px-3 py-1 rounded-full text-sm font-semibold border border-orange-500/30">
                  {st.type || "2D"}
                </span>
              </div>
            );
          })}
        </div>
      )}
    </main>
  );
}
