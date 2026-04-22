export type Exercise = {
    id: number;
    name: string;
    description: string | null;
    muscle_group: string;
    exercise_type: string | null;
    created_at?: string;
};

export type ExerciseStatus = "idle" | "loading" | "success" | "error";
