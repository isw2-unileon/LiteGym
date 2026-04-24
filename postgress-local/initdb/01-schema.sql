DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE public.friendship_status AS ENUM ('pending', 'accepted', 'rejected');
CREATE TYPE public.ticket_status AS ENUM ('open', 'in_progress', 'closed');
CREATE TYPE public.user_role AS ENUM ('user', 'admin');

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
    FOREIGN KEY (user_id) REFERENCES public.users(id)
);

CREATE TABLE public.body_metrics (
  id BIGSERIAL PRIMARY KEY,
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
    FOREIGN KEY (user_id) REFERENCES public.users(id)
);

CREATE TABLE public.exercises (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR NOT NULL,
  description TEXT,
  muscle_group VARCHAR NOT NULL,
  secondary_muscle_group VARCHAR DEFAULT '' NOT NULL,
  exercise_type VARCHAR,
  is_official BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  owner_user_id UUID,
  CONSTRAINT exercises_owner_user_id_fkey
    FOREIGN KEY (owner_user_id) REFERENCES public.users(id)
);

CREATE TABLE public.exercise_secondary_muscle_groups (
  id BIGSERIAL PRIMARY KEY,
  exercise_id BIGINT NOT NULL,
  muscle_group VARCHAR NOT NULL,
  CONSTRAINT exercise_secondary_muscle_groups_exercise_id_fkey
    FOREIGN KEY (exercise_id) REFERENCES public.exercises(id),
  CONSTRAINT exercise_secondary_muscle_groups_unique
    UNIQUE (exercise_id, muscle_group)
);

CREATE TABLE public.friendships (
  id BIGSERIAL PRIMARY KEY,
  requester_id UUID NOT NULL,
  addressee_id UUID NOT NULL,
  status public.friendship_status NOT NULL DEFAULT 'pending',
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT friendships_requester_id_fkey
    FOREIGN KEY (requester_id) REFERENCES public.users(id),
  CONSTRAINT friendships_addressee_id_fkey
    FOREIGN KEY (addressee_id) REFERENCES public.users(id)
);

CREATE TABLE public.routines (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL,
  name VARCHAR NOT NULL,
  description TEXT,
  is_predefined BOOLEAN NOT NULL DEFAULT false,
  is_public BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT routines_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id)
);

CREATE TABLE public.routine_exercises (
  id BIGSERIAL PRIMARY KEY,
  routine_id BIGINT NOT NULL,
  exercise_id BIGINT NOT NULL,
  exercise_order INTEGER NOT NULL CHECK (exercise_order > 0),
  notes TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT routine_exercises_routine_id_fkey
    FOREIGN KEY (routine_id) REFERENCES public.routines(id),
  CONSTRAINT routine_exercises_exercise_id_fkey
    FOREIGN KEY (exercise_id) REFERENCES public.exercises(id)
);

CREATE TABLE public.shared_routines (
  id BIGSERIAL PRIMARY KEY,
  routine_id BIGINT NOT NULL,
  owner_user_id UUID NOT NULL,
  shared_with_user_id UUID NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT shared_routines_routine_id_fkey
    FOREIGN KEY (routine_id) REFERENCES public.routines(id),
  CONSTRAINT shared_routines_owner_user_id_fkey
    FOREIGN KEY (owner_user_id) REFERENCES public.users(id),
  CONSTRAINT shared_routines_shared_with_user_id_fkey
    FOREIGN KEY (shared_with_user_id) REFERENCES public.users(id)
);

CREATE TABLE public.support_tickets (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL,
  title VARCHAR NOT NULL,
  description TEXT NOT NULL,
  status public.ticket_status NOT NULL DEFAULT 'open',
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  updated_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT support_tickets_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id)
);

CREATE TABLE public.workout_sessions (
  id BIGSERIAL PRIMARY KEY,
  user_id UUID NOT NULL,
  routine_id BIGINT,
  name VARCHAR,
  started_at TIMESTAMP NOT NULL,
  ended_at TIMESTAMP,
  duration_minutes INTEGER CHECK (duration_minutes >= 0),
  calories_burned NUMERIC CHECK (calories_burned >= 0),
  notes TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT workout_sessions_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES public.users(id),
  CONSTRAINT workout_sessions_routine_id_fkey
    FOREIGN KEY (routine_id) REFERENCES public.routines(id)
);

CREATE TABLE public.workout_exercises (
  id BIGSERIAL PRIMARY KEY,
  workout_session_id BIGINT NOT NULL,
  exercise_id BIGINT NOT NULL,
  exercise_order INTEGER NOT NULL CHECK (exercise_order > 0),
  notes TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT workout_exercises_workout_session_id_fkey
    FOREIGN KEY (workout_session_id) REFERENCES public.workout_sessions(id),
  CONSTRAINT workout_exercises_exercise_id_fkey
    FOREIGN KEY (exercise_id) REFERENCES public.exercises(id)
);

CREATE TABLE public.workout_sets (
  id BIGSERIAL PRIMARY KEY,
  workout_exercise_id BIGINT NOT NULL,
  set_number INTEGER NOT NULL CHECK (set_number > 0),
  reps INTEGER CHECK (reps >= 0),
  weight_kg NUMERIC CHECK (weight_kg >= 0),
  duration_seconds INTEGER CHECK (duration_seconds >= 0),
  distance_km NUMERIC CHECK (distance_km >= 0),
  rir INTEGER CHECK (rir >= 0 AND rir <= 10),
  completed BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP NOT NULL DEFAULT now(),
  CONSTRAINT workout_sets_workout_exercise_id_fkey
    FOREIGN KEY (workout_exercise_id) REFERENCES public.workout_exercises(id)
);
