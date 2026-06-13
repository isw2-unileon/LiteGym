-- 01-required-base-seed.sql
-- Datos base necesarios para que docker_massive_seed.sql pueda ejecutarse.
-- Ejecutar ANTES de 02-seed.sql.

BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

INSERT INTO public.users (id, username, email, password_hash, role, is_active)
VALUES
  ('11111111-1111-1111-1111-111111111111', 'diego', 'diego@example.com', crypt('1234', gen_salt('bf')), 'admin', true),
  ('22222222-2222-2222-2222-222222222222', 'laura', 'laura@example.com', crypt('1234', gen_salt('bf')), 'user', true),
  ('33333333-3333-3333-3333-333333333333', 'sergio', 'sergio@example.com', crypt('1234', gen_salt('bf')), 'user', true),
  ('44444444-4444-4444-4444-444444444444', 'marta', 'marta@example.com', crypt('1234', gen_salt('bf')), 'user', true)
ON CONFLICT (id) DO NOTHING;

INSERT INTO public.user_profiles (
  user_id, first_name, last_name, age, weight_kg, height_cm, goal, experience_level
)
VALUES
  ('11111111-1111-1111-1111-111111111111', 'Diego', 'Martinez', 27, 78.5, 178, 'Ganar masa muscular', 'intermediate'),
  ('22222222-2222-2222-2222-222222222222', 'Laura', 'Gomez', 25, 60.0, 165, 'Tonificar', 'beginner'),
  ('33333333-3333-3333-3333-333333333333', 'Sergio', 'Lopez', 31, 84.2, 181, 'Perder grasa', 'intermediate'),
  ('44444444-4444-4444-4444-444444444444', 'Marta', 'Ruiz', 29, 56.8, 162, 'Mejorar resistencia', 'advanced')
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO public.exercises (
  id, name, description, muscle_group, exercise_type, is_official, owner_user_id
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
  ('66666666-6666-6666-6666-666666666670', 'Flexiones diamante', 'Variante enfocada en tríceps', 'triceps', 'bodyweight', false, '11111111-1111-1111-1111-111111111111')
ON CONFLICT (id) DO NOTHING;

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
  ('66666666-6666-6666-6666-666666666670', 'chest')
ON CONFLICT DO NOTHING;

INSERT INTO public.routines (id, user_id, name, description, routine_type, is_predefined, is_public)
VALUES
  ('88888888-8888-8888-8888-888888888881', '11111111-1111-1111-1111-111111111111', 'Push Pull Legs', 'Rutina dividida en 3 días', 'Fuerza', false, true),
  ('88888888-8888-8888-8888-888888888882', '22222222-2222-2222-2222-222222222222', 'Full Body Inicio', 'Rutina sencilla de cuerpo completo', 'Sin clasificar', false, true),
  ('88888888-8888-8888-8888-888888888883', '11111111-1111-1111-1111-111111111111', 'Torso Pierna', 'Rutina de 4 días', 'Fuerza', true, true),
  ('88888888-8888-8888-8888-888888888884', '33333333-3333-3333-3333-333333333333', 'Definición Express', 'Rutina con superseries y cardio', 'Resistencia', false, false)
ON CONFLICT (id) DO NOTHING;

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
  ('99999999-9999-9999-9999-99999999999b', '88888888-8888-8888-8888-888888888884', '66666666-6666-6666-6666-666666666667', 2, 'Core al final')
ON CONFLICT (id) DO NOTHING;

COMMIT;
-- docker_massive_seed.sql
-- Seed masivo para PostgreSQL/Supabase local en Docker.
-- Ejecutar DESPUÉS de crear el schema y DESPUÉS de tu seed base.
-- No crea usuarios nuevos: solo genera datos para Diego, Laura, Sergio y Marta.
-- Compatible con psql y con /docker-entrypoint-initdb.d si el schema ya existe antes.

BEGIN;

-- UUID determinista sin depender de extensiones adicionales.
CREATE OR REPLACE FUNCTION pg_temp.seed_uuid(input text)
RETURNS uuid
LANGUAGE sql
IMMUTABLE
AS $$
  SELECT (
    substr(md5(input), 1, 8) || '-' ||
    substr(md5(input), 9, 4) || '-' ||
    substr(md5(input), 13, 4) || '-' ||
    substr(md5(input), 17, 4) || '-' ||
    substr(md5(input), 21, 12)
  )::uuid;
$$;

-- Rutinas extra para que cada usuario tenga más variedad histórica.
INSERT INTO public.routines (id, user_id, name, description, routine_type, is_predefined, is_public)
VALUES
  (pg_temp.seed_uuid('docker:routine:diego:push_fuerza'), '11111111-1111-1111-1111-111111111111', 'Push fuerza largo', 'Bloque de fuerza de empuje para histórico docker', 'Fuerza', false, true),
  (pg_temp.seed_uuid('docker:routine:diego:pull_hipertrofia'), '11111111-1111-1111-1111-111111111111', 'Pull hipertrofia largo', 'Bloque de espalda y bíceps para histórico docker', 'Fuerza', false, true),
  (pg_temp.seed_uuid('docker:routine:laura:gluteo_core'), '22222222-2222-2222-2222-222222222222', 'Glúteo core largo', 'Bloque de tren inferior y core para histórico docker', 'Fuerza', false, false),
  (pg_temp.seed_uuid('docker:routine:laura:full_body'), '22222222-2222-2222-2222-222222222222', 'Full body constante', 'Rutina general para varios meses de uso', 'Sin clasificar', false, false),
  (pg_temp.seed_uuid('docker:routine:sergio:definicion'), '33333333-3333-3333-3333-333333333333', 'Definición larga', 'Trabajo de fuerza y gasto calórico para histórico docker', 'Resistencia', false, false),
  (pg_temp.seed_uuid('docker:routine:sergio:torso_pierna'), '33333333-3333-3333-3333-333333333333', 'Torso pierna definición', 'Rutina dividida para varios meses', 'Fuerza', false, false),
  (pg_temp.seed_uuid('docker:routine:marta:resistencia'), '44444444-4444-4444-4444-444444444444', 'Resistencia avanzada larga', 'Circuitos de resistencia y core', 'Resistencia', false, true),
  (pg_temp.seed_uuid('docker:routine:marta:movilidad'), '44444444-4444-4444-4444-444444444444', 'Movilidad avanzada larga', 'Movilidad, core y fuerza ligera', 'Movilidad', false, false)
ON CONFLICT (id) DO NOTHING;

-- Ejercicios planificados para esas rutinas extra.
INSERT INTO public.routine_exercises (id, routine_id, exercise_id, exercise_order, notes)
VALUES
  (pg_temp.seed_uuid('docker:re:diego:push:1'), pg_temp.seed_uuid('docker:routine:diego:push_fuerza'), '66666666-6666-6666-6666-666666666661', 1, 'Press principal'),
  (pg_temp.seed_uuid('docker:re:diego:push:2'), pg_temp.seed_uuid('docker:routine:diego:push_fuerza'), '66666666-6666-6666-6666-666666666665', 2, 'Press vertical'),
  (pg_temp.seed_uuid('docker:re:diego:push:3'), pg_temp.seed_uuid('docker:routine:diego:push_fuerza'), '66666666-6666-6666-6666-666666666670', 3, 'Trabajo final de tríceps'),
  (pg_temp.seed_uuid('docker:re:diego:pull:1'), pg_temp.seed_uuid('docker:routine:diego:pull_hipertrofia'), '66666666-6666-6666-6666-666666666669', 1, 'Remo principal'),
  (pg_temp.seed_uuid('docker:re:diego:pull:2'), pg_temp.seed_uuid('docker:routine:diego:pull_hipertrofia'), '66666666-6666-6666-6666-666666666664', 2, 'Tracción vertical'),
  (pg_temp.seed_uuid('docker:re:diego:pull:3'), pg_temp.seed_uuid('docker:routine:diego:pull_hipertrofia'), '66666666-6666-6666-6666-666666666666', 3, 'Bíceps final'),

  (pg_temp.seed_uuid('docker:re:laura:gluteo:1'), pg_temp.seed_uuid('docker:routine:laura:gluteo_core'), '66666666-6666-6666-6666-666666666668', 1, 'Hip thrust principal'),
  (pg_temp.seed_uuid('docker:re:laura:gluteo:2'), pg_temp.seed_uuid('docker:routine:laura:gluteo_core'), '66666666-6666-6666-6666-666666666662', 2, 'Sentadilla técnica'),
  (pg_temp.seed_uuid('docker:re:laura:gluteo:3'), pg_temp.seed_uuid('docker:routine:laura:gluteo_core'), '66666666-6666-6666-6666-666666666667', 3, 'Core final'),
  (pg_temp.seed_uuid('docker:re:laura:full:1'), pg_temp.seed_uuid('docker:routine:laura:full_body'), '66666666-6666-6666-6666-666666666661', 1, 'Empuje base'),
  (pg_temp.seed_uuid('docker:re:laura:full:2'), pg_temp.seed_uuid('docker:routine:laura:full_body'), '66666666-6666-6666-6666-666666666664', 2, 'Espalda asistida'),
  (pg_temp.seed_uuid('docker:re:laura:full:3'), pg_temp.seed_uuid('docker:routine:laura:full_body'), '66666666-6666-6666-6666-666666666667', 3, 'Core'),

  (pg_temp.seed_uuid('docker:re:sergio:def:1'), pg_temp.seed_uuid('docker:routine:sergio:definicion'), '66666666-6666-6666-6666-666666666669', 1, 'Remo denso'),
  (pg_temp.seed_uuid('docker:re:sergio:def:2'), pg_temp.seed_uuid('docker:routine:sergio:definicion'), '66666666-6666-6666-6666-666666666668', 2, 'Glúteo y gasto'),
  (pg_temp.seed_uuid('docker:re:sergio:def:3'), pg_temp.seed_uuid('docker:routine:sergio:definicion'), '66666666-6666-6666-6666-666666666667', 3, 'Core metabólico'),
  (pg_temp.seed_uuid('docker:re:sergio:tp:1'), pg_temp.seed_uuid('docker:routine:sergio:torso_pierna'), '66666666-6666-6666-6666-666666666661', 1, 'Torso pesado'),
  (pg_temp.seed_uuid('docker:re:sergio:tp:2'), pg_temp.seed_uuid('docker:routine:sergio:torso_pierna'), '66666666-6666-6666-6666-666666666662', 2, 'Pierna'),
  (pg_temp.seed_uuid('docker:re:sergio:tp:3'), pg_temp.seed_uuid('docker:routine:sergio:torso_pierna'), '66666666-6666-6666-6666-666666666663', 3, 'Cadena posterior'),

  (pg_temp.seed_uuid('docker:re:marta:res:1'), pg_temp.seed_uuid('docker:routine:marta:resistencia'), '66666666-6666-6666-6666-666666666664', 1, 'Tracción'),
  (pg_temp.seed_uuid('docker:re:marta:res:2'), pg_temp.seed_uuid('docker:routine:marta:resistencia'), '66666666-6666-6666-6666-666666666670', 2, 'Empuje peso corporal'),
  (pg_temp.seed_uuid('docker:re:marta:res:3'), pg_temp.seed_uuid('docker:routine:marta:resistencia'), '66666666-6666-6666-6666-666666666667', 3, 'Core largo'),
  (pg_temp.seed_uuid('docker:re:marta:mov:1'), pg_temp.seed_uuid('docker:routine:marta:movilidad'), '66666666-6666-6666-6666-666666666667', 1, 'Core controlado'),
  (pg_temp.seed_uuid('docker:re:marta:mov:2'), pg_temp.seed_uuid('docker:routine:marta:movilidad'), '66666666-6666-6666-6666-666666666668', 2, 'Activación'),
  (pg_temp.seed_uuid('docker:re:marta:mov:3'), pg_temp.seed_uuid('docker:routine:marta:movilidad'), '66666666-6666-6666-6666-666666666662', 3, 'Sentadilla ligera')
ON CONFLICT (id) DO NOTHING;

-- Series planificadas para todas las routine_exercises docker.
INSERT INTO public.routine_exercise_sets (
  id, routine_exercise_id, set_number,
  target_reps_min, target_reps_max, target_weight_kg, target_duration_seconds,
  target_rir, rest_seconds, notes
)
SELECT
  pg_temp.seed_uuid('docker:planned-set:' || re.id::text || ':' || gs.set_number),
  re.id,
  gs.set_number,
  CASE WHEN re.exercise_id = '66666666-6666-6666-6666-666666666667' THEN NULL ELSE 8 END,
  CASE WHEN re.exercise_id = '66666666-6666-6666-6666-666666666667' THEN NULL ELSE 12 END,
  CASE
    WHEN re.exercise_id = '66666666-6666-6666-6666-666666666661' THEN 70 + gs.set_number * 2.5
    WHEN re.exercise_id = '66666666-6666-6666-6666-666666666662' THEN 55 + gs.set_number * 2.5
    WHEN re.exercise_id = '66666666-6666-6666-6666-666666666663' THEN 95 + gs.set_number * 5
    WHEN re.exercise_id = '66666666-6666-6666-6666-666666666668' THEN 60 + gs.set_number * 5
    WHEN re.exercise_id = '66666666-6666-6666-6666-666666666669' THEN 55 + gs.set_number * 2.5
    WHEN re.exercise_id = '66666666-6666-6666-6666-666666666666' THEN 12.5 + gs.set_number * 1.25
    ELSE NULL
  END,
  CASE WHEN re.exercise_id = '66666666-6666-6666-6666-666666666667' THEN 40 + gs.set_number * 10 ELSE NULL END,
  CASE WHEN gs.set_number = 1 THEN 3 WHEN gs.set_number = 2 THEN 2 ELSE 1 END,
  CASE WHEN re.exercise_id = '66666666-6666-6666-6666-666666666667' THEN 45 ELSE 90 END,
  NULL
FROM public.routine_exercises re
JOIN generate_series(1, 4) AS gs(set_number) ON true
WHERE re.id IN (
  pg_temp.seed_uuid('docker:re:diego:push:1'), pg_temp.seed_uuid('docker:re:diego:push:2'), pg_temp.seed_uuid('docker:re:diego:push:3'),
  pg_temp.seed_uuid('docker:re:diego:pull:1'), pg_temp.seed_uuid('docker:re:diego:pull:2'), pg_temp.seed_uuid('docker:re:diego:pull:3'),
  pg_temp.seed_uuid('docker:re:laura:gluteo:1'), pg_temp.seed_uuid('docker:re:laura:gluteo:2'), pg_temp.seed_uuid('docker:re:laura:gluteo:3'),
  pg_temp.seed_uuid('docker:re:laura:full:1'), pg_temp.seed_uuid('docker:re:laura:full:2'), pg_temp.seed_uuid('docker:re:laura:full:3'),
  pg_temp.seed_uuid('docker:re:sergio:def:1'), pg_temp.seed_uuid('docker:re:sergio:def:2'), pg_temp.seed_uuid('docker:re:sergio:def:3'),
  pg_temp.seed_uuid('docker:re:sergio:tp:1'), pg_temp.seed_uuid('docker:re:sergio:tp:2'), pg_temp.seed_uuid('docker:re:sergio:tp:3'),
  pg_temp.seed_uuid('docker:re:marta:res:1'), pg_temp.seed_uuid('docker:re:marta:res:2'), pg_temp.seed_uuid('docker:re:marta:res:3'),
  pg_temp.seed_uuid('docker:re:marta:mov:1'), pg_temp.seed_uuid('docker:re:marta:mov:2'), pg_temp.seed_uuid('docker:re:marta:mov:3')
)
ON CONFLICT (id) DO NOTHING;

-- Historial corporal: 52 semanas por usuario.
DO $$
DECLARE
  u record;
  week_index int;
  base_date timestamptz := date_trunc('day', now()) - interval '364 days';
  recorded timestamptz;
  weight numeric;
  fat numeric;
  muscle numeric;
  chest numeric;
  waist numeric;
  arm numeric;
  leg numeric;
BEGIN
  FOR u IN
    SELECT * FROM (VALUES
      ('11111111-1111-1111-1111-111111111111'::uuid, 'Diego', 80.2, -0.018, 20.5, -0.040, 35.0, 0.035, 101.0, 86.0, 35.0, 57.0),
      ('22222222-2222-2222-2222-222222222222'::uuid, 'Laura', 61.8, -0.010, 25.0, -0.030, 21.6, 0.020, 87.0, 73.0, 26.5, 49.5),
      ('33333333-3333-3333-3333-333333333333'::uuid, 'Sergio', 88.0, -0.075, 26.0, -0.070, 33.2, 0.025, 101.0, 94.0, 34.0, 56.0),
      ('44444444-4444-4444-4444-444444444444'::uuid, 'Marta', 56.4, 0.004, 20.5, -0.010, 22.5, 0.018, 86.5, 68.5, 25.5, 48.5)
    ) AS t(user_id, name, start_weight, weekly_weight_delta, start_fat, weekly_fat_delta, start_muscle, weekly_muscle_delta, start_chest, start_waist, start_arm, start_leg)
  LOOP
    FOR week_index IN 0..51 LOOP
      recorded := base_date + make_interval(days => week_index * 7) + interval '8 hours';
      weight := round((u.start_weight + u.weekly_weight_delta * week_index + ((week_index % 5) - 2) * 0.12)::numeric, 1);
      fat := round((u.start_fat + u.weekly_fat_delta * week_index + ((week_index % 4) - 1) * 0.08)::numeric, 1);
      muscle := round((u.start_muscle + u.weekly_muscle_delta * week_index + ((week_index % 3) - 1) * 0.05)::numeric, 1);
      chest := round((u.start_chest + week_index * 0.025 + ((week_index % 4) - 1) * 0.1)::numeric, 1);
      waist := round((u.start_waist + CASE WHEN u.name IN ('Sergio','Diego','Laura') THEN -week_index * 0.055 ELSE -week_index * 0.015 END + ((week_index % 3) - 1) * 0.08)::numeric, 1);
      arm := round((u.start_arm + week_index * 0.015 + ((week_index % 4) - 1) * 0.04)::numeric, 1);
      leg := round((u.start_leg + week_index * 0.018 + ((week_index % 5) - 2) * 0.04)::numeric, 1);

      INSERT INTO public.body_metrics (
        id, user_id, recorded_at, weight_kg, body_fat_percentage, muscle_mass_kg,
        chest_cm, waist_cm, arm_cm, leg_cm, notes
      ) VALUES (
        pg_temp.seed_uuid('docker:body:' || u.user_id::text || ':' || week_index),
        u.user_id,
        recorded,
        weight,
        fat,
        muscle,
        chest,
        waist,
        arm,
        leg,
        CASE
          WHEN week_index = 0 THEN 'Inicio histórico docker'
          WHEN week_index % 13 = 0 THEN 'Revisión trimestral'
          WHEN week_index % 4 = 0 THEN 'Control mensual'
          ELSE 'Registro semanal'
        END
      ) ON CONFLICT (id) DO NOTHING;
    END LOOP;
  END LOOP;
END $$;

-- Historial de entrenamientos: unas 650-750 sesiones, 2k+ ejercicios y 7k+ series.
DO $$
DECLARE
  u record;
  d date;
  day_offset int;
  session_id uuid;
  workout_exercise_id uuid;
  selected_routine uuid;
  selected_exercises uuid[];
  exercise_id uuid;
  exercise_idx int;
  set_idx int;
  set_count int;
  base_weight numeric;
  progress numeric;
  reps_value int;
  duration_value int;
  calories int;
  session_minutes int;
  session_name text;
  note_text text;
BEGIN
  FOR u IN
    SELECT * FROM (VALUES
      ('11111111-1111-1111-1111-111111111111'::uuid, 'Diego', 0, ARRAY[
        pg_temp.seed_uuid('docker:routine:diego:push_fuerza'),
        pg_temp.seed_uuid('docker:routine:diego:pull_hipertrofia'),
        '88888888-8888-8888-8888-888888888881'::uuid,
        '88888888-8888-8888-8888-888888888883'::uuid
      ]),
      ('22222222-2222-2222-2222-222222222222'::uuid, 'Laura', 1, ARRAY[
        pg_temp.seed_uuid('docker:routine:laura:gluteo_core'),
        pg_temp.seed_uuid('docker:routine:laura:full_body'),
        '88888888-8888-8888-8888-888888888882'::uuid
      ]),
      ('33333333-3333-3333-3333-333333333333'::uuid, 'Sergio', 2, ARRAY[
        pg_temp.seed_uuid('docker:routine:sergio:definicion'),
        pg_temp.seed_uuid('docker:routine:sergio:torso_pierna'),
        '88888888-8888-8888-8888-888888888884'::uuid
      ]),
      ('44444444-4444-4444-4444-444444444444'::uuid, 'Marta', 3, ARRAY[
        pg_temp.seed_uuid('docker:routine:marta:resistencia'),
        pg_temp.seed_uuid('docker:routine:marta:movilidad')
      ])
    ) AS t(user_id, name, user_phase, routines)
  LOOP
    FOR day_offset IN 0..364 LOOP
      d := (date_trunc('day', now())::date - 364 + day_offset);

      -- Patrón realista: cada persona entrena 3-5 días por semana, con semanas de descarga.
      IF ((day_offset + u.user_phase) % 7 IN (0, 2, 4)
          OR (u.name IN ('Diego', 'Marta') AND (day_offset + u.user_phase) % 7 = 5)
          OR (u.name = 'Sergio' AND day_offset % 14 = 9))
         AND NOT (day_offset % 47 IN (0, 1))
      THEN
        selected_routine := u.routines[1 + ((day_offset + u.user_phase) % array_length(u.routines, 1))];
        session_id := pg_temp.seed_uuid('docker:session:' || u.user_id::text || ':' || d::text);
        session_minutes := 42 + ((day_offset + u.user_phase) % 38);
        calories := CASE u.name
          WHEN 'Diego' THEN 430 + ((day_offset * 7) % 190)
          WHEN 'Laura' THEN 260 + ((day_offset * 5) % 130)
          WHEN 'Sergio' THEN 420 + ((day_offset * 9) % 230)
          ELSE 330 + ((day_offset * 6) % 180)
        END;
        session_name := CASE
          WHEN u.name = 'Diego' AND day_offset % 4 = 0 THEN 'Push fuerza'
          WHEN u.name = 'Diego' THEN 'Pull y torso'
          WHEN u.name = 'Laura' AND day_offset % 3 = 0 THEN 'Glúteo y core'
          WHEN u.name = 'Laura' THEN 'Full body'
          WHEN u.name = 'Sergio' AND day_offset % 3 = 0 THEN 'Definición circuito'
          WHEN u.name = 'Sergio' THEN 'Torso pierna'
          WHEN u.name = 'Marta' AND day_offset % 3 = 0 THEN 'Resistencia avanzada'
          ELSE 'Core y movilidad'
        END;
        note_text := CASE
          WHEN day_offset % 31 = 0 THEN 'Sesión más suave por fatiga acumulada'
          WHEN day_offset % 17 = 0 THEN 'Buen ritmo y técnica estable'
          WHEN day_offset % 11 = 0 THEN 'Subida ligera de carga'
          ELSE 'Entrenamiento completado desde seed docker'
        END;

        INSERT INTO public.workout_sessions (
          id, user_id, routine_id, name, performed_at, planned_at,
          duration_minutes, calories_burned, notes
        ) VALUES (
          session_id,
          u.user_id,
          selected_routine,
          session_name,
          d::timestamptz + make_interval(hours => 7 + ((day_offset + u.user_phase) % 13)),
          CASE WHEN day_offset % 6 = 0 THEN d::timestamptz + interval '7 hours' ELSE NULL END,
          session_minutes,
          calories,
          note_text
        ) ON CONFLICT (id) DO NOTHING;

        selected_exercises := CASE
          WHEN u.name = 'Diego' AND day_offset % 2 = 0 THEN ARRAY[
            '66666666-6666-6666-6666-666666666661'::uuid,
            '66666666-6666-6666-6666-666666666665'::uuid,
            '66666666-6666-6666-6666-666666666670'::uuid,
            '66666666-6666-6666-6666-666666666667'::uuid
          ]
          WHEN u.name = 'Diego' THEN ARRAY[
            '66666666-6666-6666-6666-666666666669'::uuid,
            '66666666-6666-6666-6666-666666666664'::uuid,
            '66666666-6666-6666-6666-666666666666'::uuid,
            '66666666-6666-6666-6666-666666666663'::uuid
          ]
          WHEN u.name = 'Laura' THEN ARRAY[
            '66666666-6666-6666-6666-666666666668'::uuid,
            '66666666-6666-6666-6666-666666666662'::uuid,
            '66666666-6666-6666-6666-666666666667'::uuid
          ]
          WHEN u.name = 'Sergio' THEN ARRAY[
            '66666666-6666-6666-6666-666666666669'::uuid,
            '66666666-6666-6666-6666-666666666661'::uuid,
            '66666666-6666-6666-6666-666666666668'::uuid,
            '66666666-6666-6666-6666-666666666667'::uuid
          ]
          ELSE ARRAY[
            '66666666-6666-6666-6666-666666666664'::uuid,
            '66666666-6666-6666-6666-666666666670'::uuid,
            '66666666-6666-6666-6666-666666666667'::uuid,
            '66666666-6666-6666-6666-666666666662'::uuid
          ]
        END;

        FOR exercise_idx IN 1..array_length(selected_exercises, 1) LOOP
          exercise_id := selected_exercises[exercise_idx];
          workout_exercise_id := pg_temp.seed_uuid('docker:workout-exercise:' || session_id::text || ':' || exercise_idx);

          INSERT INTO public.workout_exercises (
            id, workout_session_id, exercise_id, exercise_order, notes
          ) VALUES (
            workout_exercise_id,
            session_id,
            exercise_id,
            exercise_idx,
            CASE
              WHEN exercise_idx = 1 THEN 'Ejercicio principal'
              WHEN exercise_id = '66666666-6666-6666-6666-666666666667' THEN 'Core al final'
              ELSE 'Trabajo complementario'
            END
          ) ON CONFLICT (id) DO NOTHING;

          set_count := CASE WHEN exercise_id = '66666666-6666-6666-6666-666666666667' THEN 3 ELSE 4 END;
          progress := day_offset / 365.0;
          base_weight := CASE exercise_id::text
            WHEN '66666666-6666-6666-6666-666666666661' THEN CASE u.name WHEN 'Diego' THEN 72 WHEN 'Sergio' THEN 62 ELSE 32 END
            WHEN '66666666-6666-6666-6666-666666666662' THEN CASE u.name WHEN 'Diego' THEN 80 WHEN 'Laura' THEN 42 WHEN 'Sergio' THEN 72 ELSE 38 END
            WHEN '66666666-6666-6666-6666-666666666663' THEN CASE u.name WHEN 'Diego' THEN 105 WHEN 'Sergio' THEN 92 ELSE 55 END
            WHEN '66666666-6666-6666-6666-666666666665' THEN CASE u.name WHEN 'Diego' THEN 38 ELSE 22 END
            WHEN '66666666-6666-6666-6666-666666666666' THEN CASE u.name WHEN 'Diego' THEN 12 ELSE 9 END
            WHEN '66666666-6666-6666-6666-666666666668' THEN CASE u.name WHEN 'Laura' THEN 52 WHEN 'Marta' THEN 48 ELSE 70 END
            WHEN '66666666-6666-6666-6666-666666666669' THEN CASE u.name WHEN 'Diego' THEN 55 WHEN 'Sergio' THEN 50 ELSE 35 END
            ELSE NULL
          END;

          FOR set_idx IN 1..set_count LOOP
            reps_value := CASE
              WHEN exercise_id = '66666666-6666-6666-6666-666666666667' THEN NULL
              WHEN exercise_id IN ('66666666-6666-6666-6666-666666666664','66666666-6666-6666-6666-666666666670') THEN 6 + ((day_offset + set_idx + exercise_idx) % 9)
              ELSE 6 + ((day_offset + set_idx + exercise_idx) % 6)
            END;
            duration_value := CASE
              WHEN exercise_id = '66666666-6666-6666-6666-666666666667' THEN 35 + ((day_offset + set_idx * 7) % 55)
              ELSE NULL
            END;

            INSERT INTO public.workout_sets (
              id, workout_exercise_id, set_number, reps, weight_kg,
              duration_seconds, distance_km, rir, completed
            ) VALUES (
              pg_temp.seed_uuid('docker:set:' || workout_exercise_id::text || ':' || set_idx),
              workout_exercise_id,
              set_idx,
              reps_value,
              CASE
                WHEN exercise_id IN ('66666666-6666-6666-6666-666666666664','66666666-6666-6666-6666-666666666670','66666666-6666-6666-6666-666666666667') THEN NULL
                ELSE round((base_weight + progress * CASE u.name WHEN 'Diego' THEN 18 WHEN 'Laura' THEN 9 WHEN 'Sergio' THEN 14 ELSE 7 END + set_idx * 1.25 + ((day_offset % 5) - 2) * 0.75)::numeric, 1)
              END,
              duration_value,
              NULL,
              CASE WHEN set_idx = set_count THEN 1 ELSE 2 + ((set_idx + day_offset) % 2) END,
              CASE WHEN day_offset % 53 = 0 AND set_idx = set_count THEN false ELSE true END
            ) ON CONFLICT (id) DO NOTHING;
          END LOOP;
        END LOOP;
      END IF;
    END LOOP;
  END LOOP;
END $$;

-- Tickets y actividad de soporte adicional.
INSERT INTO public.support_tickets (id, user_id, title, description, status)
VALUES
  (pg_temp.seed_uuid('docker:ticket:laura:avatar-2'), '22222222-2222-2222-2222-222222222222', 'Avatar no se actualiza en móvil', 'La imagen aparece cambiada en perfil pero no en el header de la app.', 'closed'),
  (pg_temp.seed_uuid('docker:ticket:diego:stats'), '11111111-1111-1111-1111-111111111111', 'Duda con estadísticas de volumen', 'Quiere entender por qué el volumen semanal no coincide con el detalle del ejercicio.', 'closed'),
  (pg_temp.seed_uuid('docker:ticket:sergio:rutina'), '33333333-3333-3333-3333-333333333333', 'No puedo duplicar una rutina', 'Al duplicar la rutina de definición aparece un error intermitente.', 'in_progress'),
  (pg_temp.seed_uuid('docker:ticket:marta:metricas'), '44444444-4444-4444-4444-444444444444', 'Medidas guardadas con fecha incorrecta', 'El registro semanal aparece desplazado un día.', 'open'),
  (pg_temp.seed_uuid('docker:ticket:diego:export'), '11111111-1111-1111-1111-111111111111', 'Exportación de historial', 'Solicita exportar entrenamientos por rango de fechas.', 'open'),
  (pg_temp.seed_uuid('docker:ticket:laura:plancha'), '22222222-2222-2222-2222-222222222222', 'Series por tiempo poco claras', 'Las series de plancha no muestran duración en el resumen.', 'closed'),
  (pg_temp.seed_uuid('docker:ticket:sergio:calorias'), '33333333-3333-3333-3333-333333333333', 'Calorías estimadas muy altas', 'Algunas sesiones de circuito muestran calorías por encima de lo esperado.', 'closed'),
  (pg_temp.seed_uuid('docker:ticket:marta:compartir'), '44444444-4444-4444-4444-444444444444', 'Rutina pública no visible', 'Una rutina marcada como pública no aparece en el listado compartido.', 'in_progress')
ON CONFLICT (id) DO NOTHING;

COMMIT;

-- Comprobaciones opcionales:
-- SELECT count(*) FROM public.workout_sessions;
-- SELECT count(*) FROM public.workout_exercises;
-- SELECT count(*) FROM public.workout_sets;
-- SELECT count(*) FROM public.body_metrics;
