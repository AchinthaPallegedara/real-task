# Google OAuth Integration Examples

This document provides examples of how to integrate Google OAuth with the Task Manager API.

## 🔧 Setup

1. Configure your `.env` file with Google OAuth credentials:

```env
GOOGLE_CLIENT_ID="your_google_client_id.apps.googleusercontent.com"
GOOGLE_CLIENT_SECRET="your_google_client_secret"
GOOGLE_REDIRECT_URL="http://localhost:8080/api/auth/google/callback"
```

2. Start the server:

```bash
go run cmd/api/main.go
```

## 🌐 Frontend Integration

### Method 1: Redirect Flow (Traditional Web App)

```javascript
// Get Google OAuth URL
const response = await fetch("/api/auth/google/url");
const { auth_url, state } = await response.json();

// Store state for verification
localStorage.setItem("oauth_state", state);

// Redirect user to Google
window.location.href = auth_url;

// After Google redirects back, the server will handle the callback
// and return tokens in the response
```

### Method 2: Popup Flow (SPA)

```javascript
// Frontend JavaScript example using Google's Sign-In library
function initGoogleSignIn() {
  google.accounts.id.initialize({
    client_id: "your_google_client_id.apps.googleusercontent.com",
    callback: handleGoogleSignIn,
  });

  google.accounts.id.renderButton(
    document.getElementById("google-signin-button"),
    { theme: "outline", size: "large" }
  );
}

async function handleGoogleSignIn(response) {
  try {
    // Send the authorization code to your backend
    const apiResponse = await fetch("/api/auth/google", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        code: response.code,
        state: "your_state_parameter",
      }),
    });

    const data = await apiResponse.json();

    if (apiResponse.ok) {
      // Store tokens
      localStorage.setItem("access_token", data.access_token);
      localStorage.setItem("refresh_token", data.refresh_token);

      // Redirect to dashboard
      window.location.href = "/dashboard";
    } else {
      console.error("Login failed:", data.error);
    }
  } catch (error) {
    console.error("OAuth error:", error);
  }
}
```

### Method 3: React Example

```jsx
import { useState } from "react";

function GoogleLoginButton() {
  const [loading, setLoading] = useState(false);

  const handleGoogleLogin = async () => {
    setLoading(true);
    try {
      // Get Google OAuth URL
      const response = await fetch("/api/auth/google/url");
      const { auth_url } = await response.json();

      // Open popup window
      const popup = window.open(
        auth_url,
        "google-oauth",
        "width=500,height=600,scrollbars=yes,resizable=yes"
      );

      // Listen for popup to close or send message
      const checkClosed = setInterval(() => {
        if (popup.closed) {
          clearInterval(checkClosed);
          setLoading(false);
          // Check if login was successful
          window.location.reload();
        }
      }, 1000);
    } catch (error) {
      console.error("Google login failed:", error);
      setLoading(false);
    }
  };

  return (
    <button
      onClick={handleGoogleLogin}
      disabled={loading}
      className="google-login-btn">
      {loading ? "Signing in..." : "Sign in with Google"}
    </button>
  );
}
```

## 🧪 Testing with cURL

### Get OAuth URL

```bash
curl -X GET http://localhost:8080/api/auth/google/url
```

Response:

```json
{
  "auth_url": "https://accounts.google.com/o/oauth2/auth?access_type=offline&client_id=...&response_type=code&scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fuserinfo.email+https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fuserinfo.profile&state=random_state",
  "state": "random_state_string"
}
```

### Check Available Providers

```bash
curl -X GET http://localhost:8080/api/auth/providers
```

Response:

```json
{
  "providers": [
    {
      "name": "Google",
      "id": "google",
      "endpoint": "/api/auth/google",
      "enabled": true
    }
  ]
}
```

### Login with Authorization Code (after user consent)

```bash
curl -X POST http://localhost:8080/api/auth/google \
  -H "Content-Type: application/json" \
  -d '{
    "code": "authorization_code_from_google",
    "state": "state_parameter"
  }'
```

Response:

```json
{
  "message": "Login successful",
  "access_token": "jwt_token_here",
  "refresh_token": "refresh_token_here",
  "expires_at": "2023-12-07T11:30:00Z",
  "token_type": "Bearer",
  "user": {
    "id": 1,
    "email": "user@gmail.com",
    "name": "User Name",
    "email_verified": true,
    "two_factor_enabled": false,
    "provider": "google"
  }
}
```

## 🔒 Security Notes

1. **State Parameter**: Always verify the state parameter to prevent CSRF attacks
2. **HTTPS**: Use HTTPS in production for OAuth redirects
3. **Token Storage**: Store tokens securely (HttpOnly cookies recommended)
4. **Scope Limitation**: Only request necessary OAuth scopes
5. **Token Expiration**: Handle token refresh properly

## 🚫 Troubleshooting

### Common Issues

1. **"Google OAuth not configured" error**

   - Check your `.env` file has correct Google credentials
   - Ensure `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` are set

2. **"Invalid redirect URI" error**

   - Verify the redirect URI in Google Console matches `GOOGLE_REDIRECT_URL`
   - For development, use `http://localhost:8080/api/auth/google/callback`

3. **"Invalid client ID" error**

   - Double-check your Google Client ID
   - Ensure it ends with `.apps.googleusercontent.com`

4. **CORS issues in browser**
   - The API includes CORS middleware
   - For production, configure proper allowed origins

## 📱 Mobile App Integration

For mobile apps, you can still use the same endpoints:

1. Use the native Google Sign-In SDK for your platform
2. Get the authorization code from the native SDK
3. Send the code to `/api/auth/google` endpoint
4. Receive JWT tokens for API authentication

This approach works for React Native, Flutter, native iOS/Android apps, etc.
