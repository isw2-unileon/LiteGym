import { useEffect, useState } from "react";

const mobileQuery = "(max-width: 767px)";

export function useIsMobile() {
  const [isMobile, setIsMobile] = useState(() =>
    typeof window !== "undefined" && "matchMedia" in window
      ? window.matchMedia(mobileQuery).matches
      : false,
  );

  useEffect(() => {
    if (!("matchMedia" in window)) {
      return;
    }

    const mediaQuery = window.matchMedia(mobileQuery);
    const handleChange = () => setIsMobile(mediaQuery.matches);

    handleChange();
    mediaQuery.addEventListener("change", handleChange);
    return () => mediaQuery.removeEventListener("change", handleChange);
  }, []);

  return isMobile;
}
