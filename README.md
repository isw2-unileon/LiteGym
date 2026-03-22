# Gym Tracking App

## Overview

This project is a desktop-oriented web application designed for users who want to track and analyze their workouts in a detailed and intelligent way. It combines workout management, performance analytics, social interaction, and AI-driven recommendations into a single platform.

The application allows users to monitor their progress, manage training routines, compare results with friends, and receive personalized suggestions through an integrated AI assistant.

---

## Key Features

### Main Dashboard

- **Workout Calendar**  
  Visual representation of past and scheduled training sessions.

- **Workout History**  
  Detailed log of all completed workouts.

- **Statistics Radar Chart**  
  Displays performance metrics across different muscle groups such as chest, back, legs, cardio, and more.

- **Muscle Heatmap**  
  Visualizes muscle engagement and training intensity over time.

- **Workout Streak System**  
  Tracks consistency and encourages habit-building through gamification.

- **Quick Start Training Button**  
  Fast access to begin a new workout session.

---

### Routine Management

- **Predefined Routines**  
  Standard workout splits such as chest, back, and legs.

- **Custom Routines**  
  Users can create, edit, and manage personalized workout plans.

- **AI-Generated Routines**  
  Automatic routine creation based on the user’s profile, fitness level, and goals.

- **Routine Improvement Suggestions**  
  Existing routines can be analyzed and optimized with AI-generated recommendations.

---

### Performance and Analytics

- **One-Rep Max Prediction**  
  Estimate the maximum weight a user can lift for a single repetition in a given exercise.

- **Fatigue Indicators**  
  Detect signs of fatigue by analyzing workout performance and progression.

- **Smart Exercise Counter**  
  Intelligent counting system adapted to the type of exercise being performed.

- **Progressive Overload Charts**  
  Visual representation of progress over time based on recorded workout data.

- **Calorie Burn Estimation**  
  Approximate the calories burned during a workout session.

- **Workout Duration Tracking**  
  Record and analyze the duration of each training session.

---

### Active Workout View

- **Session Progress Bar**  
  Shows the percentage of the current workout that has been completed.

- **Exercise List**  
  Displays the exercises included in the current routine, with the ability to mark completed sets and view previous statistics.

- **Weight and Repetition Logging**  
  Users can add, edit, and update repetitions and weights in real time during the workout.

---

### Social Features

- **Friends Statistics**  
  Compare performance, progress, and workout consistency with friends.

- **Friend System and Routine Sharing**  
  Add friends and share routines through unique invitation or sharing codes.

- **Restricted Access**  
  Only friends can access shared routines and personal performance statistics.

---

### AI Chatbot

- **Personalized Recommendations**  
  Analyze workout history, physical progress, and performance data to provide customized advice.

- **Routine Generation**  
  Create workout plans based on specific goals such as fat loss, muscle gain, or performance improvement.

---

### Administrator Role

- **User Management**  
  Create, modify, and delete user accounts.

- **Support Ticket Management**  
  Handle support requests and user issues related to official exercises or platform usage.

---

## Tech Stack

- **Backend:** Go (Golang)
- **Frontend:** React + TypeScript + Vite
- **Styling:** Tailwind CSS
- **Testing:** Playwright for end-to-end testing
- **Artificial Intelligence:** AI chatbot for recommendations and routine generation
- **Data Persistence:** To be defined

---

## Project Structure

```text
├── backend/              Go API server (Gin)
│   ├── cmd/server/       Entry point
│   └── internal/config/  Environment config
│
├── frontend/             React + TypeScript + Vite + Tailwind
│   └── src/
│
├── e2e/                  Playwright E2E tests
├── .github/workflows/    CI/CD pipelines
└── Makefile              Development commands
