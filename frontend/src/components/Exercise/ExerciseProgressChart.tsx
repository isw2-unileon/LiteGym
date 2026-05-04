import type { ExerciseProgressPoint } from "../../types/exercise";

type ExerciseProgressChartProps = {
  points: ExerciseProgressPoint[];
  metric: "max_weight_kg" | "max_reps" | "volume_kg";
};

const metricLabels = {
  max_weight_kg: "Peso maximo",
  max_reps: "Repeticiones",
  volume_kg: "Volumen",
};

const metricUnits = {
  max_weight_kg: "kg",
  max_reps: "reps",
  volume_kg: "kg",
};

const numberFormatter = new Intl.NumberFormat("es-ES", {
  maximumFractionDigits: 1,
});

export default function ExerciseProgressChart({
  points,
  metric,
}: ExerciseProgressChartProps) {
  const chartPoints = points
    .map((point) => ({
      date: point.date,
      value: getMetricValue(point, metric),
    }))
    .filter((point): point is { date: string; value: number } => point.value != null);

  if (chartPoints.length < 2) {
    return (
      <div className="rounded-2xl border border-dashed border-[#1f1b16]/15 bg-[#fffaf0]/80 p-4">
        <p className="text-sm font-semibold text-[#5d5348]">
          Todavia no hay suficientes sesiones para dibujar progresion.
        </p>
      </div>
    );
  }

  const width = 320;
  const height = 150;
  const padding = 18;
  const values = chartPoints.map((point) => point.value);
  const min = Math.min(...values);
  const max = Math.max(...values);
  const range = max - min || 1;

  const polyline = chartPoints
    .map((point, index) => {
      const x =
        padding +
        (index / Math.max(1, chartPoints.length - 1)) * (width - padding * 2);
      const y =
        height -
        padding -
        ((point.value - min) / range) * (height - padding * 2);
      return `${x},${y}`;
    })
    .join(" ");

  const lastPoint = chartPoints[chartPoints.length - 1];
  if (!lastPoint) {
    return null;
  }

  return (
    <div className="rounded-2xl border border-[#1f1b16]/10 bg-white/70 p-4">
      <div className="flex items-start justify-between gap-3">
        <div>
          <p className="text-[11px] font-black uppercase tracking-[0.18em] text-[#265c52]">
            {metricLabels[metric]}
          </p>
          <p className="mt-1 text-sm font-semibold text-[#5d5348]">
            Ultimo dato: {numberFormatter.format(lastPoint.value)}{" "}
            {metricUnits[metric]}
          </p>
        </div>
      </div>

      <svg
        className="mt-4 h-auto w-full"
        viewBox={`0 0 ${width} ${height}`}
        role="img"
        aria-label={`Progresion de ${metricLabels[metric].toLowerCase()}`}
      >
        <line
          x1={padding}
          x2={width - padding}
          y1={height - padding}
          y2={height - padding}
          stroke="rgba(31,27,22,0.16)"
          strokeWidth="2"
        />
        <polyline
          points={polyline}
          fill="none"
          stroke="#265c52"
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth="4"
        />
      </svg>
    </div>
  );
}

function getMetricValue(
  point: ExerciseProgressPoint,
  metric: ExerciseProgressChartProps["metric"],
) {
  const value = point[metric];
  return typeof value === "number" ? value : null;
}
