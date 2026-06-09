import { useEffect, useState } from 'react';
import { useLocation, useNavigate, useParams } from 'react-router-dom';

interface TicketState {
  seat: string;
  type: 'Normal' | 'Reduced';
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

export function TicketSelection({ token }: { token: string | null }) {
  const { showtime_id } = useParams<{ showtime_id: string }>();
  const navigate = useNavigate();
  const location = useLocation();
  const seats: string[] = location.state?.seats || [];
  
  const [tickets, setTickets] = useState<TicketState[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!token) {
       navigate('/login');
       return;
    }
    if (seats.length === 0) {
       navigate(`/book/${showtime_id}/seats`);
       return;
    }
    setTickets(seats.map(s => ({ seat: s, type: 'Normal' })));
  }, [seats, token, navigate, showtime_id]);

  const setTicketType = (seat: string, type: 'Normal' | 'Reduced') => {
    setTickets(prev => prev.map(t => t.seat === seat ? { ...t, type } : t));
  };

  const confirmBooking = async () => {
    const payload = parseJwt(token!);
    if (!payload?.userId) {
       setError("Invalid session");
       return;
    }
    setLoading(true);
    
    // We need to fetch the user's locked bookings to get the booking IDs to confirm
    try {
       const userBookings = await fetch(`/api/bookings?user_id=${payload.userId}`).then(res => res.json());
       
       let successCount = 0;
       for (const t of tickets) {
          // Find the locked booking for this showtime and seat
          const bk = userBookings?.find((b: any) => 
               b.showtime_id === parseInt(showtime_id!) && 
               b.seat_number === t.seat && 
               b.status === 'locked'
          );
          
          if (!bk) {
             setError(`Lock expired or missing for seat ${t.seat}`);
             continue;
          }
          
          const res = await fetch('/api/bookings/confirm', {
             method: 'POST',
             headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
             body: JSON.stringify({
                booking_id: bk.id,
                user_id: payload.userId,
                ticket_type: t.type
             })
          });
          if (res.ok) successCount++;
       }
       
       if (successCount === tickets.length) {
          navigate('/my-tickets', { state: { success: true } });
       } else {
          setError("Failed to confirm some tickets. Locks might have expired.");
       }
    } catch (err) {
       setError("Error during confirmation.");
    } finally {
       setLoading(false);
    }
  };

  return (
    <main className="max-w-3xl mx-auto flex flex-col items-center">
      <h1 className="text-3xl font-extrabold mb-8 text-center bg-gradient-to-r from-yellow-400 to-orange-400 bg-clip-text text-transparent">
        Select Ticket Types
      </h1>

      {error && <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-2 rounded mb-6 w-full text-center shadow-lg">{error}</div>}

      <div className="w-full bg-gray-800 p-8 rounded-3xl shadow-2xl border border-gray-700">
         <div className="space-y-6">
            {tickets.map(t => (
               <div key={t.seat} className="flex justify-between items-center p-4 bg-gray-900 rounded-xl border border-gray-800">
                  <div className="flex gap-4 items-center">
                    <div className="w-12 h-12 bg-yellow-500/20 text-yellow-400 rounded-lg flex items-center justify-center font-bold text-xl border border-yellow-500/50">
                       {t.seat.split('-')[1]}
                    </div>
                    <div>
                       <div className="text-sm text-gray-500 uppercase tracking-wider">Row {t.seat.split('-')[0]}</div>
                       <div className="text-lg font-bold text-white">Seat {t.seat.split('-')[1]}</div>
                    </div>
                  </div>
                  <div className="flex bg-gray-800 rounded-lg p-1 border border-gray-700">
                     <button 
                        onClick={() => setTicketType(t.seat, 'Normal')}
                        className={`px-4 py-2 rounded-md text-sm font-semibold transition-all ${t.type === 'Normal' ? 'bg-yellow-500 text-gray-900 outline-none shadow-lg' : 'text-gray-400 hover:text-white hover:bg-gray-700'}`}
                     >Normal</button>
                     <button 
                        onClick={() => setTicketType(t.seat, 'Reduced')}
                        className={`px-4 py-2 rounded-md text-sm font-semibold transition-all ${t.type === 'Reduced' ? 'bg-orange-500 text-white outline-none shadow-lg' : 'text-gray-400 hover:text-white hover:bg-gray-700'}`}
                     >Reduced</button>
                  </div>
               </div>
            ))}
         </div>
         
         <div className="mt-12 flex justify-between items-center border-t border-gray-700 pt-6">
            <div className="text-gray-400">
               {tickets.length} tickets total
            </div>
            <button 
               onClick={confirmBooking} 
               disabled={loading}
               className="px-8 py-4 bg-gradient-to-r from-yellow-500 to-orange-500 hover:from-yellow-400 hover:to-orange-400 text-white font-bold rounded-xl shadow-lg hover:shadow-orange-500/30 transition-all active:scale-95 disabled:opacity-50"
            >
               {loading ? 'Confirming...' : 'Confirm Booking'}
            </button>
         </div>
      </div>
    </main>
  );
}
