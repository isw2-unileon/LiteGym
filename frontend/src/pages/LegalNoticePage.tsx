import LegalPageShell from "../components/LegalPageShell";

const legalContactEmail = import.meta.env.VITE_LEGAL_CONTACT_EMAIL ?? "legal@tu-dominio.com";

export default function LegalNoticePage() {
  return (
    <LegalPageShell
      title="Aviso legal"
      subtitle="Información general sobre la titularidad, el uso del sitio y la responsabilidad asociada al servicio LiteGym."
      updatedAt="Mayo de 2026"
    >
      <h2>1. Titular del sitio</h2>
      <p>
        Este sitio web y la aplicación LiteGym son operados por el equipo responsable del proyecto.
        El canal de contacto legal se publica para gestiones de privacidad, derechos y consultas
        sobre el uso del servicio.
      </p>

      <h2>2. Contacto</h2>
      <p>
        Correo de contacto legal: <a href={`mailto:${legalContactEmail}`}>{legalContactEmail}</a>
      </p>
      <p>
        Para incidencias operativas también puedes usar la sección de{" "}
        <a href="/support">soporte técnico</a> una vez autenticado.
      </p>

      <h2>3. Condiciones de uso</h2>
      <p>El acceso y uso del servicio implica aceptar estas condiciones generales:</p>
      <ul>
        <li>No está permitido usar la plataforma para actividades ilícitas, automatizadas o abusivas.</li>
        <li>El usuario debe mantener la confidencialidad de sus credenciales y de su sesión.</li>
        <li>El servicio puede suspender cuentas por fraude, abuso, spam o uso indebido.</li>
        <li>El contenido mostrado puede ser actualizado, corregido o retirado sin previo aviso cuando sea necesario.</li>
      </ul>

      <h2>4. Responsabilidad sobre contenidos</h2>
      <p>
        LiteGym procura mostrar información útil y mantener la disponibilidad del servicio, pero no
        garantiza la ausencia total de errores, interrupciones o contenidos generados por terceros.
        El usuario debe revisar la información antes de tomar decisiones basadas en ella.
      </p>
    </LegalPageShell>
  );
}
