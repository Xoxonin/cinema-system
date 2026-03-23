import { useEffect } from 'react'
import './index.css'

function App() {
  useEffect(() => {
    // This will fetch from the Catalog service once the environment is set up.
    // fetch('/api/movies') ...
  }, [])

  return (
    <div className="w-full min-h-screen text-white bg-gray-900 p-8 font-sans">
      <header className="max-w-6xl mx-auto flex justify-between items-center bg-gray-800 p-6 rounded-xl shadow-lg mb-12 border border-gray-700">
        <h1 className="text-3xl font-extrabold tracking-tight bg-gradient-to-r from-purple-500 to-pink-500 bg-clip-text text-transparent">Nexus Cinema</h1>
        <nav className="flex space-x-6 text-lg font-medium">
          <a href="#" className="hover:text-purple-400 transition-colors">Movies</a>
          <a href="#" className="hover:text-pink-400 transition-colors">My Tickets</a>
          <a href="#" className="px-4 py-2 bg-gray-700 hover:bg-gray-600 rounded-lg transition-colors">Login</a>
        </nav>
      </header>
      
      <main className="max-w-6xl mx-auto grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-8">
        {/* Placeholder Movie cards */}
        {[1, 2, 3, 4].map(i => (
          <div key={i} className="bg-gray-800 p-5 rounded-xl shadow-md hover:shadow-2xl hover:-translate-y-1 transition-all duration-300 border border-gray-700 flex flex-col">
             <div className="w-full aspect-[2/3] bg-gray-700 rounded-lg mb-4 bg-gradient-to-t from-gray-800 to-transparent"></div>
             <h2 className="text-2xl font-bold mb-1">Movie Title {i}</h2>
             <p className="text-sm text-pink-500 font-medium mb-4">120 mins • Action</p>
             <div className="mt-auto">
               <button className="w-full bg-gradient-to-r from-purple-600 to-pink-600 hover:from-purple-500 hover:to-pink-500 text-white font-bold py-3 px-4 rounded-lg shadow-lg hover:shadow-pink-500/25 transition-all outline-none focus:ring-2 focus:ring-pink-400">
                 Book Ticket
               </button>
             </div>
          </div>
        ))}
      </main>
    </div>
  )
}

export default App
