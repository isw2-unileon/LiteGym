const muscleGroupLabels: Record<string, string> = {
  chest: "Pecho",
  back: "Espalda",
  legs: "Piernas",
  shoulders: "Hombros",
  biceps: "Bíceps",
  triceps: "Tríceps",
  core: "Core",
  glutes: "Glúteos",
  forearms: "Antebrazos",
  calves: "Gemelos",
  hamstrings: "Isquiotibiales",
  quadriceps: "Cuádriceps",
  lats: "Dorsales",
  traps: "Trapecios",
  rear_delts: "Deltoides posteriores",
  front_delts: "Deltoides anteriores",
  side_delts: "Deltoides laterales",
  full_body: "Cuerpo completo",
  cardio: "Cardio",
};

const exerciseTypeLabels: Record<string, string> = {
  strength: "Fuerza",
  cardio: "Cardio",
  mobility: "Movilidad",
  balance: "Equilibrio",
  plyometric: "Pliometría",
  flexibility: "Flexibilidad",
  endurance: "Resistencia",
  warmup: "Calentamiento",
  cooldown: "Vuelta a la calma",
  technique: "Técnica",
  rehab: "Rehabilitación",
};

export function muscleGroupLabel(value: string): string {
  const key = value.trim().toLowerCase();
  return muscleGroupLabels[key] ?? value;
}

export function exerciseTypeLabel(value: string): string {
  const key = value.trim().toLowerCase();
  return exerciseTypeLabels[key] ?? value;
}
