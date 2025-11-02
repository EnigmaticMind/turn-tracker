# Deployment Guide

## Backend Deployment (fly.io)

### Prerequisites

- [flyctl CLI](https://fly.io/docs/getting-started/installing-flyctl/) installed
- fly.io account created

### Steps

1. **Navigate to backend directory:**

   ```bash
   cd backend
   ```

2. **Login to fly.io:**

   ```bash
   fly auth login
   ```

3. **Initialize fly.io app (first time only):**

   ```bash
   fly launch
   ```

   - Choose a unique app name (or it will generate one)
   - Select a region close to your users
   - Don't deploy yet if prompted

4. **Update fly.toml:**

   - Edit `fly.toml` and set `app = "your-app-name"` (if needed)
   - The configuration is already set up for WebSocket support

5. **Deploy:**

   ```bash
   fly deploy
   ```

6. **Get your app URL:**

   ```bash
   fly status
   ```

   Your app will be available at `https://your-app-name.fly.dev`

7. **View logs:**
   ```bash
   fly logs
   ```

### WebSocket Connection

The backend WebSocket endpoint will be available at:

- Production: `wss://your-app-name.fly.dev/ws`
- The fly.toml is configured to handle WebSocket upgrades automatically

## Frontend Deployment (GitHub Pages)

### Prerequisites

- GitHub repository set up
- GitHub Actions enabled

### Configuration

1. **Update WebSocket URL:**

   - Edit `frontend/src/lib/websocket/WebSocketConnection.ts`
   - Update the production URL to match your fly.io app:
     ```typescript
     const WS_URL =
       import.meta.env.VITE_WS_URL ||
       (import.meta.env.PROD
         ? "wss://your-app-name.fly.dev/ws" // Update this!
         : "ws://localhost:8080/ws");
     ```

2. **Set environment variable for GitHub Pages (optional):**

   - In GitHub repository settings, add a secret `VITE_WS_URL` if you want to override
   - Otherwise, the code will use the hardcoded production URL

3. **Create GitHub Actions workflow** (if not already set up):
   Create `.github/workflows/deploy.yml`:

   ```yaml
   name: Deploy to GitHub Pages

   on:
     push:
       branches: [main]
     workflow_dispatch:

   permissions:
     contents: read
     pages: write
     id-token: write

   jobs:
     deploy:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v4

         - name: Setup Node.js
           uses: actions/setup-node@v4
           with:
             node-version: "20"
             cache: "npm"
             cache-dependency-path: frontend/package-lock.json

         - name: Install dependencies
           working-directory: frontend
           run: npm ci

         - name: Build
           working-directory: frontend
           run: npm run build
           env:
             VITE_WS_URL: ${{ secrets.VITE_WS_URL || 'wss://your-app-name.fly.dev/ws' }}

         - name: Setup Pages
           uses: actions/configure-pages@v4

         - name: Upload artifact
           uses: actions/upload-pages-artifact@v3
           with:
             path: frontend/build/client

         - name: Deploy to GitHub Pages
           id: deployment
           uses: actions/deploy-pages@v4
   ```

## Local Development

### Backend

```bash
cd backend
go run .
```

Backend will run on `http://localhost:8080`

### Frontend

```bash
cd frontend
npm install
npm run dev
```

### Environment Variables

- **Frontend**: Create `frontend/.env.local` with `VITE_WS_URL=ws://localhost:8080/ws` if needed
- **Backend**: Uses `PORT` environment variable (defaults to 8080)

## Testing Production Locally

To test the production build locally:

```bash
cd frontend
npm run build
npm run preview
```

This will use the production WebSocket URL configuration.
