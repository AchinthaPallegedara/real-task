# 🛡️ Real-Time Collaborative Task Management API

A robust, real-time collaborative task management system built with Go, featuring **modern authentication with 2FA**, email verification, JWT authentication, WebSocket real-time updates, background notifications, and comprehensive task management capabilities.

## ⭐ Key Features

- **🔐 Modern Authentication System** - Complete auth flow with email verification, 2FA (TOTP), password reset
- **📧 Email Gateway Integration** - Automated email notifications for verification, password reset, and security alerts
- **🛡️ Two-Factor Authentication** - TOTP-based 2FA with backup codes and QR code setup
- **🔒 Advanced Security** - Account lockout, login attempt tracking, secure token generation
- **👥 Project Collaboration** - Multi-user project management with invitation system
- **📋 Task Management** - Complete CRUD operations for tasks with assignment capabilities
- **⚡ Real-Time Updates** - WebSocket-powered live collaboration with typing indicators
- **🔔 Background Notifications** - Asynchronous email notifications for task events
- **👁️ User Presence** - Online/offline status tracking for collaborators
- **🌐 Cross-User Assignment** - Assign tasks to users not yet in the project (auto-invite)
- **📊 Real-Time Analytics** - Live system statistics and connected user tracking

## 🏗️ Architecture Overview

This API follows clean architecture principles with real-time capabilities:

```
├── cmd/api/                 # Application entry point
├── internal/
│   ├── database/           # Database connection and configuration
│   ├── handlers/           # HTTP request handlers (controllers)
│   ├── middleware/         # HTTP middleware (JWT auth)
│   ├── models/             # Database models and structures
│   ├── services/           # Business logic layer
│   ├── utils/              # Utility functions (JWT, password hashing)
│   ├── websocket/          # Real-time WebSocket hub and client management
│   └── notifications/      # Background notification service
└── test-realtime-features.sh   # Comprehensive testing script
```

## 🔐 Authentication with HttpOnly Cookies

This API uses HttpOnly cookies for secure authentication, providing better protection against XSS attacks compared to storing tokens in localStorage.

### How it works

1. **Login Process**:

   - Upon successful login, both access_token and refresh_token are sent as HttpOnly cookies
   - Access token expires after 1 hour
   - Refresh token expires after 7 days
   - Tokens are also returned in the response body for backward compatibility

2. **Authenticated Requests**:

   - The API automatically checks for the access_token cookie in requests
   - No need to manually send Authorization headers when cookies are enabled
   - Browsers automatically include cookies with each request to the domain

3. **Token Refresh**:

   - When the access token expires, call the refresh endpoint
   - The API checks for the refresh_token cookie
   - New tokens are set as cookies automatically

4. **Logout**:
   - The logout endpoint clears both cookies by setting them to expire immediately
   - The server also invalidates the refresh token in the database

### Frontend Integration

```javascript
// Example of login with credentials (using axios)
const login = async (email, password) => {
  try {
    const response = await axios.post(
      "/login",
      { email, password },
      {
        withCredentials: true, // Important to enable cookies
      }
    );
    return response.data;
  } catch (error) {
    throw error;
  }
};

// For authenticated requests, just include withCredentials
const getProfile = async () => {
  try {
    const response = await axios.get("/api/profile", {
      withCredentials: true,
    });
    return response.data;
  } catch (error) {
    throw error;
  }
};

// Logout - the server will clear the cookies
const logout = async () => {
  try {
    await axios.post(
      "/api/auth/logout",
      {},
      {
        withCredentials: true,
      }
    );
  } catch (error) {
    throw error;
  }
};
```

### CORS Configuration

For cross-domain requests, ensure your frontend application is included in the allowed origins and credentials are enabled:

```go
// CORSMiddleware configuration already includes:
c.Header("Access-Control-Allow-Credentials", "true")
```

### Using HttpOnly Cookies with Postman

When testing the API with Postman, you'll need to configure it to handle HttpOnly cookies properly:

1. **Enable Cookie Management in Postman**:

   - Open Postman settings (click the gear icon)
   - Go to the "General" tab
   - Ensure "Automatically follow redirects" is enabled
   - Make sure "Save cookies with requests" is enabled

2. **Configure a Postman Collection**:

   - Create a new collection for your API
   - Go to the collection settings (click the "..." next to your collection name)
   - Select the "Settings" tab
   - Under "Settings", enable "Automatically store cookies"

3. **Testing the Authentication Flow**:

   - **Login**: When you make a successful login request, Postman will automatically store the HttpOnly cookies
   - **Authenticated Requests**: Postman will automatically include the cookies in subsequent requests to the same domain
   - **Refresh**: When your access token expires, call the refresh endpoint to get new cookies
   - **Logout**: The logout endpoint will clear the cookies

4. **Example Login Request**:

   ```
   POST http://localhost:8080/login
   Content-Type: application/json

   {
     "email": "user@example.com",
     "password": "password123"
   }
   ```

   After successful login, you can check the cookies by:

   - Going to the "Cookies" button in the footer of Postman
   - Selecting your domain
   - You should see `access_token` and `refresh_token` cookies listed

5. **Making Authenticated Requests**:

   - Simply make your requests to protected endpoints
   - Postman will automatically include the cookies
   - No need to manually set Authorization headers

6. **Troubleshooting**:
   - If you're getting authentication errors, check if cookies are being stored properly
   - You can manually inspect and manage cookies via Postman's Cookie Manager
   - Try clearing cookies and logging in again if you experience issues

This cookie-based approach simplifies testing as Postman handles the authentication state automatically between requests.

## 🗄️ Database Schema

### Core Models

**User Model:**

- `ID` (Primary Key)
- `Email` (Unique, required)
- `Password` (Hashed, bcrypt)
- `Name` (Optional display name)
- `EmailVerified` (Email verification status)
- `EmailVerificationToken` (Temporary verification token)
- `EmailVerificationExpiry` (Token expiration time)
- `TwoFactorEnabled` (2FA enabled status)
- `TwoFactorSecret` (TOTP secret key, encrypted)
- `TwoFactorBackupCodes` (Recovery codes for 2FA)
- `PasswordResetToken` (Temporary reset token)
- `PasswordResetExpiry` (Reset token expiration)
- `LastLoginAt` (Last successful login timestamp)
- `LoginAttempts` (Failed login attempt counter)
- `LockedUntil` (Account lockout expiration)
- `IsActive` (Account status)
- Relationships: OwnedProjects, Collaborations, AssignedTasks

**Project Model:**

- `ID` (Primary Key)
- `Name` (Required)
- `Description`
- `OwnerID` (Foreign Key to User)
- Relationships: Owner, Tasks, Collaborators

**Task Model:**

- `ID` (Primary Key)
- `Title` (Required)
- `Description`
- `Status` (pending, in_progress, completed)
- `ProjectID` (Foreign Key to Project)
- `AssigneeID` (Optional Foreign Key to User)

### Collaboration Models

**ProjectCollaborator Model:**

- `ID` (Primary Key)
- `ProjectID` (Foreign Key)
- `UserID` (Foreign Key)
- `Role` (owner, collaborator, viewer)
- `JoinedAt` (Timestamp)

**ProjectInvitation Model:**

- `ID` (Primary Key)
- `ProjectID` (Foreign Key)
- `UserID` (Foreign Key - invited user)
- `InviterID` (Foreign Key - user who sent invitation)
- `Status` (pending, accepted, declined)
- `Message` (Optional invitation message)

## 🚀 Getting Started

### Prerequisites

- Go 1.24.3 or higher
- PostgreSQL database
- Environment variables configured

### Environment Setup

Create a `.env` file with:

```env
# Database
DB_URL="your_postgresql_connection_string"

# JWT
JWT_SECRET="your_super_secret_jwt_key"

# Server
PORT="8080"
BASE_URL="http://localhost:8080"

# Email Configuration (for production)
SMTP_HOST="smtp.gmail.com"
SMTP_PORT="587"
SMTP_USERNAME="your_email@gmail.com"
SMTP_PASSWORD="your_app_password"
FROM_EMAIL="noreply@yourdomain.com"

# Google OAuth (Optional)
GOOGLE_CLIENT_ID="your_google_client_id.apps.googleusercontent.com"
GOOGLE_CLIENT_SECRET="your_google_client_secret"
GOOGLE_REDIRECT_URL="http://localhost:8080/api/auth/google/callback"
```

**Note**: For development, email sending is simulated in console logs. For production, configure SMTP settings to enable real email delivery.

#### Google OAuth Setup (Optional)

To enable Google authentication:

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google+ API or People API
4. Create OAuth 2.0 credentials (Web application)
5. Add authorized redirect URI: `http://localhost:8080/api/auth/google/callback`
6. Copy the Client ID and Client Secret to your `.env` file

Without Google OAuth configuration, the `/api/auth/providers` endpoint will show Google as disabled.

### Installation & Running

```bash
# Clone and navigate to the project
cd real-task

# Install dependencies
go mod tidy

# Run the application
go run cmd/api/main.go
```

The server will start on `http://localhost:8080`

## 🎯 Quick Start Guide

### 1. Start the Server

```bash
# Make sure you have Go 1.21+ installed
go run cmd/api/main.go
```

You should see:

```
🚀 Real-time WebSocket system initialized!
📧 Background notification system initialized!
Server starting on :8080
```

### 2. Run the Test Script (Optional)

```bash
# Make the script executable and run it
chmod +x test-realtime-features.sh
./test-realtime-features.sh
```

This will:

- Create test users (Alice and Bob)
- Set up a collaborative project
- Create tasks and test real-time features
- Provide JWT tokens for WebSocket testing

### 3. Test Modern Authentication Features

#### Quick Registration and Login Flow

```bash
# 1. Register a new user
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"yourname@example.com","password":"securepass123"}'

# 2. Check console logs for verification email and copy the token
# Look for: "Email would be sent to yourname@example.com with subject: Verify Your Email Address"

# 3. Verify email (replace TOKEN with actual token from logs)
curl -X POST http://localhost:8080/api/auth/verify-email \
  -H "Content-Type: application/json" \
  -d '{"token":"YOUR_VERIFICATION_TOKEN_HERE"}'

# 4. Login successfully
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"yourname@example.com","password":"securepass123"}'

# Save the JWT token from the response for API testing
export TOKEN="your_jwt_token_here"
```

### 4. Test Two-Factor Authentication

```bash
# 1. Setup 2FA
curl -X POST http://localhost:8080/api/auth/2fa/setup \
  -H "Authorization: Bearer $TOKEN"

# 2. Use an authenticator app (Google Authenticator, Authy) to scan the QR code
# 3. Enter the 6-digit code to confirm (replace 123456 with actual code)
curl -X POST http://localhost:8080/api/auth/2fa/confirm \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"token":"123456"}'

# 4. Save the backup codes from the response!
```

### 5. Test Real-Time Features

#### Option A: Use Postman for WebSocket Testing

1. **Create a new WebSocket request in Postman**
2. **Set URL:** `ws://localhost:8080/api/ws`
3. **Add Authorization header:**
   - Key: `Authorization`
   - Value: `Bearer YOUR_JWT_TOKEN`
4. **Connect** and test real-time features

#### Option B: Use Browser Developer Console

```javascript
// For browser testing, you'll need to send the Authorization header
// This requires a more advanced setup with a proper WebSocket client
const token = "YOUR_JWT_TOKEN";
const ws = new WebSocket("ws://localhost:8080/api/ws");

// Note: Browser WebSocket connections don't support custom headers natively
// For production browser clients, consider using a WebSocket library that supports headers
```

#### Option B: Manual Testing with curl

```bash
# Get connected users for a project
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/projects/1/connected-users

# Get user presence
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/users/2/presence

# Get system stats
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/admin/realtime-stats
```

## ⚡ Real-Time Features

### 🌐 WebSocket Connection

Connect to the WebSocket endpoint for real-time updates using Authorization header:

```http
GET /api/ws
Authorization: Bearer YOUR_JWT_TOKEN
Upgrade: websocket
Connection: Upgrade
```

**Postman Example:**

```
URL: ws://localhost:8080/api/ws
Headers:
  Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Production Client Example:**

```javascript
// For production applications, use a WebSocket library that supports headers
// such as 'ws' library for Node.js or similar for other platforms
const WebSocket = require("ws");
const ws = new WebSocket("ws://localhost:8080/api/ws", {
  headers: {
    Authorization: "Bearer YOUR_JWT_TOKEN",
  },
});
```

### 📡 Real-Time Message Types

The WebSocket connection broadcasts these message types:

| Message Type             | Description                    | Data                               |
| ------------------------ | ------------------------------ | ---------------------------------- |
| `connection_established` | Welcome message when connected | User info and project list         |
| `task_created`           | New task created in project    | Task details and creator info      |
| `task_updated`           | Task updated by collaborator   | Task changes and updater info      |
| `task_deleted`           | Task removed from project      | Task ID and deleter info           |
| `task_assigned`          | Task assigned to user          | Assignment details                 |
| `user_joined_project`    | New collaborator joined        | User and project info              |
| `user_left_project`      | Collaborator left project      | User and project info              |
| `typing_indicator`       | User typing status             | User ID, project ID, typing status |
| `user_presence`          | Online/offline status          | User ID and presence status        |

### 🎮 Interactive Features

Send these messages from client to server:

```javascript
// Start typing indicator
ws.send(
  JSON.stringify({
    type: "typing_start",
    data: { project_id: 1 },
  })
);

// Stop typing indicator
ws.send(
  JSON.stringify({
    type: "typing_stop",
    data: { project_id: 1 },
  })
);

// Update presence status
ws.send(
  JSON.stringify({
    type: "presence_update",
    data: { status: "away" }, // 'online', 'away', 'offline'
  })
);
```

### 🔔 Background Notifications

The system automatically sends notifications for:

- **Task Assignment**: When a task is assigned to a user
- **Task Updates**: When someone updates a task you're assigned to
- **Task Completion**: When a task is marked as completed
- **Project Invitations**: When invited to join a project

_Note: Notifications are currently simulated in logs. In production, integrate with email services like SendGrid, AWS SES, etc._

### 🧪 Testing Real-Time Features

1. **Use Postman for WebSocket Testing**: Create WebSocket request with Authorization header
2. **Run the Test Script**: Execute `./test-realtime-features.sh` for automated testing and JWT tokens
3. **Production Client**: Use WebSocket libraries that support custom headers for real applications

## 📚 API Documentation

### 🔐 Modern Authentication System

This API implements a comprehensive authentication system with modern security features:

- **Email Verification**: Users must verify their email address before login
- **Two-Factor Authentication (2FA)**: TOTP-based 2FA with backup codes
- **Password Reset**: Secure password reset via email tokens
- **Account Security**: Login attempt tracking, account lockout, security alerts
- **Session Management**: JWT-based authentication with proper token handling

#### Authentication Flow

1. **Registration** → Email verification required
2. **Email Verification** → Account activated
3. **Login** → JWT token issued (2FA required if enabled)
4. **2FA Setup** (Optional) → QR code scan, backup codes generation
5. **Protected Operations** → JWT token validation

### Authentication Endpoints

### Authentication Endpoints

#### Register User

```http
POST /register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}

Response (201):
{
  "message": "User registered successfully. Please check your email for verification link.",
  "user_id": 1,
  "email_verified": false
}
```

#### Verify Email

```http
# Via GET request (email link)
GET /api/auth/verify-email?token=verification_token_here

# Via POST request
POST /api/auth/verify-email
Content-Type: application/json

{
  "token": "verification_token_here"
}

Response (200):
{
  "message": "Email verified successfully"
}
```

#### Resend Verification Email

```http
POST /api/auth/resend-verification
Content-Type: application/json

{
  "email": "user@example.com"
}

Response (200):
{
  "message": "Verification email sent successfully"
}
```

#### Login User

```http
POST /login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "two_factor_token": "123456"  // Required if 2FA is enabled
}

Response (200):
{
  "message": "Login successful",
  "token": "jwt_token_here",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "User Name",
    "email_verified": true,
    "two_factor_enabled": false
  }
}

# If 2FA required:
Response (401):
{
  "error": "Two-factor authentication required",
  "two_factor_required": true
}

# If email not verified:
Response (401):
{
  "error": "Email not verified. Please check your email for verification link.",
  "email_verified": false
}
```

#### Google OAuth Authentication

**Get Google OAuth URL**

```http
GET /api/auth/google/url

Response (200):
{
  "auth_url": "https://accounts.google.com/oauth/authorize?client_id=...",
  "state": "random_state_string"
}
```

**Google OAuth Callback (for web redirect)**

```http
GET /api/auth/google/callback?code=AUTH_CODE&state=STATE

Response (200):
{
  "message": "Login successful",
  "access_token": "jwt_access_token_here",
  "refresh_token": "jwt_refresh_token_here",
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

**Google OAuth Login (for client-side apps)**

```http
POST /api/auth/google
Content-Type: application/json

{
  "code": "authorization_code_from_google",
  "state": "state_parameter"
}

Response (200):
{
  "message": "Login successful",
  "access_token": "jwt_access_token_here",
  "refresh_token": "jwt_refresh_token_here",
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

**Get Available OAuth Providers**

```http
GET /api/auth/providers

Response (200):
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

#### Request Password Reset

```http
POST /api/auth/request-password-reset
Content-Type: application/json

{
  "email": "user@example.com"
}

Response (200):
{
  "message": "Password reset email sent if account exists"
}
```

#### Reset Password

```http
# Via GET request (email link) - shows reset form
GET /api/auth/reset-password?token=reset_token_here

# Via POST request - actually resets password
POST /api/auth/reset-password
Content-Type: application/json

{
  "token": "reset_token_here",
  "new_password": "newpassword123"
}

Response (200):
{
  "message": "Password reset successfully"
}
```

### 🛡️ Two-Factor Authentication Endpoints

#### Setup 2FA

```http
POST /api/auth/2fa/setup
Authorization: Bearer <token>

Response (200):
{
  "message": "Two-factor authentication setup initiated",
  "secret": "JBSWY3DPEHPK3PXP",
  "qr_code": "otpauth://totp/Task%20Manager:user@example.com?secret=JBSWY3DPEHPK3PXP&issuer=Task%20Manager",
  "instructions": "Scan the QR code with your authenticator app and enter the generated token to confirm setup"
}
```

#### Confirm 2FA Setup

```http
POST /api/auth/2fa/confirm
Authorization: Bearer <token>
Content-Type: application/json

{
  "token": "123456"  // 6-digit code from authenticator app
}

Response (200):
{
  "message": "Two-factor authentication enabled successfully",
  "backup_codes": ["ABC12345", "DEF67890", ...],
  "warning": "Save these backup codes in a secure location. Each code can only be used once."
}
```

#### Disable 2FA

```http
POST /api/auth/2fa/disable
Authorization: Bearer <token>
Content-Type: application/json

{
  "password": "current_password"
}

Response (200):
{
  "message": "Two-factor authentication disabled successfully"
}
```

#### Verify 2FA Token

```http
POST /api/auth/2fa/verify
Authorization: Bearer <token>
Content-Type: application/json

{
  "token": "123456"  // 6-digit code or backup code
}

Response (200):
{
  "message": "Two-factor authentication verified"
}
```

### 👤 User Account Management

#### Get Authentication Status

```http
GET /api/auth/status
Authorization: Bearer <token>

Response (200):
{
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "User Name",
    "email_verified": true,
    "two_factor_enabled": true,
    "is_active": true,
    "last_login_at": "2023-12-06T10:30:00Z"
  }
}
```

#### Change Password

```http
POST /api/auth/change-password
Authorization: Bearer <token>
Content-Type: application/json

{
  "current_password": "oldpassword123",
  "new_password": "newpassword123"
}

Response (200):
{
  "message": "Password changed successfully"
}
```

### Protected Endpoints

All endpoints below require the `Authorization: Bearer <token>` header.

### User Endpoints

#### Get User Profile

```http
GET /api/profile
Authorization: Bearer <token>
```

### Project Endpoints

#### Create Project

```http
POST /api/projects
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "My Awesome Project",
  "description": "Project description here"
}
```

#### Get All Projects

```http
GET /api/projects
Authorization: Bearer <token>

# Returns both owned and collaborated projects
```

#### Get Specific Project

```http
GET /api/projects/:project_id
Authorization: Bearer <token>

# Returns project details if user has access (owner or collaborator)
```

#### Update Project

```http
PUT /api/projects/:project_id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Updated Project Name",        # Optional
  "description": "Updated description"   # Optional
}

# Only project owners can update projects
# You can update just name, just description, or both
```

#### Delete Project

```http
DELETE /api/projects/:project_id
Authorization: Bearer <token>

# Only project owners can delete projects
# Deletes all related data: tasks, collaborators, invitations
Response:
{
  "message": "Project deleted successfully"
}
```

### Task Endpoints

#### Create Task in Project

```http
POST /api/projects/:project_id/tasks
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Implement user authentication",
  "description": "Add JWT-based authentication system",
  "assignee_id": 2  // Optional: assign to any user (auto-invites if not a collaborator)
}

# 🔥 NEW FEATURE: Auto-Invitation System
# If the assignee is not a collaborator on the project, they will automatically
# receive an invitation to join the project when assigned a task!
```

#### Get All Tasks in Project

```http
GET /api/projects/:project_id/tasks
Authorization: Bearer <token>
```

#### Get Specific Task

```http
GET /api/tasks/:task_id
Authorization: Bearer <token>
```

#### Update Task

```http
PUT /api/tasks/:task_id
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Updated task title",           // Optional
  "description": "Updated description",    // Optional
  "status": "in_progress",                // Optional: pending, in_progress, completed
  "assignee_id": 3                        // Optional: reassign task
}
```

#### Delete Task

```http
DELETE /api/tasks/:task_id
Authorization: Bearer <token>
```

### Collaboration Endpoints

#### Invite User to Project

```http
POST /api/projects/:project_id/invite
Authorization: Bearer <token>
Content-Type: application/json

{
  "user_email": "collaborator@example.com",
  "message": "Would you like to collaborate on this project?"  // Optional
}

# Only project owners can send invitations
```

#### Get Pending Invitations

```http
GET /api/invitations
Authorization: Bearer <token>

# Returns all pending invitations for the authenticated user
```

#### Respond to Invitation

```http
POST /api/invitations/:invitation_id/respond
Authorization: Bearer <token>
Content-Type: application/json

{
  "accept": true  // true to accept, false to decline
}
```

#### Get Project Collaborators

```http
GET /api/projects/:project_id/collaborators
Authorization: Bearer <token>

# Returns all collaborators for the project
# Only accessible to project owners and collaborators
```

#### Remove Collaborator

```http
DELETE /api/projects/:project_id/collaborators/:user_id
Authorization: Bearer <token>

# Only project owners can remove collaborators
# Cannot remove the project owner
```

### Real-Time Endpoints

#### WebSocket Connection

```http
GET /api/ws
Authorization: Bearer <token>
Upgrade: websocket
Connection: Upgrade

# Establishes WebSocket connection for real-time updates
# Requires JWT authentication via Authorization header
```

#### Get Connected Users

```http
GET /api/projects/:project_id/connected-users
Authorization: Bearer <token>

# Returns list of currently connected users for the project
Response:
{
  "connected_users": [
    {
      "user_id": 1,
      "status": "online",
      "last_seen": "2023-12-06T10:30:00Z"
    }
  ],
  "total_connected": 1
}
```

#### Get User Presence

```http
GET /api/users/:user_id/presence
Authorization: Bearer <token>

# Returns presence information for a specific user
# Only works if users share at least one project
Response:
{
  "user_id": 2,
  "status": "online",  // "online", "offline", "away"
  "last_seen": "2023-12-06T10:30:00Z"
}
```

#### Get Real-Time System Stats

```http
GET /api/admin/realtime-stats
Authorization: Bearer <token>

# Returns system-wide real-time statistics
Response:
{
  "status": "active",
  "total_connected": 5,
  "connected_users": [...],
  "system_info": {
    "websocket_hub_active": true,
    "notification_service_active": true
  }
}
```

## 🔒 Security Features

### Authentication & Authorization

- **JWT-based Authentication**: Secure token-based authentication
- **Password Hashing**: Bcrypt encryption for password storage
- **Role-based Access Control**: Different permissions for owners vs collaborators
- **Project-level Security**: Users can only access projects they own or collaborate on

### Input Validation

- **Request Validation**: JSON binding with validation tags
- **Type Safety**: Strict type checking for all inputs
- **Error Handling**: Comprehensive error responses with appropriate HTTP status codes

## 🧪 Testing with Postman

### 🔐 Modern Authentication Testing Flow

#### 1. Complete Registration and Verification Flow

**Step 1: Register a new user**

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"testuser@example.com","password":"securepass123"}'

# Response:
# {
#   "message": "User registered successfully. Please check your email for verification link.",
#   "user_id": 1,
#   "email_verified": false
# }
```

**Step 2: Check console logs for verification email**

Look for console output like:

```
Email would be sent to testuser@example.com with subject: Verify Your Email Address
Body: <h2>Welcome to Task Manager!</h2>...
```

Extract the verification token from the URL in the email body.

**Step 3: Verify email**

```bash
curl -X POST http://localhost:8080/api/auth/verify-email \
  -H "Content-Type: application/json" \
  -d '{"token":"YOUR_VERIFICATION_TOKEN_HERE"}'

# Response:
# {"message": "Email verified successfully"}
```

**Step 4: Login successfully**

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"testuser@example.com","password":"securepass123"}'

# Response:
# {
#   "message": "Login successful",
#   "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
#   "user": {
#     "id": 1,
#     "email": "testuser@example.com",
#     "email_verified": true,
#     "two_factor_enabled": false
#   }
# }
```

#### 2. Two-Factor Authentication Setup

**Step 1: Setup 2FA (requires login token)**

```bash
export TOKEN="your_jwt_token_here"

curl -X POST http://localhost:8080/api/auth/2fa/setup \
  -H "Authorization: Bearer $TOKEN"

# Response:
# {
#   "message": "Two-factor authentication setup initiated",
#   "secret": "JBSWY3DPEHPK3PXP",
#   "qr_code": "otpauth://totp/Task%20Manager:testuser@example.com?secret=JBSWY3DPEHPK3PXP&issuer=Task%20Manager",
#   "instructions": "Scan the QR code with your authenticator app..."
# }
```

**Step 2: Configure authenticator app**

1. Open your authenticator app (Google Authenticator, Authy, etc.)
2. Scan the QR code or manually enter the secret
3. Get the 6-digit code from your app

**Step 3: Confirm 2FA setup**

```bash
curl -X POST http://localhost:8080/api/auth/2fa/confirm \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"token":"123456"}'  # Replace with actual 6-digit code

# Response:
# {
#   "message": "Two-factor authentication enabled successfully",
#   "backup_codes": ["ABC12345", "DEF67890", ...],
#   "warning": "Save these backup codes in a secure location..."
# }
```

**Step 4: Test login with 2FA**

```bash
# This will now require 2FA token
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"testuser@example.com","password":"securepass123"}'

# Response (401):
# {
#   "error": "Two-factor authentication required",
#   "two_factor_required": true
# }

# Login with 2FA token
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"testuser@example.com","password":"securepass123","two_factor_token":"123456"}'

# Response (200):
# {"message": "Login successful", "token": "...", "user": {...}}
```

#### 3. Password Reset Flow

**Step 1: Request password reset**

```bash
curl -X POST http://localhost:8080/api/auth/request-password-reset \
  -H "Content-Type: application/json" \
  -d '{"email":"testuser@example.com"}'

# Response:
# {"message": "Password reset email sent if account exists"}
```

**Step 2: Check console logs for reset email**

Look for console output with the reset token.

**Step 3: Reset password**

```bash
curl -X POST http://localhost:8080/api/auth/reset-password \
  -H "Content-Type: application/json" \
  -d '{"token":"YOUR_RESET_TOKEN_HERE","new_password":"newpassword123"}'

# Response:
# {"message": "Password reset successfully"}
```

#### 4. Security Features Testing

**Test account lockout (wrong password 5 times)**

```bash
for i in {1..6}; do
  curl -X POST http://localhost:8080/login \
    -H "Content-Type: application/json" \
    -d '{"email":"testuser@example.com","password":"wrongpassword"}'
  echo "Attempt $i"
done

# After 5 attempts:
# {"error": "account is temporarily locked due to too many failed attempts"}
```

**Test 2FA with backup codes**

```bash
# Use one of the backup codes instead of TOTP token
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"testuser@example.com","password":"securepass123","two_factor_token":"ABC12345"}'
```

### WebSocket Connection Setup

1. **Open Postman** and create a new WebSocket request
2. **Set the URL:** `ws://localhost:8080/api/ws`
3. **Add Headers:**
   - Key: `Authorization`
   - Value: `Bearer YOUR_JWT_TOKEN_HERE`
4. **Connect** - You should receive a welcome message with your user info and accessible projects

### Getting a JWT Token

First, get a valid JWT token by logging in:

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"password123"}'

# Response includes the token:
# {"message":"Login successful","token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}
```

### Testing Real-Time Messages

Once connected, you can:

1. **Send typing indicators:**

```json
{
  "type": "typing_start",
  "data": { "project_id": 1 }
}
```

2. **Update presence status:**

```json
{
  "type": "presence_update",
  "data": { "status": "away" }
}
```

3. **Create tasks via REST API** (in another Postman tab) and watch real-time updates in your WebSocket connection

### Example Testing Flow

1. **Register two users:**

```bash
# User 1 (Project Owner)
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"owner@example.com","password":"password123"}'

# User 2 (Collaborator)
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"collaborator@example.com","password":"password123"}'
```

2. **Login as User 1:**

```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"owner@example.com","password":"password123"}'

# Save the returned token
```

3. **Create a project:**

```bash
curl -X POST http://localhost:8080/api/projects \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Project","description":"A test project"}'

# Response will include the project ID:
# {"ID":1,"CreatedAt":"...","Name":"Test Project",...}
```

4. **Get the specific project:**

```bash
curl -X GET http://localhost:8080/api/projects/1 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

5. **Update the project:**

```bash
curl -X PUT http://localhost:8080/api/projects/1 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Project Name","description":"Updated description"}'
```

6. **Invite User 2 to collaborate:**

```bash
curl -X POST http://localhost:8080/api/projects/1/invite \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"user_email":"collaborator@example.com","message":"Join my project!"}'
```

#### 7. Login as User 2 and accept invitation:

```bash
# Login as User 2
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"collaborator@example.com","password":"password123"}'

# Get pending invitations
curl -X GET http://localhost:8080/api/invitations \
  -H "Authorization: Bearer USER2_TOKEN"

# Accept invitation
curl -X POST http://localhost:8080/api/invitations/1/respond \
  -H "Authorization: Bearer USER2_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"accept":true}'
```

8. **Create and manage tasks:**

```bash
# Create a task (as either user)
curl -X POST http://localhost:8080/api/projects/1/tasks \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Setup database","description":"Configure PostgreSQL"}'

# Assign task to collaborator
curl -X PUT http://localhost:8080/api/tasks/1 \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"assignee_id":2,"status":"in_progress"}'
```

9. **Clean up (optional) - Delete project:**

```bash
# Only project owner can delete projects
curl -X DELETE http://localhost:8080/api/projects/1 \
  -H "Authorization: Bearer YOUR_TOKEN"

# This will delete the project and all related data (tasks, collaborators, invitations)
```

## 🌟 What's Implemented

✅ **Modern Authentication System**

- **Email verification** on registration with secure token-based verification
- **Two-Factor Authentication (2FA)** with TOTP (Time-based One-Time Password)
- **QR code generation** for easy authenticator app setup
- **Backup codes** for 2FA recovery (10 single-use codes)
- **Password reset** with secure email-based token system
- **Account security** with login attempt tracking and automatic lockout
- **Security alerts** via email for sensitive account changes
- **JWT-based authentication** with proper token validation

✅ **Email Gateway Integration**

- **Email service** with SMTP configuration support
- **Verification emails** with HTML templates and secure links
- **Password reset emails** with time-limited tokens
- **Security alert notifications** for account changes
- **Development mode** with console logging (production-ready SMTP integration)

✅ **Core Features**

- **Complete project management CRUD operations** (Create, Read, Update, Delete)
- **Complete task management CRUD operations** with assignment capabilities
- **Project collaboration** with invitation system and role-based access
- **Cross-user task assignment** with auto-invitation system

✅ **Real-Time Features**

- WebSocket-based real-time updates
- Typing indicators for collaborative editing
- User presence tracking (online/offline/away)
- Live task creation, updates, and deletions
- Real-time collaboration notifications

✅ **Background Notifications**

- Asynchronous notification processing using Goroutines
- Task assignment notifications
- Task update and completion notifications
- Project invitation notifications
- Scalable worker pool architecture

✅ **Enhanced API Endpoints**

- Connected users tracking per project
- User presence status checking
- Real-time system statistics
- WebSocket connection management

✅ **Developer Experience**

- **Postman-Ready Testing**: Professional API testing with comprehensive authentication flows
- **Complete 2FA Testing**: QR codes, backup codes, and TOTP integration testing
- Comprehensive testing script for automated setup
- Detailed API documentation with examples
- Clean architecture with separation of concerns

## 🚀 Production Considerations

### Scaling Real-Time Features

- Use Redis for WebSocket connection state in multi-instance deployments
- Implement WebSocket connection pooling and load balancing
- Add rate limiting for WebSocket messages

### Notification System Enhancement

- Replace simulated emails with real email service (SendGrid, AWS SES)
- Add SMS notifications for critical updates
- Implement push notifications for mobile apps

### Performance Optimization

- Add caching for frequently accessed data
- Implement database connection pooling
- Add API response caching with Redis

### Security Hardening

- Add CORS configuration for production
- Implement request rate limiting
- Add input validation middleware
- Use environment variables for sensitive configuration

## 🔄 What's Next?

The current implementation provides a comprehensive collaborative task management system with modern authentication. Here are potential enhancements for scaling to enterprise level:

### Advanced Collaboration Features

- **Task comments and activity logs** with real-time updates
- **File attachments** for tasks with cloud storage integration
- **Project templates** for quick project setup
- **Task dependencies** and Gantt chart visualization
- **Advanced search and filtering** with full-text search
- **Project analytics** and reporting dashboards

### Enterprise Features

- **Role-based permissions** (admin, manager, developer, viewer roles)
- **Organization management** with multi-tenant support
- **SSO integration** (OAuth2, SAML) for enterprise authentication
- **Audit logging** for compliance and security tracking
- **API rate limiting** and usage analytics
- **Advanced caching** with Redis for improved performance

### Mobile and Integration

- **Mobile app APIs** optimized for mobile clients
- **Webhook notifications** for third-party integrations
- **Slack/Teams integration** for notifications
- **Calendar integration** for task deadlines
- **Time tracking** capabilities with reporting

### Infrastructure Scaling

- **Microservices architecture** for better scalability
- **Database sharding** for large-scale deployments
- **CDN integration** for file attachments
- **Container orchestration** with Kubernetes
- **Monitoring and observability** with proper metrics

## 🔧 Refresh Token Implementation

This document describes the refresh token functionality added to the authentication system.

### Overview

The application now supports JWT refresh tokens for enhanced security and better user experience. This implementation follows OAuth 2.0 best practices.

## Key Features

- **Short-lived Access Tokens**: Access tokens expire in 1 hour
- **Long-lived Refresh Tokens**: Refresh tokens expire in 7 days
- **Secure Storage**: Refresh tokens are stored securely in the database
- **Token Rotation**: New refresh tokens are generated on each refresh
- **Logout Support**: Both single device and all devices logout

## API Endpoints

### 1. Login

**POST** `/login`

Returns both access and refresh tokens:

```json
{
  "message": "Login successful",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "dGVzdF9yZWZyZXNoX3Rva2Vu...",
  "expires_at": "2025-07-06T15:30:00Z",
  "token_type": "Bearer",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe",
    "email_verified": true,
    "two_factor_enabled": false
  }
}
```

### 2. Refresh Token

**POST** `/api/auth/refresh`

Request body:

```json
{
  "refresh_token": "dGVzdF9yZWZyZXNoX3Rva2Vu..."
}
```

Response:

```json
{
  "message": "Token refreshed successfully",
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "bmV3X3JlZnJlc2hfdG9rZW4...",
  "expires_at": "2025-07-06T16:30:00Z",
  "token_type": "Bearer"
}
```

### 3. Logout

**POST** `/api/auth/logout` (Protected)

Revokes the current user's refresh token.

Response:

```json
{
  "message": "Logged out successfully"
}
```

### 4. Logout All Devices

**POST** `/api/auth/logout-all` (Protected)

Revokes all refresh tokens for the current user.

Response:

```json
{
  "message": "Logged out from all devices successfully"
}
```

## Database Changes

Added to `User` model:

- `RefreshToken` (string): Stores the current refresh token
- `RefreshTokenExpiry` (\*time.Time): Expiry time for the refresh token

## Security Considerations

1. **Access Token Expiry**: Short-lived (1 hour) to minimize exposure
2. **Refresh Token Expiry**: Longer-lived (7 days) but still expires
3. **Token Rotation**: New refresh token generated on each refresh
4. **Secure Storage**: Refresh tokens stored hashed in database
5. **Logout Functionality**: Ability to revoke tokens immediately

## Client Implementation

### Using Access Tokens

Include access tokens in the Authorization header:

```
Authorization: Bearer <access_token>
```

### Handling Token Expiry

When an API call returns 401 Unauthorized:

1. Use the refresh token to get a new access token
2. Retry the original request with the new access token
3. If refresh fails, redirect to login

### Example Client Flow

```javascript
async function apiCall(url, options = {}) {
  const token = localStorage.getItem("access_token");

  const response = await fetch(url, {
    ...options,
    headers: {
      ...options.headers,
      Authorization: `Bearer ${token}`,
    },
  });

  if (response.status === 401) {
    // Try to refresh token
    const refreshed = await refreshToken();
    if (refreshed) {
      // Retry with new token
      return apiCall(url, options);
    } else {
      // Redirect to login
      window.location.href = "/login";
    }
  }

  return response;
}

async function refreshToken() {
  const refreshToken = localStorage.getItem("refresh_token");

  try {
    const response = await fetch("/api/auth/refresh", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (response.ok) {
      const data = await response.json();
      localStorage.setItem("access_token", data.access_token);
      localStorage.setItem("refresh_token", data.refresh_token);
      return true;
    }
  } catch (error) {
    console.error("Failed to refresh token:", error);
  }

  return false;
}
```

## Backward Compatibility

The `GenerateToken()` function in `utils/jwt.go` is maintained for backward compatibility but now generates access tokens with 1-hour expiry instead of 24-hour expiry.

## Environment Variables

Ensure `JWT_SECRET` is set in your environment for token signing.

## Migration

When upgrading, existing users will need to log in again to get refresh tokens. The new fields will be automatically added to the database via GORM auto-migration.
