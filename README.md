# Di Gröne

A web application for tracking floorball game statistics at both player and team levels.

## Technologies

This project is built with:

- Vite
- TypeScript
- React
- shadcn-ui
- Tailwind CSS
- React Router
- TanStack Query
- Recharts (for data visualization)

## Getting Started

### Prerequisites

- Node.js (v18 or higher recommended)
- npm

### Installation

```sh
# Enter the frontend application
cd frontend

# Install dependencies
npm ci

# Start the development server
npm run dev
```

The application will be available at `http://localhost:8080`

### Available Scripts

Run these commands from `frontend/`:

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run build:dev` - Build in development mode
- `npm run lint` - Run ESLint
- `npm run preview` - Preview production build

## Project Structure

- `/frontend` - React application, package files, and frontend build configuration
- `/frontend/src` - Frontend source code
- `/frontend/public` - Static assets and current JSON data
- `/frontend/src/components` - React components
- `/frontend/src/pages` - Page components
- `/backend` - Go API, PostgreSQL configuration, and migrations; see [backend setup](backend/README.md)
- `/.github` - Repository workflows

Shared repository files, including this README and `.gitignore`, remain at the root.
The frontend and backend will manage their dependencies and configuration separately.

## Development

This application uses React Router for navigation and TanStack Query for data management. The UI is built with shadcn-ui components and styled with Tailwind CSS.

GitHub Pages builds the frontend from `frontend/` and deploys `frontend/dist/`.
The Go backend foundation is available; see [backend setup](backend/README.md).
The homepage and statistics page read from the Go API. The admin and cup pages
continue using the existing JSON/GitHub storage until their integration is complete.
