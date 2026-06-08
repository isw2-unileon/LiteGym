import LegalPageShell from "../components/LegalPageShell";

export default function TermsPage() {
  return (
    <LegalPageShell
      title="Términos y condiciones"
      subtitle="Normas de uso del servicio LiteGym, incluyendo el contenido generado por IA y las responsabilidades del usuario."
      updatedAt="Mayo de 2026"
    >
      <h2>1. Uso aceptable</h2>
      <p>El usuario se compromete a usar la plataforma de forma lícita, diligente y respetuosa.</p>
      <ul>
        <li>No se permite el abuso de la plataforma, la automatización no autorizada ni el intento de eludir límites de seguridad.</li>
        <li>No se permite introducir contenido dañino, engañoso o contrario a la ley.</li>
        <li>El usuario es responsable de la información que carga, comparte o genera dentro del servicio.</li>
      </ul>

      <h2>2. Edad mínima y responsabilidad personal</h2>
      <p>
        El servicio está pensado para personas con capacidad legal para aceptar estas condiciones.
        El usuario debe revisar que el uso de rutinas y recomendaciones sea adecuado a su estado
        físico y, si es necesario, consultar a un profesional sanitario.
      </p>

      <h2>3. Uso de IA</h2>
      <p>
        LiteGym puede generar rutinas y sugerencias mediante un sistema de IA. Ese contenido es
        orientativo y puede contener errores o limitaciones. No sustituye el criterio de un
        entrenador, fisioterapeuta o médico.
      </p>
      <ul>
        <li>Las rutinas generadas deben revisarse antes de ejecutarse.</li>
        <li>El usuario no debe introducir datos sensibles innecesarios en las notas o prompts.</li>
        <li>El servicio puede limitar el uso de IA para proteger la disponibilidad y evitar abuso.</li>
      </ul>

      <h2>4. Cuentas, suspensión y borrado</h2>
      <p>
        LiteGym puede suspender o cerrar cuentas por incumplimiento de estas condiciones, por uso
        fraudulento o por protección de la plataforma y sus usuarios. El usuario también puede
        solicitar el cierre/borrado de su cuenta mediante los canales de soporte y privacidad.
      </p>

      <h2>5. Limitación de responsabilidad</h2>
      <p>
        El servicio se ofrece “tal cual” en la medida permitida por la ley. No se garantiza
        disponibilidad ininterrumpida ni exactitud absoluta del contenido generado o mostrado.
      </p>
    </LegalPageShell>
  );
}
