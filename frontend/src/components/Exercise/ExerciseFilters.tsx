import type { SelectOption } from "../../types/exercise";

type ExerciseFiltersProps = {
    search: string;
    typeFilter: string;
    muscleFilter: string;
    exerciseTypes: SelectOption[];
    muscleGroups: SelectOption[];
    onSearchChange: (value: string) => void;
    onTypeFilterChange: (value: string) => void;
    onMuscleFilterChange: (value: string) => void;
};

export default function ExerciseFilters({
    search,
    typeFilter,
    muscleFilter,
    exerciseTypes,
    muscleGroups,
    onSearchChange,
    onTypeFilterChange,
    onMuscleFilterChange,
}: ExerciseFiltersProps) {
    return (
        <div className="grid gap-4">
            <input
                type="text"
                placeholder="Buscar por nombre..."
                value={search}
                onChange={(e) => onSearchChange(e.target.value)}
                className="w-full rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm text-[#1f1b16] outline-none"
            />

            <div className="grid gap-4 sm:grid-cols-2">
                <select
                    value={typeFilter}
                    onChange={(e) => onTypeFilterChange(e.target.value)}
                    className="rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm text-[#1f1b16] outline-none"
                >
                    <option value="">Todos los tipos</option>
                    {exerciseTypes.length === 0 && (
                        <option value="" disabled>
                            Cargando tipos...
                        </option>
                    )}
                    {exerciseTypes.map((type) => (
                        <option key={type.value} value={type.value}>
                            {type.label}
                        </option>
                    ))}
                </select>

                <select
                    value={muscleFilter}
                    onChange={(e) => onMuscleFilterChange(e.target.value)}
                    className="rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm text-[#1f1b16] outline-none"
                >
                    <option value="">Todos los músculos</option>
                    {muscleGroups.length === 0 && (
                        <option value="" disabled>
                            Cargando músculos...
                        </option>
                    )}
                    {muscleGroups.map((muscle) => (
                        <option key={muscle.value} value={muscle.value}>
                            {muscle.label}
                        </option>
                    ))}
                </select>
            </div>
        </div>
    );
}
