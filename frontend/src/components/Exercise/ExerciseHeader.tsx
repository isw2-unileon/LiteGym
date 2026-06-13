export default function ExerciseHeader() {
    return (
        <div className="mt-4 max-w-lg" data-block="page-header">
            <h1
                className="font-['Aptos_Display','Trebuchet_MS',sans-serif] text-4xl font-black leading-[1.02] tracking-[-0.04em] text-[#1f1b16] sm:text-5xl"
                data-block="page-title"
            >
                Explora tus ejercicios.
            </h1>

            <p
                className="mt-3 max-w-md text-sm font-semibold leading-6 text-[#5d5348] sm:text-base"
                data-block="page-description"
            >
                Aquí puedes ver los ejercicios disponibles para usarlos al crear
                tus rutinas.
            </p>
        </div>
    );
}
