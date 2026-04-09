import { BrowserRouter, Routes, Route, Link } from "react-router-dom";
import Profile from "./components/Profile"; 

function Home() {
  return (
    <div className="text-center">
      <h1 className="text-4xl font-bold text-gray-800 mb-4">LiteGym</h1>
      <p className="text-gray-500 mb-8">Bienvenido a LiteGym</p>
      
      {}
      <Link 
        to="/profile" 
        className="inline-block px-6 py-2 bg-blue-600 text-white font-medium rounded-lg shadow hover:bg-blue-700 transition duration-200"
      >
        Ir a Mi Perfil
      </Link>
    </div>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <main className="min-h-screen bg-gray-50 flex items-center justify-center p-4">
        <div className="w-full max-w-md">
          
          {}
          <Routes>
            {}
            <Route path="/" element={<Home />} />
            
            {}
            <Route 
              path="/profile" 
              element={
                <div>
                  {}
                  <div className="mb-6">
                    <Link to="/" className="text-blue-500 hover:text-blue-700 font-medium flex items-center gap-2">
                      <span>Volver al inicio</span>
                    </Link>
                  </div>
                  
                  {}
                  <Profile />
                </div>
              } 
            />
          </Routes>

        </div>
      </main>
    </BrowserRouter>
  );
}
