-- Seed data adjusted for UUID primary keys and references.
-- Assumes the schema creates UUID columns with DEFAULT gen_random_uuid()
-- for primary keys and that foreign keys point to UUID columns.

INSERT INTO public.users (id, username, email, password_hash, role, is_active)
VALUES
  ('11111111-1111-1111-1111-111111111111', 'diego', 'diego@example.com', crypt('1234', gen_salt('bf')), 'admin', true),
  ('22222222-2222-2222-2222-222222222222', 'laura', 'laura@example.com', crypt('1234', gen_salt('bf')), 'user', true),
  ('33333333-3333-3333-3333-333333333333', 'sergio', 'sergio@example.com', crypt('1234', gen_salt('bf')), 'user', true),
  ('44444444-4444-4444-4444-444444444444', 'marta', 'marta@example.com', crypt('1234', gen_salt('bf')), 'user', true);

INSERT INTO public.user_profiles (
  user_id, first_name, last_name, age, weight_kg, height_cm, goal, experience_level
)
VALUES
  ('11111111-1111-1111-1111-111111111111', 'Diego', 'Martinez', 27, 78.5, 178, 'Ganar masa muscular', 'intermediate'),
  ('22222222-2222-2222-2222-222222222222', 'Laura', 'Gomez', 25, 60.0, 165, 'Tonificar', 'beginner'),
  ('33333333-3333-3333-3333-333333333333', 'Sergio', 'Lopez', 31, 84.2, 181, 'Perder grasa', 'intermediate'),
  ('44444444-4444-4444-4444-444444444444', 'Marta', 'Ruiz', 29, 56.8, 162, 'Mejorar resistencia', 'advanced');

INSERT INTO public.body_metrics (
  id,
  user_id,
  recorded_at,
  weight_kg,
  body_fat_percentage,
  muscle_mass_kg,
  chest_cm,
  waist_cm,
  arm_cm,
  leg_cm,
  notes
)
VALUES
  ('55555555-5555-5555-5555-555555555551', '11111111-1111-1111-1111-111111111111', now() - interval '10 days', 79.2, 18.5, 36.0, 102, 84, 36, 58, 'Inicio de volumen'),
  ('55555555-5555-5555-5555-555555555552', '11111111-1111-1111-1111-111111111111', now() - interval '3 days', 78.5, 17.9, 36.4, 103, 83, 36.5, 58.5, 'Buenas sensaciones'),
  ('55555555-5555-5555-5555-555555555553', '22222222-2222-2222-2222-222222222222', now() - interval '7 days', 60.0, 24.0, 22.0, 88, 72, 27, 50, 'Primer registro'),
  ('55555555-5555-5555-5555-555555555554', '33333333-3333-3333-3333-333333333333', now() - interval '5 days', 84.2, 22.5, 34.0, 100, 89, 34, 57, 'Objetivo definición');

INSERT INTO public.exercises (
  id,
  name,
  description,
  muscle_group,
  exercise_type,
  is_official,
  owner_user_id
)
VALUES
  ('66666666-6666-6666-6666-666666666661', 'Press banca', 'Ejercicio básico de pecho con barra', 'chest', 'strength', true, NULL),
  ('66666666-6666-6666-6666-666666666662', 'Sentadilla', 'Ejercicio básico de pierna con barra', 'legs', 'strength', true, NULL),
  ('66666666-6666-6666-6666-666666666663', 'Peso muerto', 'Ejercicio compuesto de cadena posterior', 'back', 'strength', true, NULL),
  ('66666666-6666-6666-6666-666666666664', 'Dominadas', 'Ejercicio de tracción vertical', 'back', 'bodyweight', true, NULL),
  ('66666666-6666-6666-6666-666666666665', 'Press militar', 'Empuje vertical para hombros', 'shoulders', 'strength', true, NULL),
  ('66666666-6666-6666-6666-666666666666', 'Curl bíceps mancuerna', 'Aislamiento de bíceps', 'biceps', 'hypertrophy', true, NULL),
  ('66666666-6666-6666-6666-666666666667', 'Plancha abdominal', 'Trabajo isométrico de core', 'core', 'isometric', true, NULL),
  ('66666666-6666-6666-6666-666666666668', 'Hip thrust', 'Trabajo de glúteo', 'glutes', 'strength', true, NULL),
  ('66666666-6666-6666-6666-666666666669', 'Remo con barra', 'Tracción horizontal', 'back', 'strength', true, NULL),
  ('66666666-6666-6666-6666-666666666670', 'Flexiones diamante', 'Variante enfocada en tríceps', 'triceps', 'bodyweight', false, '11111111-1111-1111-1111-111111111111');

INSERT INTO public.exercise_secondary_muscle_groups (exercise_id, muscle_group)
VALUES
  ('66666666-6666-6666-6666-666666666661', 'triceps'),
  ('66666666-6666-6666-6666-666666666661', 'shoulders'),
  ('66666666-6666-6666-6666-666666666662', 'glutes'),
  ('66666666-6666-6666-6666-666666666663', 'legs'),
  ('66666666-6666-6666-6666-666666666664', 'biceps'),
  ('66666666-6666-6666-6666-666666666665', 'triceps'),
  ('66666666-6666-6666-6666-666666666668', 'legs'),
  ('66666666-6666-6666-6666-666666666669', 'biceps'),
  ('66666666-6666-6666-6666-666666666670', 'chest');

INSERT INTO public.friendships (id, requester_id, addressee_id, status)
VALUES
  ('77777777-7777-7777-7777-777777777771', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222', 'accepted'),
  ('77777777-7777-7777-7777-777777777772', '11111111-1111-1111-1111-111111111111', '33333333-3333-3333-3333-333333333333', 'pending'),
  ('77777777-7777-7777-7777-777777777773', '22222222-2222-2222-2222-222222222222', '44444444-4444-4444-4444-444444444444', 'accepted');

INSERT INTO public.routines (id, user_id, name, description, is_predefined, is_public)
VALUES
  ('88888888-8888-8888-8888-888888888881', '11111111-1111-1111-1111-111111111111', 'Push Pull Legs', 'Rutina dividida en 3 días', false, true),
  ('88888888-8888-8888-8888-888888888882', '22222222-2222-2222-2222-222222222222', 'Full Body Inicio', 'Rutina sencilla de cuerpo completo', false, true),
  ('88888888-8888-8888-8888-888888888883', '11111111-1111-1111-1111-111111111111', 'Torso Pierna', 'Rutina de 4 días', true, true),
  ('88888888-8888-8888-8888-888888888884', '33333333-3333-3333-3333-333333333333', 'Definición Express', 'Rutina con superseries y cardio', false, false);

INSERT INTO public.routine_exercises (id, routine_id, exercise_id, exercise_order, notes)
VALUES
  ('99999999-9999-9999-9999-999999999991', '88888888-8888-8888-8888-888888888881', '66666666-6666-6666-6666-666666666661', 1, '4 series de 6-8 reps'),
  ('99999999-9999-9999-9999-999999999992', '88888888-8888-8888-8888-888888888881', '66666666-6666-6666-6666-666666666665', 2, '3 series de 8-10 reps'),
  ('99999999-9999-9999-9999-999999999993', '88888888-8888-8888-8888-888888888881', '66666666-6666-6666-6666-666666666670', 3, '3 series al fallo técnico'),
  ('99999999-9999-9999-9999-999999999994', '88888888-8888-8888-8888-888888888882', '66666666-6666-6666-6666-666666666662', 1, '3 series de 10 reps'),
  ('99999999-9999-9999-9999-999999999995', '88888888-8888-8888-8888-888888888882', '66666666-6666-6666-6666-666666666664', 2, '3 series de 6-8 reps asistidas si hace falta'),
  ('99999999-9999-9999-9999-999999999996', '88888888-8888-8888-8888-888888888882', '66666666-6666-6666-6666-666666666667', 3, '3 bloques de 30 segundos'),
  ('99999999-9999-9999-9999-999999999997', '88888888-8888-8888-8888-888888888883', '66666666-6666-6666-6666-666666666661', 1, 'Día torso'),
  ('99999999-9999-9999-9999-999999999998', '88888888-8888-8888-8888-888888888883', '66666666-6666-6666-6666-666666666669', 2, 'Controlar técnica'),
  ('99999999-9999-9999-9999-999999999999', '88888888-8888-8888-8888-888888888883', '66666666-6666-6666-6666-666666666662', 3, 'Día pierna'),
  ('99999999-9999-9999-9999-99999999999a', '88888888-8888-8888-8888-888888888884', '66666666-6666-6666-6666-666666666668', 1, 'Pausa arriba'),
  ('99999999-9999-9999-9999-99999999999b', '88888888-8888-8888-8888-888888888884', '66666666-6666-6666-6666-666666666667', 2, 'Core al final');

INSERT INTO public.shared_routines (id, routine_id, owner_user_id, shared_with_user_id)
VALUES
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1', '88888888-8888-8888-8888-888888888881', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222'),
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2', '88888888-8888-8888-8888-888888888881', '11111111-1111-1111-1111-111111111111', '33333333-3333-3333-3333-333333333333'),
  ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3', '88888888-8888-8888-8888-888888888882', '22222222-2222-2222-2222-222222222222', '44444444-4444-4444-4444-444444444444');

INSERT INTO public.support_tickets (id, user_id, title, description, status)
VALUES
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb1', '22222222-2222-2222-2222-222222222222', 'No puedo cambiar mi avatar', 'Cuando intento subir una imagen da error.', 'open'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb2', '33333333-3333-3333-3333-333333333333', 'Rutina no aparece compartida', 'Mi amigo dice que me la ha compartido pero no la veo.', 'in_progress'),
  ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb3', '44444444-4444-4444-4444-444444444444', 'Error al guardar medidas', 'Se queda cargando al guardar body metrics.', 'closed');

INSERT INTO public.workout_sessions (id, user_id, routine_id, name, performed_at, planned_at, duration_minutes, calories_burned, notes)
VALUES
  ('cccccccc-cccc-cccc-cccc-ccccccccccc1', '11111111-1111-1111-1111-111111111111', '88888888-8888-8888-8888-888888888881', 'Push Day 1', now() - interval '2 days', NULL, 70, 540, 'Buen rendimiento'),
  ('cccccccc-cccc-cccc-cccc-ccccccccccc2', '22222222-2222-2222-2222-222222222222', '88888888-8888-8888-8888-888888888882', 'Full Body lunes', now() - interval '1 day', NULL, 55, 320, 'Primer entreno completado'),
  ('cccccccc-cccc-cccc-cccc-ccccccccccc3', '33333333-3333-3333-3333-333333333333', '88888888-8888-8888-8888-888888888884', 'Definición circuito', now() - interval '3 days', NULL, 60, 470, 'Mucho cardio final');

INSERT INTO public.workout_exercises (id, workout_session_id, exercise_id, exercise_order, notes)
VALUES
  ('dddddddd-dddd-dddd-dddd-ddddddddddd1', 'cccccccc-cccc-cccc-cccc-ccccccccccc1', '66666666-6666-6666-6666-666666666661', 1, 'Muy buenas sensaciones'),
  ('dddddddd-dddd-dddd-dddd-ddddddddddd2', 'cccccccc-cccc-cccc-cccc-ccccccccccc1', '66666666-6666-6666-6666-666666666665', 2, 'Última serie muy dura'),
  ('dddddddd-dddd-dddd-dddd-ddddddddddd3', 'cccccccc-cccc-cccc-cccc-ccccccccccc1', '66666666-6666-6666-6666-666666666670', 3, 'Cerca del fallo'),
  ('dddddddd-dddd-dddd-dddd-ddddddddddd4', 'cccccccc-cccc-cccc-cccc-ccccccccccc2', '66666666-6666-6666-6666-666666666662', 1, 'Técnica correcta'),
  ('dddddddd-dddd-dddd-dddd-ddddddddddd5', 'cccccccc-cccc-cccc-cccc-ccccccccccc2', '66666666-6666-6666-6666-666666666664', 2, 'Con goma de asistencia'),
  ('dddddddd-dddd-dddd-dddd-ddddddddddd6', 'cccccccc-cccc-cccc-cccc-ccccccccccc2', '66666666-6666-6666-6666-666666666667', 3, 'Core estable'),
  ('dddddddd-dddd-dddd-dddd-ddddddddddd7', 'cccccccc-cccc-cccc-cccc-ccccccccccc3', '66666666-6666-6666-6666-666666666668', 1, 'Buena activación'),
  ('dddddddd-dddd-dddd-dddd-ddddddddddd8', 'cccccccc-cccc-cccc-cccc-ccccccccccc3', '66666666-6666-6666-6666-666666666667', 2, 'Core fatigado');

INSERT INTO public.workout_sets (id, workout_exercise_id, set_number, reps, weight_kg, duration_seconds, distance_km, rir, completed)
VALUES
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeee1', 'dddddddd-dddd-dddd-dddd-ddddddddddd1', 1, 8, 80, NULL, NULL, 2, true),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeee2', 'dddddddd-dddd-dddd-dddd-ddddddddddd1', 2, 8, 80, NULL, NULL, 1, true),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeee3', 'dddddddd-dddd-dddd-dddd-ddddddddddd1', 3, 7, 82.5, NULL, NULL, 1, true),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeee4', 'dddddddd-dddd-dddd-dddd-ddddddddddd2', 1, 10, 40, NULL, NULL, 2, true),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeee5', 'dddddddd-dddd-dddd-dddd-ddddddddddd2', 2, 9, 42.5, NULL, NULL, 1, true),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeee6', 'dddddddd-dddd-dddd-dddd-ddddddddddd3', 1, 15, NULL, NULL, NULL, 1, true),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeee7', 'dddddddd-dddd-dddd-dddd-ddddddddddd4', 1, 10, 50, NULL, NULL, 2, true),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeee8', 'dddddddd-dddd-dddd-dddd-ddddddddddd5', 1, 6, NULL, NULL, NULL, 3, true),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeee9', 'dddddddd-dddd-dddd-dddd-ddddddddddd6', 1, NULL, NULL, 45, NULL, 2, true),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeea', 'dddddddd-dddd-dddd-dddd-ddddddddddd7', 1, 12, 90, NULL, NULL, 2, true),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeeb', 'dddddddd-dddd-dddd-dddd-ddddddddddd8', 1, NULL, NULL, 60, NULL, 1, true);

-- Additional history for dashboard views across month/year ranges.
INSERT INTO public.body_metrics (
  id,
  user_id,
  recorded_at,
  weight_kg,
  body_fat_percentage,
  muscle_mass_kg,
  chest_cm,
  waist_cm,
  arm_cm,
  leg_cm,
  notes
)
VALUES
  ('55555555-5555-5555-5555-555555555555', '11111111-1111-1111-1111-111111111111', now() - interval '45 days', 79.8, 19.3, 35.6, 101.5, 85, 35.8, 57.8, 'Cierre del mes anterior'),
  ('55555555-5555-5555-5555-555555555556', '11111111-1111-1111-1111-111111111111', now() - interval '90 days', 80.4, 20.1, 35.1, 100.8, 86, 35.2, 57.2, 'Vuelta tras vacaciones'),
  ('55555555-5555-5555-5555-555555555557', '11111111-1111-1111-1111-111111111111', now() - interval '180 days', 77.9, 18.9, 34.8, 99.9, 84.5, 35.0, 56.9, 'Inicio del bloque anterior');

INSERT INTO public.workout_sessions (id, user_id, routine_id, name, performed_at, planned_at, duration_minutes, calories_burned, notes)
VALUES
  ('cccccccc-cccc-cccc-cccc-ccccccccccc4', '11111111-1111-1111-1111-111111111111', '88888888-8888-8888-8888-888888888881', 'Pull Day 1', now() - interval '6 days', NULL, 68, 510, 'Buena progresión en espalda'),
  ('cccccccc-cccc-cccc-cccc-ccccccccccc5', '11111111-1111-1111-1111-111111111111', '88888888-8888-8888-8888-888888888883', 'Torso pesado', now() - interval '10 days', NULL, 75, 560, 'Sesión intensa'),
  ('cccccccc-cccc-cccc-cccc-ccccccccccc6', '11111111-1111-1111-1111-111111111111', '88888888-8888-8888-8888-888888888883', 'Pierna técnica', now() - interval '15 days', NULL, 72, 590, 'Mejor control de tempo'),
  ('cccccccc-cccc-cccc-cccc-ccccccccccc7', '11111111-1111-1111-1111-111111111111', '88888888-8888-8888-8888-888888888881', 'Push Day 2', now() - interval '24 days', NULL, 66, 500, 'Volumen moderado'),
  ('cccccccc-cccc-cccc-cccc-ccccccccccc8', '11111111-1111-1111-1111-111111111111', '88888888-8888-8888-8888-888888888883', 'Torso de control', now() - interval '40 days', NULL, 64, 485, 'Semana de descarga'),
  ('cccccccc-cccc-cccc-cccc-ccccccccccc9', '11111111-1111-1111-1111-111111111111', '88888888-8888-8888-8888-888888888881', 'Pull Day 2', now() - interval '95 days', NULL, 69, 515, 'Retomando volumen'),
  ('cccccccc-cccc-cccc-cccc-ccccccccccca', '11111111-1111-1111-1111-111111111111', '88888888-8888-8888-8888-888888888883', 'Pierna base', now() - interval '170 days', NULL, 71, 605, 'Buen trabajo de base');

INSERT INTO public.workout_exercises (id, workout_session_id, exercise_id, exercise_order, notes)
VALUES
  ('dddddddd-dddd-dddd-dddd-ddddddddddd9',  'cccccccc-cccc-cccc-cccc-ccccccccccc4', '66666666-6666-6666-6666-666666666669', 1, 'Prioridad técnica'),
  ('dddddddd-dddd-dddd-dddd-ddddddddddda', 'cccccccc-cccc-cccc-cccc-ccccccccccc4', '66666666-6666-6666-6666-666666666664', 2, 'Rango completo'),
  ('dddddddd-dddd-dddd-dddd-dddddddddddb', 'cccccccc-cccc-cccc-cccc-ccccccccccc5', '66666666-6666-6666-6666-666666666661', 1, 'Top set pesado'),
  ('dddddddd-dddd-dddd-dddd-dddddddddddc', 'cccccccc-cccc-cccc-cccc-ccccccccccc6', '66666666-6666-6666-6666-666666666662', 1, 'Profundidad controlada'),
  ('dddddddd-dddd-dddd-dddd-dddddddddddd', 'cccccccc-cccc-cccc-cccc-ccccccccccc7', '66666666-6666-6666-6666-666666666665', 1, 'Sube carga'),
  ('dddddddd-dddd-dddd-dddd-ddddddddddde', 'cccccccc-cccc-cccc-cccc-ccccccccccc8', '66666666-6666-6666-6666-666666666669', 1, 'Semana ligera'),
  ('dddddddd-dddd-dddd-dddd-dddddddddddf', 'cccccccc-cccc-cccc-cccc-ccccccccccc9', '66666666-6666-6666-6666-666666666663', 1, 'Recuperando fuerza'),
  ('dddddddd-dddd-dddd-dddd-ddddddddddca', 'cccccccc-cccc-cccc-cccc-ccccccccccca', '66666666-6666-6666-6666-666666666668', 1, 'Trabajo de glúteo');

INSERT INTO public.workout_sets (id, workout_exercise_id, set_number, reps, weight_kg, duration_seconds, distance_km, rir, completed)
VALUES
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeec', 'dddddddd-dddd-dddd-dddd-ddddddddddd9', 1, 10, 65, NULL, NULL, 2, true),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeed', 'dddddddd-dddd-dddd-dddd-ddddddddddda', 1, 7, NULL, NULL, NULL, 2, true),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'dddddddd-dddd-dddd-dddd-dddddddddddb', 1, 6, 85, NULL, NULL, 1, true),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeef', 'dddddddd-dddd-dddd-dddd-dddddddddddc', 1, 8, 95, NULL, NULL, 2, true),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeea', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 1, 8, 45, NULL, NULL, 1, true),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeef01', 'dddddddd-dddd-dddd-dddd-ddddddddddde', 1, 12, 55, NULL, NULL, 3, true),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeef02', 'dddddddd-dddd-dddd-dddd-dddddddddddf', 1, 5, 120, NULL, NULL, 2, true),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeef03', 'dddddddd-dddd-dddd-dddd-ddddddddddca', 1, 10, 100, NULL, NULL, 2, true);

INSERT INTO public.user_goals (user_id, short_term, long_term, target_days_per_week, updated_at)
VALUES
  ('11111111-1111-1111-1111-111111111111', 'Levantar 100kg en Press Banca', 'Bajar al 15% de grasa corporal', 4, CURRENT_TIMESTAMP),
  ('22222222-2222-2222-2222-222222222222', 'Hacer mi primera dominada', 'Crear el hábito de entrenar', 3, CURRENT_TIMESTAMP),
  ('33333333-3333-3333-3333-333333333333', 'Perder 3kg este mes', 'Correr 10km sin parar', 4, CURRENT_TIMESTAMP);

INSERT INTO public.body_metrics (
  id, user_id, recorded_at, weight_kg, body_fat_percentage, chest_cm, waist_cm, arm_cm, leg_cm, notes
)
VALUES
  (gen_random_uuid(), '11111111-1111-1111-1111-111111111111', now() - interval '28 days', 81.0, 19.5, 101, 86, 35, 57, 'Inicio de la racha actual'),
  (gen_random_uuid(), '11111111-1111-1111-1111-111111111111', now() - interval '21 days', 80.2, 19.1, 101.5, 85, 35.5, 57.5, 'Bajando peso, subiendo fuerza'),
  (gen_random_uuid(), '11111111-1111-1111-1111-111111111111', now() - interval '14 days', 79.5, 18.8, 102, 84.5, 36, 58, 'Semana dura');

INSERT INTO public.workout_sessions (
  id, user_id, routine_id, name, started_at, ended_at, duration_minutes, calories_burned, notes
)
VALUES
  ('cccccccc-cccc-cccc-cccc-ccccccccccab', '44444444-4444-4444-4444-444444444444',
   '88888888-8888-8888-8888-888888888882', 'Full Body intenso',
   now() - interval '2 days', now() - interval '2 days' + interval '60 minutes', 60, 420, 'Buen ritmo'),

  ('cccccccc-cccc-cccc-cccc-ccccccccccac', '44444444-4444-4444-4444-444444444444',
   '88888888-8888-8888-8888-888888888882', 'Cardio + core',
   now() - interval '5 days', now() - interval '5 days' + interval '50 minutes', 50, 380, 'Mucho cardio'),

  ('cccccccc-cccc-cccc-cccc-ccccccccccad', '44444444-4444-4444-4444-444444444444',
   '88888888-8888-8888-8888-888888888882', 'Pierna + glúteo',
   now() - interval '9 days', now() - interval '9 days' + interval '70 minutes', 70, 500, 'Intenso en piernas');

INSERT INTO public.workout_exercises (
  id, workout_session_id, exercise_id, exercise_order, notes
)
VALUES
  ('dddddddd-dddd-dddd-dddd-dddddddddeaa', 'cccccccc-cccc-cccc-cccc-ccccccccccab', '66666666-6666-6666-6666-666666666661', 1, 'Press banca controlado'),
  ('dddddddd-dddd-dddd-dddd-dddddddddeab', 'cccccccc-cccc-cccc-cccc-ccccccccccab', '66666666-6666-6666-6666-666666666667', 2, 'Core estable'),

  ('dddddddd-dddd-dddd-dddd-dddddddddeac', 'cccccccc-cccc-cccc-cccc-ccccccccccac', '66666666-6666-6666-6666-666666666667', 1, 'Abdomen'),
  ('dddddddd-dddd-dddd-dddd-dddddddddead', 'cccccccc-cccc-cccc-cccc-ccccccccccac', '66666666-6666-6666-6666-666666666664', 2, 'Dominadas asistidas'),

  ('dddddddd-dddd-dddd-dddd-dddddddddeae', 'cccccccc-cccc-cccc-cccc-ccccccccccad', '66666666-6666-6666-6666-666666666668', 1, 'Hip thrust fuerte'),
  ('dddddddd-dddd-dddd-dddd-dddddddddeaf', 'cccccccc-cccc-cccc-cccc-ccccccccccad', '66666666-6666-6666-6666-666666666662', 2, 'Sentadilla técnica');

INSERT INTO public.workout_sets (
  id, workout_exercise_id, set_number, reps, weight_kg, duration_seconds, distance_km, rir, completed
)
VALUES
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeee12', 'dddddddd-dddd-dddd-dddd-dddddddddeaa', 1, 10, 55, NULL, NULL, 2, true),
  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeee13', 'dddddddd-dddd-dddd-dddd-dddddddddeaa', 2, 9, 55, NULL, NULL, 1, true),

  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeee14', 'dddddddd-dddd-dddd-dddd-dddddddddeae', 1, 12, 90, NULL, NULL, 2, true),

  ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeee15', 'dddddddd-dddd-dddd-dddd-dddddddddeaf', 1, 10, 70, NULL, NULL, 2, true);