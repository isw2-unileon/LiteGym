export type Exercise = {
    id: string;
    name: string;
    description: string | null;
    muscle_group: string;
    secondary_muscle_group?: string | null;
    secondary_muscle_groups?: string[] | null;
    exercise_type: string | null;
    is_official?: boolean;
    created_at?: string;
};

export type ExerciseStatus = "idle" | "loading" | "success" | "error";

export type ExerciseWorkoutSessionSummary = {
    id: string;
    name?: string | null;
    routine_name?: string | null;
    started_at: string;
    duration_minutes: number;
    exercise_order: number;
    set_count: number;
};

export type ExerciseInsightTrend = "up" | "stable" | "down" | "empty";

export type ExerciseInsightSummary = {
    session_count: number;
    set_count: number;
    total_volume_kg: number;
    max_weight_kg?: number | null;
    max_reps?: number | null;
    last_performed_at?: string | null;
    first_performed_at?: string | null;
    average_days_between?: number | null;
    trend: ExerciseInsightTrend;
};

export type ExerciseInsightSet = {
    session_id: string;
    session_name?: string | null;
    routine_name?: string | null;
    performed_at: string;
    set_number: number;
    reps?: number | null;
    weight_kg?: number | null;
    volume_kg: number;
    rir?: number | null;
    completed: boolean;
};

export type ExercisePersonalRecord = {
    type: string;
    label: string;
    value: number;
    unit: string;
    performed_at: string;
};

export type ExerciseProgressPoint = {
    session_id: string;
    date: string;
    max_weight_kg?: number | null;
    max_reps?: number | null;
    volume_kg: number;
    set_count: number;
};

export type ExerciseInsightSessionHistory = {
    session_id: string;
    session_name?: string | null;
    routine_name?: string | null;
    performed_at: string;
    duration_minutes: number;
    volume_kg: number;
    sets: ExerciseInsightSet[];
};

export type ExerciseInsights = {
    summary: ExerciseInsightSummary;
    best_set?: ExerciseInsightSet | null;
    personal_records: ExercisePersonalRecord[];
    progression: ExerciseProgressPoint[];
    history: ExerciseInsightSessionHistory[];
};

export type SelectOption = {
    value: string;
    label: string;
};

export type ExerciseMetadataResponse = {
    exercise_types: SelectOption[];
    muscle_groups: SelectOption[];
};
