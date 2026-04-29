import { Route, Routes } from "react-router-dom";
import AuthenticatedLayoutRoute from "./components/AuthenticatedLayoutRoute";
import AdminPage from "./pages/AdminPage";
import CreateExercisePage from "./pages/CreateExercisePage";
import CreateRoutinePage from "./pages/CreateRoutinePage";
import DashboardPage from "./pages/DashboardPage";
import ExercisePage from "./pages/ExercisePage";
import LoginPage from "./pages/LoginPage";
import NotFoundRedirect from "./pages/NotFoundRedirect";
import ProfilePage from "./pages/ProfilePage";
import UserRoutinesPage from "./pages/UserRoutinesPage";

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<LoginPage />} />
      <Route element={<AuthenticatedLayoutRoute />}>
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/exercises" element={<ExercisePage />} />
        <Route path="/profile" element={<ProfilePage />} />
        <Route path="/routines/new" element={<CreateRoutinePage />} />
        <Route path="/exercises/new" element={<CreateExercisePage />} />
        <Route path="/routines" element={<UserRoutinesPage />} />
        <Route path="/admin" element={<AdminPage />} />
      </Route>
      <Route path="*" element={<NotFoundRedirect />} />
    </Routes>
  );
}
