import { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';

interface SeatLayout {
  row: number;
  seats: number;
}

interface Room {
  id: number;
  name: string;
  capacity: number;
  seating_layout: SeatLayout[];
}

interface Showtime {
  id: number;
  movie_id: number;
  room_id: number;
}

interface Booking {
  id: number;
  seat_number: string;
  status: string;
  locked_until?: { Time: string; Valid: boolean };
}

interface JwtPayload {
  userId: number;
}

function parseJwt(token: string): JwtPayload | null {
  try {
    return JSON.parse(atob(token.split('.')[1]));
  } catch (e) {
    return null;
  }
}

export function SeatSelection({ token }: { token: string | null }) {
  const { showtime_id } = useParams<{ showtime_id: string }>();
  const navigate = useNavigate();
  const [layout, setLayout] = useState<SeatLayout[]>([]);
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedSeats, setSelectedSeats] = useState<string[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!token) {
      navigate('/login');
      return;
    }
    Promise.all([
      fetch(`/api/showtimes/${showtime_id}`).then(res => res.json()),
      fetch(`/api/rooms`).then(res => res.json()),
      fetch(`/api/bookings?showtime_id=${showtime_id}`).then(res => res.json())
    ]).then(([showtime, allRooms, bks]: [Showtime, Room[], Booking[]]) => {
      const rm = allRooms.find(r => r.id === showtime.room_id);
      if (rm) setLayout(rm.seating_layout || []);
      setBookings(bks || []);
      setLoading(false);
    }).catch(err => {
      console.error(err);
      setError("Failed to load seat layout");
      setLoading(false);
    });
  }, [showtime_id, token, navigate]);

  const isSeatTaken = (r: number, s: number) => {
    const sn = `${r}-${s}`;
    const bk = bookings.find(b => b.seat_number === sn);
    if (!bk) return false;
    if (bk.status === 'reserved') return true;
    if (bk.status === 'locked' && bk.locked_until?.Valid) {
       const lockTime = new Date(bk.locked_until.Time).getTime();
       if (lockTime > Date.now()) return true;
    }
    return false;
  };

  const toggleSeat = (r: number, s: number) => {
    if (isSeatTaken(r, s)) return;
    const sn = `${r}-${s}`;
    setSelectedSeats(prev => prev.includes(sn) ? prev.filter(x => x !== sn) : [...prev, sn]);
  };

  const proceedToTickets = async () => {
    if (selectedSeats.length === 0) return;
    const payload = parseJwt(token!);
    if (!payload?.userId) {
       setError("Invalid session");
       return;
    }
    setLoading(true);
    let successCount = 0;
    try {
      for (const seat of selectedSeats) {
        const res = await fetch('/api/bookings/lock', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
          body: JSON.stringify({ user_id: payload.userId, showtime_id: parseInt(showtime_id!), seat_number: seat })
        });
        if (res.ok) successCount++;
      }
      if (successCount === selectedSeats.length) {
         navigate(`/book/${showtime_id}/confirm`, { state: { seats: selectedSeats }});
      } else {
         setError("Some seats are no longer available. Please select again.");
         // Refresh bookings
         const updatedStr = await fetch(`/api/bookings?showtime_id=${showtime_id}`).then(r => r.json());
         setBookings(updatedStr || []);
         setSelectedSeats([]);
      }
    } catch(err) {
      setError("Error locking seats.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="max-w-5xl mx-auto flex flex-col items-center">
      <h1 className="text-3xl font-extrabold mb-8 text-center bg-gradient-to-r from-purple-400 to-pink-400 bg-clip-text text-transparent">
        Select Your Seats
      </h1>
      
      {error && <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-2 rounded mb-6 w-full max-w-xl text-center shadow-lg">{error}</div>}
      
      {loading ? (
         <div className="text-gray-400 animate-pulse">Loading seats...</div>
      ) : (
         <div className="bg-gray-800 p-8 rounded-3xl shadow-2xl border border-gray-700 w-full overflow-x-auto">
           <div className="w-4/5 mx-auto bg-gradient-to-b from-gray-400/20 to-transparent h-4 rounded-[100%] shadow-[0_15px_30px_rgba(255,255,255,0.05)] mb-12 flex justify-center sticky top-0">
              <span className="text-gray-500 text-xs tracking-[0.3em] uppercase mt-2">Screen</span>
           </div>
           
           <div className="flex flex-col items-center gap-4">
             {layout.map((rowLayout) => (
                <div key={rowLayout.row} className="flex gap-3 justify-center">
                  <span className="w-8 text-center text-gray-500 font-mono self-center text-sm">{rowLayout.row}</span>
                  {Array.from({length: rowLayout.seats}).map((_, i) => {
                     const s = i + 1;
                     const taken = isSeatTaken(rowLayout.row, s);
                     const sn = `${rowLayout.row}-${s}`;
                     const selected = selectedSeats.includes(sn);
                     
                     return (
                        <button 
                           key={s} 
                           disabled={taken}
                           onClick={() => toggleSeat(rowLayout.row, s)}
                           className={`w-10 h-10 rounded-t-xl rounded-b sm flex items-center justify-center text-xs font-bold transition-all shadow-md
                             ${taken ? 'bg-gray-700 text-gray-600 cursor-not-allowed shadow-inner' : 
                               selected ? 'bg-pink-500 text-white shadow-pink-500/50 scale-110' : 
                               'bg-purple-600 hover:bg-purple-500 text-purple-100 hover:-translate-y-1'}`}
                        >
                          {s}
                        </button>
                     );
                  })}
                </div>
             ))}
           </div>
           
           <div className="flex justify-center gap-8 mt-12 bg-gray-900/50 p-4 rounded-xl max-w-lg mx-auto">
              <div className="flex items-center gap-2">
                <div className="w-6 h-6 rounded bg-purple-600 shadow"></div>
                <span className="text-sm text-gray-400">Available</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-6 h-6 rounded bg-pink-500 shadow"></div>
                <span className="text-sm text-gray-400">Selected</span>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-6 h-6 rounded bg-gray-700 shadow-inner"></div>
                <span className="text-sm text-gray-400">Taken</span>
              </div>
           </div>
         </div>
      )}
      
      {selectedSeats.length > 0 && (
         <div className="fixed bottom-0 left-0 w-full bg-gray-900/95 backdrop-blur-md border-t border-gray-800 p-6 flex justify-between items-center shadow-[0_-10px_40px_rgba(0,0,0,0.5)] z-50 animate-in slide-in-from-bottom">
           <div className="max-w-6xl mx-auto w-full flex justify-between items-center">
             <div className="flex flex-col">
               <span className="text-gray-400 text-sm">Seats selected:</span>
               <span className="text-2xl font-bold bg-gradient-to-r from-purple-400 to-pink-400 bg-clip-text text-transparent">
                 {selectedSeats.length} ({selectedSeats.join(', ')})
               </span>
             </div>
             <button 
                onClick={proceedToTickets} 
                disabled={loading}
                className="px-8 py-4 bg-gradient-to-r from-purple-600 to-pink-600 hover:from-purple-500 hover:to-pink-500 text-white font-bold rounded-xl shadow-lg hover:shadow-pink-500/30 transition-all active:scale-95 disabled:opacity-50"
             >
                {loading ? 'Locking...' : 'Proceed to Tickets'}
             </button>
           </div>
         </div>
      )}
    </main>
  );
}
