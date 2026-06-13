import type { ExerciseProgressPoint } from "../../types/exercise";

type ExerciseProgressChartProps = {
  points: ExerciseProgressPoint[];
  metric: "max_weight_kg" | "max_reps" | "volume_kg";
};

const metricLabels = {
  max_weight_kg: "Peso máximo",
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

const dateFormatter = new Intl.DateTimeFormat("es-ES", {
  day: "2-digit",
  month: "short",
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
    .filter((point): point is { date: string; value: number } => point.value != null)
    .sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime());

  if (chartPoints.length < 2) {
    return (
      <div className="rounded-[1.25rem] border border-dashed border-[#1f1b16]/15 bg-white/75 p-4">
        <p className="text-sm font-semibold leading-6 text-[#5d5348]">
          Todavia no hay suficientes sesiones para dibujar progresion.
        </p>
      </div>
    );
  }

  const displayPoints = smoothAndDownsample(chartPoints, metric);
  const width = 560;
  const height = 220;
  const paddingX = 20;
  const paddingY = 24;
  const values = displayPoints.map((point) => point.value);
  const min = Math.min(...values);
  const max = Math.max(...values);
  const range = max - min || 1;
  const linePoints = displayPoints
    .map((point, index) => {
      const x =
        paddingX +
        (index / Math.max(1, displayPoints.length - 1)) * (width - paddingX * 2);
      const y =
        height -
        paddingY -
        ((point.value - min) / range) * (height - paddingY * 2);
      return { ...point, x, y };
    });

  const polyline = linePoints.map((point) => `${point.x},${point.y}`).join(" ");
  const area = [
    `M ${linePoints[0]?.x ?? paddingX} ${height - paddingY}`,
    ...linePoints.map((point) => `L ${point.x} ${point.y}`),
    `L ${linePoints[linePoints.length - 1]?.x ?? width - paddingX} ${height - paddingY}`,
    "Z",
  ].join(" ");
  const lastPoint = linePoints[linePoints.length - 1];
  const firstPoint = linePoints[0];
  const delta = lastPoint && firstPoint ? lastPoint.value - firstPoint.value : 0;
  const deltaPrefix = delta > 0 ? "+" : "";

  return (
    <article className="rounded-[1.25rem] border border-[#1f1b16]/10 bg-white/90 p-4 shadow-[0_10px_24px_rgba(31,27,22,0.05)]">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <p className="text-[11px] font-black uppercase tracking-[0.22em] text-[#265c52]">
            {metricLabels[metric]}
          </p>
          <p className="mt-2 text-sm font-semibold leading-6 text-[#5d5348]">
            Ultimo dato: {numberFormatter.format(lastPoint?.value ?? 0)}{" "}
            {metricUnits[metric]}
          </p>
        </div>

        <div className="rounded-full border border-[#1f1b16]/10 bg-[#fffaf0] px-3 py-1 text-xs font-black uppercase tracking-[0.14em] text-[#3a332c]">
          {deltaPrefix}
          {numberFormatter.format(delta)} {metricUnits[metric]} desde el inicio
        </div>
      </div>

      <div className="mt-4 overflow-hidden rounded-[1rem] border border-[#1f1b16]/10 bg-[#fffaf0]">
        <svg
          className="block h-auto w-full"
          viewBox={`0 0 ${width} ${height}`}
          role="img"
          aria-label={`Progresion de ${metricLabels[metric].toLowerCase()}`}
        >
          {[0.25, 0.5, 0.75].map((fraction) => {
            const y = paddingY + (height - paddingY * 2) * fraction;
            return (
              <line
                key={fraction}
                x1={paddingX}
                x2={width - paddingX}
                y1={y}
                y2={y}
                stroke="rgba(31,27,22,0.08)"
                strokeDasharray="4 4"
                strokeWidth="1.5"
              />
            );
          })}

          <line
            x1={paddingX}
            x2={width - paddingX}
            y1={height - paddingY}
            y2={height - paddingY}
            stroke="rgba(31,27,22,0.18)"
            strokeWidth="2"
          />

          <path d={area} fill="rgba(38,92,82,0.12)" />
          <polyline
            points={polyline}
            fill="none"
            stroke="#265c52"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="3.25"
          />

          {linePoints.map((point, index) => {
            const isStart = index === 0;
            const isEnd = index === linePoints.length - 1;
            if (!isStart && !isEnd) {
              return null;
            }
            return (
              <g key={`${point.date}-${index}`}>
                <circle
                  cx={point.x}
                  cy={point.y}
                  r={isStart || isEnd ? 5.25 : 4}
                  fill={isEnd ? "#ea7130" : "#fffaf0"}
                  stroke={isEnd ? "#ea7130" : "#265c52"}
                  strokeWidth="2.5"
                />
              </g>
            );
          })}
        </svg>
      </div>

      <div className="mt-4 flex flex-wrap gap-2 text-[11px] font-black uppercase tracking-[0.14em] text-[#3a332c]">
        <div className="rounded-full border border-[#1f1b16]/10 bg-[#fffaf0] px-3 py-2">
          <span className="text-[#265c52]">Inicio</span>{" "}
          <span className="normal-case tracking-[-0.02em]">
            {dateFormatter.format(new Date(firstPoint?.date ?? chartPoints[0]!.date))}
          </span>
        </div>
        <div className="rounded-full border border-[#1f1b16]/10 bg-[#fffaf0] px-3 py-2">
          <span className="text-[#265c52]">Max</span>{" "}
          <span className="normal-case tracking-[-0.02em]">
            {numberFormatter.format(max)} {metricUnits[metric]}
          </span>
        </div>
        <div className="rounded-full border border-[#1f1b16]/10 bg-[#fffaf0] px-3 py-2">
          <span className="text-[#265c52]">Ultimo</span>{" "}
          <span className="normal-case tracking-[-0.02em]">
            {dateFormatter.format(new Date(lastPoint?.date ?? chartPoints[chartPoints.length - 1]!.date))}
          </span>
        </div>
      </div>
    </article>
  );
}

function smoothAndDownsample(
  points: { date: string; value: number }[],
  metric: ExerciseProgressChartProps["metric"],
) {
  const maxPoints = metric === "max_weight_kg" ? 12 : 10;
  if (points.length <= maxPoints) {
    return points;
  }

  const bucketSize = Math.ceil(points.length / maxPoints);
  const bucketed: { date: string; value: number }[] = [];

  for (let index = 0; index < points.length; index += bucketSize) {
    const bucket = points.slice(index, index + bucketSize);
    if (bucket.length === 0) {
      continue;
    }

    const value = bucket.reduce((total, point) => total + point.value, 0) / bucket.length;
    bucketed.push({
      date: bucket[bucket.length - 1]!.date,
      value,
    });
  }

  return bucketed;
}

function getMetricValue(
  point: ExerciseProgressPoint,
  metric: ExerciseProgressChartProps["metric"],
) {
  const value = point[metric];
  return typeof value === "number" ? value : null;
}
