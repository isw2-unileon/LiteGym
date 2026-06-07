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
import VerifyEmailPage from "./pages/VerifyEmailPage";
import ForgotPasswordPage from "./pages/ForgotPasswordPage";
import ResetPasswordPage from "./pages/ResetPasswordPage";
import UserRoutinesPage from "./pages/UserRoutinesPage";
import SupportPage from "./pages/SupportPage";
import LegalNoticePage from "@/pages/LegalNoticePage.tsx";
import PrivacyPolicyPage from "@/pages/PrivacyPolicyPage.tsx";
import CookiesPolicyPage from "@/pages/CookiesPolicyPage.tsx";
import TermsPage from "@/pages/TermsPage.tsx";

export default function App() {
  return (
    <Routes>
      <Route element={<PublicOnlyRoute />}>
        <Route path="/" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/forgot-password" element={<ForgotPasswordPage />} />
      </Route>
      <Route path="/verify-email" element={<VerifyEmailPage />} />
      <Route path="/reset-password" element={<ResetPasswordPage />} />
      <Route element={<AuthenticatedLayoutRoute />}>
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/exercises" element={<ExercisePage />} />
        <Route path="/profile" element={<ProfilePage />} />
        <Route path="/admin" element={<AdminPage />} />
        <Route path="/support" element={<SupportPage />} />
        <Route path="/routines" element={<UserRoutinesPage />} />
      </Route>
        <Route path="/legal/aviso-legal" element={<LegalNoticePage />}/>
        <Route path="/legal/privacidad" element={<PrivacyPolicyPage />}/>
        <Route path="/legal/cookies" element={<CookiesPolicyPage />}/>
        <Route path="/legal/terminos" element={<TermsPage />}/>
      <Route path="*" element={<NotFoundRedirect />} />
    </Routes>
  );
}
