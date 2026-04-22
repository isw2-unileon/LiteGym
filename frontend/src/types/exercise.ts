export type Exercise = {
    id: number;
    name: string;
    description: string | null;
    muscle_group: string;
    exercise_type: string | null;
};

export type ExerciseStatus = "idle" | "loading" | "success" | "error";
