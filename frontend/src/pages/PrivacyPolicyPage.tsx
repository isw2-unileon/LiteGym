import LegalPageShell from "../components/LegalPageShell";

const legalContactEmail = import.meta.env.VITE_LEGAL_CONTACT_EMAIL ?? "legal@tu-dominio.com";

export default function PrivacyPolicyPage() {
  return (
    <LegalPageShell
      title="Política de privacidad"
      subtitle="Cómo tratamos los datos personales dentro de LiteGym y qué derechos tiene el usuario sobre su información."
      updatedAt="Mayo de 2026"
    >
      <h2>1. Responsable del tratamiento</h2>
      <p>
        El responsable del tratamiento es el titular del proyecto LiteGym, que decide sobre los
        datos recogidos al crear cuentas, usar rutinas, abrir tickets o acceder a la plataforma.
      </p>

      <h2>2. Datos que recogemos</h2>
      <ul>
        <li>Datos de registro y acceso: email, nombre de usuario y credenciales cifradas/gestionadas por el proveedor de autenticación.</li>
        <li>Datos de perfil y actividad: rutinas, ejercicios, progreso y tickets de soporte.</li>
        <li>Datos técnicos y de seguridad: información de sesión, IP aproximada y registros de acceso.</li>
        <li>Datos introducidos para IA: objetivos, notas y contexto de entrenamiento necesarios para generar rutinas.</li>
      </ul>

      <h2>3. Finalidades y base jurídica</h2>
      <ul>
        <li>Prestar el servicio y gestionar la cuenta del usuario.</li>
        <li>Generar y guardar rutinas, progreso y soporte técnico.</li>
        <li>Proteger la plataforma frente a abuso, fraude o uso indebido.</li>
        <li>La base jurídica principal es la ejecución del contrato/servicio solicitado por el usuario, el cumplimiento de obligaciones legales y, cuando aplique, el consentimiento o el interés legítimo.</li>
      </ul>

      <h2>4. Conservación y destinatarios</h2>
      <p>
        Conservamos los datos mientras la cuenta permanezca activa o sea necesario para prestar el
        servicio, atender incidencias o cumplir obligaciones legales. Parte del tratamiento puede
        realizarse mediante proveedores tecnológicos como el alojamiento, la base de datos, el
        sistema de autenticación o el proveedor de IA.
      </p>

      <h2>5. Derechos del usuario</h2>
      <p>
        El usuario puede ejercer sus derechos de acceso, rectificación, supresión, oposición,
        limitación y portabilidad escribiendo al correo{" "}
        <a href={`mailto:${legalContactEmail}`}>{legalContactEmail}</a> o utilizando la sección de
        soporte del producto.
      </p>

      <h2>6. Reclamaciones</h2>
      <p>
        Si el usuario considera que el tratamiento de sus datos no es correcto, también puede
        presentar una reclamación ante la autoridad de control competente.
      </p>
    </LegalPageShell>
  );
}
