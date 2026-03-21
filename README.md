# Monorepo Template: Go + React/Vite

A monorepo template for full-stack applications with a **Go** backend and a **React + TypeScript + Vite** frontend.
 
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
└── Makefile              Dev commands
```

## Prerequisites

- [Go](https://go.dev/dl/) 1.24+
- [Node.js](https://nodejs.org/) 22+

## Getting Started

```bash
make install

# Terminal 1
make run-backend    # port 8080

# Terminal 2
make run-frontend   # port 5173
```

The Vite dev server proxies `/api` requests to the backend.

## Commands

| Command              | Description                     |
|----------------------|---------------------------------|
| `make install`       | Install all dependencies        |
| `make run-backend`   | Backend with hot reload (Air)   |
| `make run-frontend`  | Frontend dev server (Vite)      |
| `make test`          | Run all tests                   |
| `make lint`          | Run all linters                 |
| `make e2e`           | Run Playwright E2E tests        |

## API

| Method | Path         | Description    |
|--------|------------- |----------------|
| `GET`  | `/health`    | Health check   |
| `GET`  | `/api/hello` | Sample endpoint|


# App de Seguimiento de Gimnasio

## Descripción
Aplicación web de escritorio diseñada para usuarios que desean llevar un control avanzado de sus entrenamientos, progresos y estadísticas físicas. La plataforma permite interactuar con amigos, recibir recomendaciones inteligentes mediante un chatbot de IA y gestionar rutinas de forma personalizada. Combina gestión de entrenamientos, análisis de datos, gamificación y asistencia mediante inteligencia artificial.

## Funcionalidades

### Vista Principal
- **Calendario de entrenamientos:** visualiza los días entrenados y los planes futuros.
- **Historial de entrenamientos:** registro detallado de días con actividad.
- **Octógono de estadísticas:** representa métricas por tipo de ejercicio (cardio, pecho, espalda, pierna, etc.).
- **Heatmap de músculos entrenados:** visualización del esfuerzo por grupo muscular.
- **Racha de entrenamientos:** seguimiento continuo al estilo gamificación.
- **Nuevo Entrenamiento:** acceso directo a la vista de rutinas para iniciar un entrenamiento.

### Vista de Rutinas
- **Rutinas predeterminadas:** secciones estándar (pecho, espalda, pierna).
- **Rutinas personalizadas:** crear, editar y nombrar nuevas rutinas según preferencias.
- **Creación de rutinas mediante IA:** generación automática de rutinas según perfil y objetivos.
- **Mejora de rutinas:** optimización de rutinas existentes mediante IA.

### Funciones
- **La aplicación podra predecir la repetición maxima que puede realizar en un ejercicio
- **La aplicación mostrara indicadores de fatiga según el rendimiento en los ejercicios
- **La aplicación mostrara un contador inteligente según el tipo de ejercicio
- **La aplicación mostrara graficas con la sobrecarga progesiva con respecto a los datos registrados
- ** Estimación de calorias consumidas durante el entreno
- ** Registro de duraciones de entreno
- ** Boton 
### Vista de Entrenamiento Actual
- **Barra de progreso:** indica el porcentaje de la sesión completada.
- **Lista de ejercicios:** opción a marcar series completadas y consultar estadísticas previas.
- **Gestión de repeticiones y peso:** añadir y modificar datos durante el entrenamiento.

### Funcionalidades Sociales
- **Estadísticas de amigos:** comparar progresos y rendimiento con otros usuarios.
- **Sistema de amistad y compartir rutinas:** códigos de solicitud de amistad y de rutinas.
- **Acceso restringido:** solo se pueden ver estadísticas de amigos y rutinas compartidas.

### Chatbot de IA
- **Recomendaciones personalizadas:** análisis del historial de entrenamientos y estadísticas para sugerir mejoras.
- **Generación de rutinas:** creación automática de rutinas según perfil y objetivos (pérdida de grasa, ganancia muscular, etc.).

### Rol Administrador
- **Gestión de usuarios:** creación, modificación y eliminación de cuentas.
- **Tickets de soporte:** gestión de solicitudes o incidencias de usuarios relacionadas con ejercicios oficiales.

## Tecnologías Utilizadas
- **Backend:** Go (Golang)
- **Frontend:** Vue.js
- **Persistencia de datos:** *(por definir)*
- **Inteligencia Artificial:** Chatbot y generación de rutinas basadas en historial y objetivos.
