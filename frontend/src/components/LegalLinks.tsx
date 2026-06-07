import { Link } from "react-router-dom";

export const legalLinks = [
  { to: "/legal/aviso-legal", label: "Aviso legal" },
  { to: "/legal/privacidad", label: "Privacidad" },
  { to: "/legal/cookies", label: "Cookies" },
  { to: "/legal/terminos", label: "Términos" },
];

type LegalLinksProps = {
  className?: string;
  itemClassName?: string;
  compact?: boolean;
};

export default function LegalLinks({ className = "", itemClassName = "", compact = false }: LegalLinksProps) {
  return (
    <nav aria-label="Enlaces legales" className={className}>
      <div className={`flex flex-wrap ${compact ? "gap-x-3 gap-y-2" : "gap-3"}`}>
        {legalLinks.map((link) => (
          <Link
            key={link.to}
            to={link.to}
            className={
              itemClassName ||
              "text-sm font-bold text-[#1f1b16]/75 transition hover:text-[#ea7130]"
            }
          >
            {link.label}
          </Link>
        ))}
      </div>
    </nav>
  );
}
