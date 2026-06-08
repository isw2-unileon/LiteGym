import LegalPageShell from "../components/LegalPageShell";

export default function CookiesPolicyPage() {
  return (
    <LegalPageShell
      title="Política de cookies"
      subtitle="LiteGym utiliza únicamente cookies técnicas necesarias para que la autenticación y la sesión funcionen correctamente."
      updatedAt="Mayo de 2026"
    >
      <h2>1. Cookies utilizadas</h2>
      <p>
        La app usa cookies técnicas y de sesión para mantener la autenticación del usuario y
        recordar su acceso entre peticiones.
      </p>
      <ul>
        <li>Cookie de autenticación: necesaria para mantener la sesión iniciada.</li>
        <li>Cookies o datos técnicos equivalentes: necesarios para seguridad, navegación y funcionamiento básico.</li>
      </ul>

      <h2>2. Consentimiento</h2>
      <p>
        Al tratarse solo de cookies esenciales, no se muestra un banner de consentimiento en este
        momento. Si en el futuro se añaden cookies de analítica, publicidad o marketing, deberá
        incorporarse un sistema de consentimiento específico.
      </p>

      <h2>3. Gestión</h2>
      <p>
        El usuario puede borrar o bloquear cookies desde su navegador, aunque eso puede impedir que
        la autenticación funcione correctamente.
      </p>
    </LegalPageShell>
  );
}
