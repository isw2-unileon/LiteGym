import { BrowserRouter, Routes, Route } from "react-router-dom";
import LoginPage from "./pages/LoginPage";
import ExercisePage from "./pages/ExercisePage";

export default function App() {
    return (
        <BrowserRouter>
            <Routes>
                <Route path="/" element={<LoginPage />} />
                <Route path="/exercises" element={<ExercisePage />} />
            </Routes>
        </BrowserRouter>
    );
}
