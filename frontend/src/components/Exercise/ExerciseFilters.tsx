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
        <div className="grid gap-3 sm:gap-4">
            <input
                type="search"
                placeholder="Buscar por nombre..."
                value={search}
                onChange={(e) => onSearchChange(e.target.value)}
                className="w-full rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3.5 text-base font-semibold text-[#1f1b16] outline-none placeholder:text-[#3a332c]/45 sm:mt-4 sm:py-3 sm:text-sm"
            />

            <div className="grid grid-cols-2 gap-2.5 sm:gap-4">
                <select
                    value={typeFilter}
                    onChange={(e) => onTypeFilterChange(e.target.value)}
                    className="min-w-0 rounded-2xl border border-[#1f1b16]/10 bg-white px-3 py-3.5 text-sm font-bold text-[#1f1b16] outline-none sm:px-4 sm:py-3"
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
                    className="min-w-0 rounded-2xl border border-[#1f1b16]/10 bg-white px-3 py-3.5 text-sm font-bold text-[#1f1b16] outline-none sm:px-4 sm:py-3"
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
