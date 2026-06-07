DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE public.friendship_status AS ENUM ('pending', 'accepted', 'rejected');
CREATE TYPE public.ticket_status AS ENUM ('open', 'in_progress', 'closed');
CREATE TYPE public.user_role AS ENUM ('user', 'admin');
CREATE TYPE public.routine_type AS ENUM ('Fuerza', 'Movilidad', 'Resistencia', 'Sin clasificar');

CREATE TABLE public.users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username VARCHAR NOT NULL UNIQUE,
  email VARCHAR NOT NULL UNIQUE,
  password_hash TEXT,
  role public.user_role NOT NULL DEFAULT 'user',
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE public.user_profiles (
  user_id UUID PRIMARY KEY,
  first_name VARCHAR,
  last_name VARCHAR,
  age INTEGER CHECK (age >= 0),
  weight_kg NUMERIC CHECK (weight_kg > 0),
  height_cm NUMERIC CHECK (height_cm > 0),
  goal TEXT,
  experience_level VARCHAR,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT user_profiles_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE
);

CREATE TABLE public.user_goals (
  user_id UUID PRIMARY KEY,
  short_term TEXT,
  long_term TEXT,
  target_days_per_week INTEGER,
  updated_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT user_goals_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE
);

CREATE TABLE public.body_metrics (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  recorded_at TIMESTAMP NOT NULL DEFAULT now(),
  weight_kg NUMERIC CHECK (weight_kg > 0),
  body_fat_percentage NUMERIC CHECK (body_fat_percentage >= 0 AND body_fat_percentage <= 100),
  muscle_mass_kg NUMERIC CHECK (muscle_mass_kg >= 0),
  chest_cm NUMERIC CHECK (chest_cm >= 0),
  waist_cm NUMERIC CHECK (waist_cm >= 0),
  arm_cm NUMERIC CHECK (arm_cm >= 0),
  leg_cm NUMERIC CHECK (leg_cm >= 0),
  notes TEXT,
  CONSTRAINT body_metrics_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE
);

CREATE TABLE public.exercises (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR NOT NULL,
  description TEXT,
  muscle_group VARCHAR NOT NULL,
  exercise_type VARCHAR,
  is_official BOOLEAN NOT NULL DEFAULT true,
  owner_user_id UUID,
  deleted_at TIMESTAMP,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT exercises_owner_user_id_fkey
    FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON DELETE CASCADE,
  CONSTRAINT exercises_official_owner_check
    CHECK (
      (is_official = true AND owner_user_id IS NULL)
      OR
      (is_official = false AND owner_user_id IS NOT NULL)
    )
);

CREATE TABLE public.exercise_secondary_muscle_groups (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  exercise_id UUID NOT NULL,
  muscle_group VARCHAR NOT NULL,
  CONSTRAINT exercise_secondary_muscle_groups_exercise_id_fkey
    FOREIGN KEY (exercise_id) REFERENCES public.exercises(id) ON DELETE CASCADE,
  CONSTRAINT exercise_secondary_muscle_groups_unique
    UNIQUE (exercise_id, muscle_group)
);

CREATE TABLE public.friendships (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  requester_id UUID NOT NULL,
  addressee_id UUID NOT NULL,
  status public.friendship_status NOT NULL DEFAULT 'pending',
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT friendships_requester_id_fkey
    FOREIGN KEY (requester_id) REFERENCES public.users(id) ON DELETE CASCADE,
  CONSTRAINT friendships_addressee_id_fkey
    FOREIGN KEY (addressee_id) REFERENCES public.users(id) ON DELETE CASCADE
);

CREATE TABLE public.routines (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  name VARCHAR NOT NULL,
  description TEXT,
  source VARCHAR NOT NULL DEFAULT 'manual',
  routine_type public.routine_type NOT NULL DEFAULT 'Sin clasificar',
  is_predefined BOOLEAN NOT NULL DEFAULT false,
  is_public BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT routines_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE
);

CREATE TABLE public.routine_exercises (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  routine_id UUID NOT NULL,
  exercise_id UUID NOT NULL,
  exercise_order INTEGER NOT NULL CHECK (exercise_order > 0),
  notes TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT routine_exercises_routine_id_fkey
    FOREIGN KEY (routine_id) REFERENCES public.routines(id) ON DELETE CASCADE,
  CONSTRAINT routine_exercises_exercise_id_fkey
    FOREIGN KEY (exercise_id) REFERENCES public.exercises(id) ON DELETE CASCADE
);

CREATE TABLE public.routine_exercise_sets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  routine_exercise_id UUID NOT NULL,
  set_number INTEGER NOT NULL CHECK (set_number > 0),
  target_reps_min INTEGER CHECK (target_reps_min >= 0),
  target_reps_max INTEGER CHECK (target_reps_max >= 0),
  target_reps_text TEXT,
  target_weight_kg NUMERIC CHECK (target_weight_kg >= 0),
  target_duration_seconds INTEGER CHECK (target_duration_seconds >= 0),
  target_distance_km NUMERIC CHECK (target_distance_km >= 0),
  target_rir INTEGER CHECK (target_rir >= 0 AND target_rir <= 10),
  rest_seconds INTEGER CHECK (rest_seconds >= 0),
  notes TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT routine_exercise_sets_routine_exercise_id_fkey
    FOREIGN KEY (routine_exercise_id) REFERENCES public.routine_exercises(id) ON DELETE CASCADE,
  CONSTRAINT routine_exercise_sets_reps_range_check
    CHECK (
      target_reps_min IS NULL
      OR target_reps_max IS NULL
      OR target_reps_min <= target_reps_max
    )
);

CREATE TABLE public.shared_routines (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  routine_id UUID NOT NULL,
  owner_user_id UUID NOT NULL,
  shared_with_user_id UUID NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT shared_routines_routine_id_fkey
    FOREIGN KEY (routine_id) REFERENCES public.routines(id) ON DELETE CASCADE,
  CONSTRAINT shared_routines_owner_user_id_fkey
    FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON DELETE CASCADE,
  CONSTRAINT shared_routines_shared_with_user_id_fkey
    FOREIGN KEY (shared_with_user_id) REFERENCES public.users(id) ON DELETE CASCADE
);

CREATE TABLE public.support_tickets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  title VARCHAR NOT NULL,
  description TEXT NOT NULL,
  status public.ticket_status NOT NULL DEFAULT 'open',
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT support_tickets_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE
);

CREATE TABLE public.workout_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  routine_id UUID,
  name VARCHAR,
  performed_at TIMESTAMP,
  planned_at TIMESTAMP,
  duration_minutes INTEGER CHECK (duration_minutes >= 0),
  calories_burned NUMERIC CHECK (calories_burned >= 0),
  notes TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT workout_sessions_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE,
  CONSTRAINT workout_sessions_routine_id_fkey
    FOREIGN KEY (routine_id) REFERENCES public.routines(id) ON DELETE SET NULL
);

CREATE TABLE public.workout_exercises (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workout_session_id UUID NOT NULL,
  exercise_id UUID NOT NULL,
  routine_exercise_id UUID,
  exercise_order INTEGER NOT NULL CHECK (exercise_order > 0),
  notes TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT workout_exercises_workout_session_id_fkey
    FOREIGN KEY (workout_session_id) REFERENCES public.workout_sessions(id) ON DELETE CASCADE,
  CONSTRAINT workout_exercises_exercise_id_fkey
    FOREIGN KEY (exercise_id) REFERENCES public.exercises(id) ON DELETE CASCADE,
  CONSTRAINT workout_exercises_routine_exercise_id_fkey
    FOREIGN KEY (routine_exercise_id) REFERENCES public.routine_exercises(id) ON DELETE SET NULL
);

CREATE TABLE public.workout_sets (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  workout_exercise_id UUID NOT NULL,
  routine_exercise_set_id UUID,
  set_number INTEGER NOT NULL CHECK (set_number > 0),
  target_reps_min INTEGER CHECK (target_reps_min >= 0),
  target_reps_max INTEGER CHECK (target_reps_max >= 0),
  target_reps_text TEXT,
  target_weight_kg NUMERIC CHECK (target_weight_kg >= 0),
  target_duration_seconds INTEGER CHECK (target_duration_seconds >= 0),
  target_distance_km NUMERIC CHECK (target_distance_km >= 0),
  target_rir INTEGER CHECK (target_rir >= 0 AND target_rir <= 10),
  rest_seconds INTEGER CHECK (rest_seconds >= 0),
  reps INTEGER CHECK (reps >= 0),
  weight_kg NUMERIC CHECK (weight_kg >= 0),
  duration_seconds INTEGER CHECK (duration_seconds >= 0),
  distance_km NUMERIC CHECK (distance_km >= 0),
  rir INTEGER CHECK (rir >= 0 AND rir <= 10),
  completed BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT workout_sets_workout_exercise_id_fkey
    FOREIGN KEY (workout_exercise_id) REFERENCES public.workout_exercises(id) ON DELETE CASCADE,
  CONSTRAINT workout_sets_routine_exercise_set_id_fkey
    FOREIGN KEY (routine_exercise_set_id) REFERENCES public.routine_exercise_sets(id) ON DELETE SET NULL,
  CONSTRAINT workout_sets_target_reps_range_check
    CHECK (
      target_reps_min IS NULL
      OR target_reps_max IS NULL
      OR target_reps_min <= target_reps_max
    )
);

CREATE UNIQUE INDEX exercises_official_name_unique
ON public.exercises (LOWER(name))
WHERE is_official = true AND deleted_at IS NULL;

CREATE UNIQUE INDEX exercises_private_owner_name_unique
ON public.exercises (owner_user_id, LOWER(name))
WHERE is_official = false AND deleted_at IS NULL;

CREATE TABLE public.ai_routine_generation_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT ai_routine_generation_logs_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE
);

CREATE INDEX ai_routine_generation_logs_user_created_at_idx
ON public.ai_routine_generation_logs (user_id, created_at DESC);
