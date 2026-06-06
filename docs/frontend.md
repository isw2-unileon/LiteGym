# Frontend guide

The frontend is a React 19 app built with TypeScript and Vite. It uses route-based navigation, cookie-backed authenticated requests, and page-specific components for each domain area.

## Frontend structure

```text
frontend/
|-- src/
|   |-- App.tsx
|   |-- main.tsx
|   |-- lib/
|   |-- pages/
|   |-- components/
|   `-- types/
|-- vite.config.ts
`-- package.json
```

## Main routing

Routes are defined in:

- `frontend/src/App.tsx`

Current route map:

- `/` -> login
- `/dashboard` -> dashboard page
- `/exercises` -> exercise management and insights
- `/profile` -> user profile page
- `/routines` -> routine library and AI routine flow
- `/admin` -> admin page
- `/support` -> support page

Any unknown route falls back through `NotFoundRedirect`.

## Authenticated application shell

Two components shape the authenticated experience:

- `frontend/src/components/AuthenticatedLayoutRoute.tsx`
- `frontend/src/components/AppLayout.tsx`

### `AuthenticatedLayoutRoute`

Responsibilities:

- calls `/api/auth/me`
- blocks anonymous access
- renders a session check screen while auth state is loading
- mounts the main application layout when the session is valid

### `AppLayout`

Responsibilities:

- sidebar and navigation shell
- logout action
- top-level page background and spacing
- admin-specific navigation link when user role is `admin`

## Main pages

### `DashboardPage`

Shows overview information such as summary cards and training-related insights.

### `ExercisePage`

Handles:

- exercise listing
- filters and pagination
- create and edit exercise flows
- workout session history for an exercise
- exercise insight visualization

### `UserRoutinesPage`

Handles:

- routine listing
- routine detail display
- AI routine generation form
- AI preview modal
- AI save confirmation flow

Recent UX improvements in this page include:

- selectable mandatory exercises by name instead of raw ids
- free-text notes for the AI
- scrollable and responsive preview modal

### `AdminPage`

Combines admin-focused views such as:

- user administration
- ticket administration
- exercise administration

### `SupportPage`

Used for support ticket interactions.

## Data access pattern

The frontend mostly uses direct `fetch(...)` calls rather than a centralized API client layer. URLs are built with:

- `frontend/src/lib/api.ts`

If `VITE_API_BASE_URL` is not set, requests stay relative and rely on the Vite proxy in development.

## Type organization

Shared frontend domain types currently exist mainly in:

- `frontend/src/types/exercise.ts`

Some pages also declare page-local response types when the type is tightly bound to one page.

## Development server behavior

`frontend/vite.config.ts` configures local development behavior, including:

- host binding
- backend proxying for `/api`
- backend proxying for `/health`
- timeout tuning useful for slower AI requests

## Styling approach

The frontend uses a utility-class-heavy approach with a custom visual language rather than browser defaults. Common patterns include:

- rounded high-radius cards
- warm neutral backgrounds with green/orange accent colors
- split-panel modals
- responsive stacking with `sm`, `md`, `lg`, and `xl` breakpoints

## Frontend extension guidelines

When adding or changing a feature:

1. identify whether the page already owns the data contract
2. keep route-level state in the page and visual subparts in components
3. preserve the established visual language
4. add a Vitest page/component test when behavior changes
5. run `npm run build` after structural UI changes
