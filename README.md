# GitHub OAuth Login Example

A simple Go web application demonstrating GitHub OAuth authentication.

## Setup

1. **Create a GitHub OAuth App:**
   - Go to https://github.com/settings/developers
   - Click "New OAuth App"
   - Set Application name: "GitHub Login Example"
   - Set Homepage URL: `http://localhost:8080`
   - Set Authorization callback URL: `http://localhost:8080/callback`
   - Copy the Client ID and Client Secret

2. **Set Environment Variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your actual GitHub OAuth credentials
   ```

3. **Install Dependencies:**
   ```bash
   go mod tidy
   ```

4. **Run the Application:**
   ```bash
   go run main.go
   ```

5. **Open Browser:**
   Visit `http://localhost:8080`

## Features

- GitHub OAuth login/logout
- User profile display
- Session management
- Responsive HTML templates

## Routes

- `/` - Home page
- `/login` - Initiate GitHub OAuth
- `/callback` - OAuth callback handler
- `/profile` - User profile page
- `/logout` - Logout and clear session