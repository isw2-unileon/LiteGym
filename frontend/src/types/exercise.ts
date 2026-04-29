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

export type SelectOption = {
    value: string;
    label: string;
};

export type ExerciseMetadataResponse = {
    exercise_types: SelectOption[];
    muscle_groups: SelectOption[];
};
