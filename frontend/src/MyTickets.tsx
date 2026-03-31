import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

interface Booking {
  id: number;
  user_id: number;
  showtime_id: number;
  seat_number: string;
  status: string;
  created_at: { Time: string; Valid: boolean };
  ticket_type: { String: string; Valid: boolean };
  showtime?: Showtime;
  movie?: any;
  room?: any;
}

interface GroupedBooking {
  showtime_id: number;
  showtime: Showtime;
  movie: any;
  tickets: {
    id: number;
    seat_number: string;
    ticket_type: string;
  }[];
}

interface Showtime {
  id: number;
  movie_id: number;
  room_id: number;
  start_time: string;
  end_time: string;
  type: string;
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

export function MyTickets({ token }: { token: string | null }) {
  const navigate = useNavigate();
  const [upcoming, setUpcoming] = useState<GroupedBooking[]>([]);
  const [past, setPast] = useState<GroupedBooking[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!token) {
       navigate('/login');
       return;
    }
    const payload = parseJwt(token);
    if (!payload?.userId) return;

    const fetchTickets = async () => {
       try {
          const bookings: Booking[] = await fetch(`/api/bookings?user_id=${payload.userId}`).then(res => res.json());
          if (!bookings) {
             setLoading(false);
             return;
          }
          // Fetch showtime details in parallel
          const bookingsWithShowtimes = await Promise.all(
             bookings.filter(b => b.status === 'reserved').map(async (bk) => {
                const st = await fetch(`/api/showtimes/${bk.showtime_id}`).then(res => res.json());
                const mv = await fetch(`/api/movies/${st.movie_id}`).then(res => res.json());
                // In a perfect system we'd join on backend, but here we do frontend join based on constraints
                return { ...bk, showtime: st, movie: mv };
             })
          );
          const groupedMap = new Map<number, GroupedBooking>();
          
          bookingsWithShowtimes.forEach(bk => {
             if (!groupedMap.has(bk.showtime_id)) {
                groupedMap.set(bk.showtime_id, {
                   showtime_id: bk.showtime_id,
                   showtime: bk.showtime!,
                   movie: bk.movie,
                   tickets: []
                });
             }
             groupedMap.get(bk.showtime_id)!.tickets.push({
                id: bk.id,
                seat_number: bk.seat_number,
                ticket_type: bk.ticket_type?.String || 'Unknown'
             });
          });

          const now = new Date('2026-03-31T11:44:03+02:00').getTime(); // System date
          const coming: GroupedBooking[] = [];
          const history: GroupedBooking[] = [];
          
          groupedMap.forEach(group => {
             const startTime = new Date(group.showtime.start_time).getTime();
             if (startTime > now) coming.push(group);
             else history.push(group);
          });
          
          setUpcoming(coming.sort((a,b) => new Date(a.showtime.start_time).getTime() - new Date(b.showtime.start_time).getTime()));
          setPast(history.sort((a,b) => new Date(b.showtime.start_time).getTime() - new Date(a.showtime.start_time).getTime()));
       } catch (err) {
          console.error("Failed to fetch tickets", err);
       } finally {
          setLoading(false);
       }
    };
    
    fetchTickets();
  }, [token, navigate]);

  if (loading) return <div className="text-center text-gray-400 py-12 animate-pulse">Loading tickets...</div>;

  return (
    <main className="max-w-5xl mx-auto p-4">
      <h1 className="text-4xl font-black mb-12 text-center bg-gradient-to-r from-purple-400 to-pink-400 bg-clip-text text-transparent">
        My Tickets
      </h1>
      
      <div className="space-y-12">
        <section>
           <h2 className="text-2xl font-bold mb-6 flex items-center gap-3">
              <div className="w-3 h-8 bg-purple-500 rounded-full"></div>
              Upcoming Shows
           </h2>
           {upcoming.length === 0 ? (
              <div className="bg-gray-800/50 rounded-2xl p-8 text-center text-gray-500 italic">No upcoming tickets</div>
           ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                 {upcoming.map(group => (
                    <div key={group.showtime_id} className="bg-gradient-to-br from-gray-800 to-gray-900 border border-purple-500/30 p-6 rounded-2xl shadow-xl hover:shadow-purple-500/20 transition-all">
                       <h3 className="text-2xl font-bold mb-2 truncate" title={group.movie?.title}>{group.movie?.title}</h3>
                       <div className="text-pink-400 font-bold mb-4 opacity-90 inline-block px-3 py-1 bg-pink-500/10 rounded-lg">
                          {new Date(group.showtime.start_time).toLocaleString()}
                       </div>
                       
                       <div className="space-y-3 mt-4">
                          <div className="text-sm font-semibold text-gray-400 uppercase tracking-widest border-b border-gray-700 pb-2 mb-3">Reserved Seats</div>
                          {group.tickets.map(t => (
                             <div key={t.id} className="flex justify-between items-center text-sm font-bold p-3 bg-gray-800 rounded-xl border border-gray-700/50 hover:border-gray-600 transition-colors">
                                <span className="text-gray-300">Seat <span className="text-white text-lg ml-2">{t.seat_number}</span></span>
                                <span className="bg-blue-500/20 text-blue-300 px-3 py-1 rounded-full text-xs">{t.ticket_type}</span>
                             </div>
                          ))}
                       </div>
                       <div className="mt-6 text-xs text-gray-500 text-center uppercase tracking-widest border-t border-gray-700 pt-3 opacity-50">
                         {group.showtime.type}
                       </div>
                    </div>
                 ))}
              </div>
           )}
        </section>

        <section>
           <h2 className="text-2xl font-bold mb-6 flex items-center gap-3">
              <div className="w-3 h-8 bg-gray-600 rounded-full"></div>
              Past Shows
           </h2>
           {past.length === 0 ? (
              <div className="bg-gray-800/50 rounded-2xl p-8 text-center text-gray-500 italic">No past tickets</div>
           ) : (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-8 opacity-75 grayscale hover:grayscale-0 transition-all">
                 {past.map(group => (
                    <div key={group.showtime_id} className="bg-gray-800 border-none p-6 rounded-2xl shadow-inner flex flex-col items-center">
                       <h3 className="text-xl font-bold mb-1 truncate w-full text-center text-gray-400">{group.movie?.title}</h3>
                       <div className="text-gray-500 text-sm mb-4">
                          {new Date(group.showtime.start_time).toLocaleDateString()}
                       </div>
                       <div className="w-full space-y-2 mt-2">
                          {group.tickets.map(t => (
                             <div key={t.id} className="w-full flex justify-between items-center text-xs font-semibold p-2 bg-gray-900 rounded border border-gray-700">
                                <span className="text-gray-500">Seat {t.seat_number}</span>
                                <span className="text-gray-500 bg-gray-800 px-2 py-1 rounded">{t.ticket_type}</span>
                             </div>
                          ))}
                       </div>
                    </div>
                 ))}
              </div>
           )}
        </section>
      </div>
    </main>
  );
}
