INSERT INTO public.users (username, email, password_hash, role, is_active)
VALUES
  ('diego', 'diego@example.com', crypt('1234', gen_salt('bf')), 'admin', true),
  ('laura', 'laura@example.com', crypt('1234', gen_salt('bf')), 'user', true),
  ('sergio', 'sergio@example.com', crypt('1234', gen_salt('bf')), 'user', true),
  ('marta', 'marta@example.com', crypt('1234', gen_salt('bf')), 'user', true);

INSERT INTO public.user_profiles (user_id, first_name, last_name, age, weight_kg, height_cm, goal, experience_level)
VALUES
  (1, 'Diego', 'Martinez', 27, 78.5, 178, 'Ganar masa muscular', 'intermediate'),
  (2, 'Laura', 'Gomez', 25, 60.0, 165, 'Tonificar', 'beginner'),
  (3, 'Sergio', 'Lopez', 31, 84.2, 181, 'Perder grasa', 'intermediate'),
  (4, 'Marta', 'Ruiz', 29, 56.8, 162, 'Mejorar resistencia', 'advanced');

INSERT INTO public.body_metrics (user_id, recorded_at, weight_kg, body_fat_percentage, muscle_mass_kg, chest_cm, waist_cm, arm_cm, leg_cm, notes)
VALUES
  (1, now() - interval '10 days', 79.2, 18.5, 36.0, 102, 84, 36, 58, 'Inicio de volumen'),
  (1, now() - interval '3 days', 78.5, 17.9, 36.4, 103, 83, 36.5, 58.5, 'Buenas sensaciones'),
  (2, now() - interval '7 days', 60.0, 24.0, 22.0, 88, 72, 27, 50, 'Primer registro'),
  (3, now() - interval '5 days', 84.2, 22.5, 34.0, 100, 89, 34, 57, 'Objetivo definición');

INSERT INTO public.exercises (name, description, muscle_group, exercise_type, is_official, owner_user_id)
VALUES
  ('Press banca', 'Ejercicio básico de pecho con barra', 'chest', 'strength', true, NULL),
  ('Sentadilla', 'Ejercicio básico de pierna con barra', 'legs', 'strength', true, NULL),
  ('Peso muerto', 'Ejercicio compuesto de cadena posterior', 'back', 'strength', true, NULL),
  ('Dominadas', 'Ejercicio de tracción vertical', 'back', 'bodyweight', true, NULL),
  ('Press militar', 'Empuje vertical para hombros', 'shoulders', 'strength', true, NULL),
  ('Curl bíceps mancuerna', 'Aislamiento de bíceps', 'biceps', 'hypertrophy', true, NULL),
  ('Plancha abdominal', 'Trabajo isométrico de core', 'core', 'isometric', true, NULL),
  ('Hip thrust', 'Trabajo de glúteo', 'glutes', 'strength', true, NULL),
  ('Remo con barra', 'Tracción horizontal', 'back', 'strength', true, NULL),
  ('Flexiones diamante', 'Variante enfocada en tríceps', 'triceps', 'bodyweight', false, 1);

INSERT INTO public.exercise_secondary_muscle_groups (exercise_id, muscle_group)
VALUES
  (1, 'triceps'),
  (1, 'shoulders'),
  (2, 'glutes'),
  (3, 'legs'),
  (4, 'biceps'),
  (5, 'triceps'),
  (8, 'legs'),
  (9, 'biceps'),
  (10, 'chest');

INSERT INTO public.friendships (requester_id, addressee_id, status)
VALUES
  (1, 2, 'accepted'),
  (1, 3, 'pending'),
  (2, 4, 'accepted');

INSERT INTO public.routines (user_id, name, description, is_predefined, is_public)
VALUES
  (1, 'Push Pull Legs', 'Rutina dividida en 3 días', false, true),
  (2, 'Full Body Inicio', 'Rutina sencilla de cuerpo completo', false, true),
  (1, 'Torso Pierna', 'Rutina de 4 días', true, true),
  (3, 'Definición Express', 'Rutina con superseries y cardio', false, false);

INSERT INTO public.routine_exercises (routine_id, exercise_id, exercise_order, notes)
VALUES
  (1, 1, 1, '4 series de 6-8 reps'),
  (1, 5, 2, '3 series de 8-10 reps'),
  (1, 10, 3, '3 series al fallo técnico'),
  (2, 2, 1, '3 series de 10 reps'),
  (2, 4, 2, '3 series de 6-8 reps asistidas si hace falta'),
  (2, 7, 3, '3 bloques de 30 segundos'),
  (3, 1, 1, 'Día torso'),
  (3, 9, 2, 'Controlar técnica'),
  (3, 2, 3, 'Día pierna'),
  (4, 8, 1, 'Pausa arriba'),
  (4, 7, 2, 'Core al final');

INSERT INTO public.shared_routines (routine_id, owner_user_id, shared_with_user_id)
VALUES
  (1, 1, 2),
  (1, 1, 3),
  (2, 2, 4);

INSERT INTO public.support_tickets (user_id, title, description, status)
VALUES
  (2, 'No puedo cambiar mi avatar', 'Cuando intento subir una imagen da error.', 'open'),
  (3, 'Rutina no aparece compartida', 'Mi amigo dice que me la ha compartido pero no la veo.', 'in_progress'),
  (4, 'Error al guardar medidas', 'Se queda cargando al guardar body metrics.', 'closed');

INSERT INTO public.workout_sessions (user_id, routine_id, name, started_at, ended_at, duration_minutes, calories_burned, notes)
VALUES
  (1, 1, 'Push Day 1', now() - interval '2 days', now() - interval '2 days' + interval '70 minutes', 70, 540, 'Buen rendimiento'),
  (2, 2, 'Full Body lunes', now() - interval '1 day', now() - interval '1 day' + interval '55 minutes', 55, 320, 'Primer entreno completado'),
  (3, 4, 'Definición circuito', now() - interval '3 days', now() - interval '3 days' + interval '60 minutes', 60, 470, 'Mucho cardio final');

INSERT INTO public.workout_exercises (workout_session_id, exercise_id, exercise_order, notes)
VALUES
  (1, 1, 1, 'Muy buenas sensaciones'),
  (1, 5, 2, 'Última serie muy dura'),
  (1, 10, 3, 'Cerca del fallo'),
  (2, 2, 1, 'Técnica correcta'),
  (2, 4, 2, 'Con goma de asistencia'),
  (2, 7, 3, 'Core estable'),
  (3, 8, 1, 'Buena activación'),
  (3, 7, 2, 'Core fatigado');

INSERT INTO public.workout_sets (workout_exercise_id, set_number, reps, weight_kg, duration_seconds, distance_km, rir, completed)
VALUES
  (1, 1, 8, 80, NULL, NULL, 2, true),
  (1, 2, 8, 80, NULL, NULL, 1, true),
  (1, 3, 7, 82.5, NULL, NULL, 1, true),
  (2, 1, 10, 40, NULL, NULL, 2, true),
  (2, 2, 9, 42.5, NULL, NULL, 1, true),
  (3, 1, 15, NULL, NULL, NULL, 1, true),
  (4, 1, 10, 50, NULL, NULL, 2, true),
  (5, 1, 6, NULL, NULL, NULL, 3, true),
  (6, 1, NULL, NULL, 45, NULL, 2, true),
  (7, 1, 12, 90, NULL, NULL, 2, true),
  (8, 1, NULL, NULL, 60, NULL, 1, true);
