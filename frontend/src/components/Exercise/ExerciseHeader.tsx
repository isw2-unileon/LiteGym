export default function ExerciseHeader() {
    return (
        <div className="mb-8 max-w-2xl" data-block="page-header">
            <p
                className="mb-5 inline-flex rounded-full border border-[#1f1b16]/15 bg-white/35 px-4 py-2 text-sm font-semibold uppercase tracking-[0.28em] text-[#265c52] backdrop-blur"
                data-block="brand-badge"
            >
                Grupo 16 Fitness
            </p>

            <h1
                className="font-['Aptos_Display','Trebuchet_MS',sans-serif] text-5xl font-black leading-[0.95] tracking-[-0.06em] text-[#1f1b16] sm:text-7xl"
                data-block="page-title"
            >
                Explora tus ejercicios.
            </h1>

            <p
                className="mt-4 max-w-xl text-base text-[#5d5348] sm:text-lg"
                data-block="page-description"
            >
                Aquí puedes ver los ejercicios disponibles para usarlos al crear
                tus rutinas.
            </p>
        </div>
    );
}
