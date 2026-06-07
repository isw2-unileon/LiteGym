import { useMemo } from "react";
import { Card, CardHeader } from "./Card";
import { SummaryTile } from "./SummaryTile";

type MuscleBreakdown = { name: string; count: number; percentage: number };

export type MuscleDistributionData = {
  year: MuscleBreakdown[];
  month: MuscleBreakdown[];
  year_exercise_count: number;
  month_exercise_count: number;
};

// Ejes fijos del grafico (octogono). Cada eje agrupa los nombres que el
// backend puede devolver para ese grupo muscular.
const AXES: Array<{ label: string; keys: string[] }> = [
  { label: "Pecho", keys: ["chest", "pecho"] },
  { label: "Hombro", keys: ["shoulder", "shoulders", "hombro", "hombros"] },
  { label: "Triceps", keys: ["triceps"] },
  { label: "Biceps", keys: ["biceps"] },
  { label: "Espalda", keys: ["back", "espalda"] },
  { label: "Core", keys: ["core", "abs", "abdomen", "abdominales", "obliques", "oblicuos"] },
  { label: "Gluteos", keys: ["glute", "glutes", "gluteo", "gluteos"] },
  { label: "Pierna", keys: ["leg", "legs", "pierna", "piernas"] },
];

const RINGS = [0.25, 0.5, 0.75, 1];
const CHART_CENTER = 108;
const CHART_RADIUS = 78;

function polarToCartesian(cx: number, cy: number, radius: number, angleInDegrees: number) {
  const angleInRadians = ((angleInDegrees - 90) * Math.PI) / 180;

  return {
    x: cx + radius * Math.cos(angleInRadians),
    y: cy + radius * Math.sin(angleInRadians),
  };
}

export function MuscleDistribution({
  data,
  range,
  onRangeChange,
  summary
}: {
  data?: MuscleDistributionData;
  range: "year" | "month";
  onRangeChange?: (range: "year" | "month") => void;
  summary: boolean;
}) {
  const breakdown = useMemo(
    () => (range === "year" ? (data?.year ?? []) : (data?.month ?? [])),
    [range, data],
  );
  const exerciseCount = range === "year" ? (data?.year_exercise_count ?? 0) : (data?.month_exercise_count ?? 0);
  const muscleGroupCount = breakdown.length;

  const stats = useMemo(() => {
    const axisCounts = AXES.map((axis) => ({
      axis: axis.label,
      count: breakdown
        .filter((group) => axis.keys.includes(group.name.trim().toLowerCase()))
        .reduce((sum, group) => sum + group.count, 0),
    }));
    const maxCount = Math.max(...axisCounts.map((item) => item.count), 1);

    return axisCounts.map((item) => ({
      axis: item.axis,
      value: item.count > 0 ? Math.max(0.22, item.count / maxCount) : 0.22,
    }));
  }, [breakdown]);

  const points = stats.map((item, index) =>
    polarToCartesian(CHART_CENTER, CHART_CENTER, CHART_RADIUS * item.value, (360 / stats.length) * index),
  );
  const polygon = points.map((point) => `${point.x},${point.y}`).join(" ");

  return (
    <Card accent="#ea7130" dark>
      <CardHeader
        kicker="Estadisticas"
        title="Distribucion muscular"
        onDark
        right={onRangeChange && (
          <div
            role="group"
            aria-label="Rango de estadisticas"
            className="inline-flex gap-0.5 rounded-[10px] border border-[#fffaf0]/20 bg-[#fffaf0]/8 p-1"
          >
            {(["year", "month"] as const).map((option) => (
              <button
                key={option}
                type="button"
                onClick={() => onRangeChange(option)}
                aria-pressed={range === option}
                className={[
                  "rounded-[7px] px-3.5 py-2.5 [font-family:'JetBrains_Mono',ui-monospace,monospace] text-[12px] font-bold uppercase leading-none tracking-[0.14em] transition",
                  range === option
                    ? "bg-[#ea7130] text-[#1f1b16]"
                    : "bg-transparent text-[#fffaf0]/70 hover:text-[#fffaf0]",
                ].join(" ")}
              >
                {option === "year" ? "Año" : "Mes"}
              </button>
            ))}
          </div>
        )}
      />

      <div className="relative z-[2] mx-auto mt-4 w-full max-w-[18rem]">
        <svg
          className="h-auto w-full"
          viewBox="-36 -8 288 232"
          role="img"
          aria-label="Grafico octogonal de distribucion muscular"
        >
          {RINGS.map((level) => {
            const ringPoints = stats
              .map((_, index) => polarToCartesian(CHART_CENTER, CHART_CENTER, CHART_RADIUS * level, (360 / stats.length) * index))
              .map((point) => `${point.x},${point.y}`)
              .join(" ");

            return <polygon fill="none" key={level} points={ringPoints} stroke="rgba(255,250,240,0.16)" strokeWidth="1.4" />;
          })}

          {stats.map((item, index) => {
            const outerPoint = polarToCartesian(CHART_CENTER, CHART_CENTER, CHART_RADIUS, (360 / stats.length) * index);
            const labelPoint = polarToCartesian(CHART_CENTER, CHART_CENTER, CHART_RADIUS + 16, (360 / stats.length) * index);
            const dx = labelPoint.x - CHART_CENTER;
            const textAnchor = Math.abs(dx) < 1 ? "middle" : dx > 0 ? "start" : "end";

            return (
              <g key={item.axis}>
                <line stroke="rgba(255,250,240,0.14)" strokeWidth="1.4" x1={CHART_CENTER} x2={outerPoint.x} y1={CHART_CENTER} y2={outerPoint.y} />
                <text
                  fill="rgba(255,250,240,0.85)"
                  fontFamily="'JetBrains Mono', ui-monospace, monospace"
                  fontSize="12"
                  fontWeight={700}
                  textAnchor={textAnchor}
                  dominantBaseline="middle"
                  x={labelPoint.x}
                  y={labelPoint.y}
                  style={{ letterSpacing: "0.06em", textTransform: "uppercase" }}
                >
                  {item.axis}
                </text>
              </g>
            );
          })}

          <polygon fill="rgba(234,113,48,0.28)" points={polygon} stroke="#ea7130" strokeWidth="2.4" strokeLinejoin="round" />
          {points.map((point, index) => (
            <circle key={index} cx={point.x} cy={point.y} r="3.5" fill="#ea7130" stroke="#fffaf0" strokeWidth="1.4" />
          ))}
        </svg>
      </div>

      {summary && (
      <div className="relative z-[2] mt-3 grid grid-cols-2 gap-2">
        <SummaryTile dark value={muscleGroupCount} label={`${muscleGroupCount === 1 ? "Grupo muscular presente":"Grupos musculares presentes"}`} />
        <SummaryTile
          dark
          value={exerciseCount}
          label={`${range === "year" ? "Último año" : "Último mes"} · ${exerciseCount} ${exerciseCount === 1 ? "ejercicio considerado": "ejercicios considerados"}`}
        />
      </div>)}
    </Card>
  );
}
