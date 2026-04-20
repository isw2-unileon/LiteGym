type ExerciseFiltersProps = {
    search: string;
    typeFilter: string;
    muscleFilter: string;
    exerciseTypes: string[];
    muscleGroups: string[];
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
        <div className="mt-6 grid gap-4 md:grid-cols-3">
            <input
                type="text"
                placeholder="Buscar por nombre, tipo o músculo"
                value={search}
                onChange={(e) => onSearchChange(e.target.value)}
                className="rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm text-[#1f1b16] outline-none"
            />

            <select
                value={typeFilter}
                onChange={(e) => onTypeFilterChange(e.target.value)}
                className="rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm text-[#1f1b16] outline-none"
            >
                <option value="">Todos los tipos</option>
                {exerciseTypes.map((type) => (
                    <option key={type} value={type}>
                        {type}
                    </option>
                ))}
            </select>

            <select
                value={muscleFilter}
                onChange={(e) => onMuscleFilterChange(e.target.value)}
                className="rounded-2xl border border-[#1f1b16]/10 bg-white px-4 py-3 text-sm text-[#1f1b16] outline-none"
            >
                <option value="">Todos los músculos</option>
                {muscleGroups.map((muscle) => (
                    <option key={muscle} value={muscle}>
                        {muscle}
                    </option>
                ))}
            </select>
        </div>
    );
}
