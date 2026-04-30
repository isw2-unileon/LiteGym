import Profile from "../components/Profile";

export default function ProfilePage() {
  return (
    <section>
      <h1 className="font-['Aptos_Display','Trebuchet_MS',sans-serif] text-4xl font-black tracking-[-0.05em]">
        Perfil
      </h1>
      <div className="mt-8">
        <Profile />
      </div>
    </section>
  );
}
