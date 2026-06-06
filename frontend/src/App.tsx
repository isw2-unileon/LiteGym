import { Route, Routes } from "react-router-dom";
import AuthenticatedLayoutRoute from "./components/AuthenticatedLayoutRoute";
import PublicOnlyRoute from "./components/PublicOnlyRoute";
import AdminPage from "./pages/AdminPage";
import DashboardPage from "./pages/DashboardPage";
import ExercisePage from "./pages/ExercisePage";
import LoginPage from "./pages/LoginPage";
import NotFoundRedirect from "./pages/NotFoundRedirect";
import ProfilePage from "./pages/ProfilePage";
import RegisterPage from "./pages/RegisterPage";
import UserRoutinesPage from "./pages/UserRoutinesPage";
import SupportPage from "./pages/SupportPage";

export default function App() {
  return (
    <Routes>
      <Route element={<PublicOnlyRoute />}>
        <Route path="/" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
      </Route>
      <Route element={<AuthenticatedLayoutRoute />}>
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/exercises" element={<ExercisePage />} />
        <Route path="/profile" element={<ProfilePage />} />
        <Route path="/routines" element={<UserRoutinesPage />} />
        <Route path="/admin" element={<AdminPage />} />
        <Route path="/support" element={<SupportPage />} />
        
      </Route>
      <Route path="*" element={<NotFoundRedirect />} />
    </Routes>
  );
}
