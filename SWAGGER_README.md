# Real Task API - Swagger Documentation

This project now includes comprehensive Swagger/OpenAPI documentation for all API endpoints.

## Accessing the Documentation

Once you start the server, you can access the interactive Swagger UI at:

```
http://localhost:8080/docs/index.html
```

## Quick Start

1. **Start the server:**

   ```bash
   go run cmd/api/main.go
   ```

2. **Open your browser and navigate to:**

   ```
   http://localhost:8080/docs/index.html
   ```

3. **Explore the API endpoints, test them directly in the browser!**

## Available API Documentation

The Swagger documentation includes all endpoints organized by the following tags:

### 🔐 Authentication

- `POST /register` - Register a new user
- `POST /login` - User login with 2FA support
- `GET|POST /api/auth/verify-email` - Verify user email
- `POST /api/auth/resend-verification` - Resend verification email
- `POST /api/auth/request-password-reset` - Request password reset
- `POST /api/auth/refresh` - Refresh access token

### 👤 User

- `GET /api/profile` - Get user profile

### 🔗 OAuth

- `GET /api/auth/providers` - Get available OAuth providers
- `GET /api/auth/google/url` - Get Google OAuth URL

### 📋 Projects

- `POST /api/projects` - Create a new project
- `GET /api/projects` - Get all user projects
- `GET /api/projects/{project_id}` - Get specific project
- `PUT /api/projects/{project_id}` - Update project
- `DELETE /api/projects/{project_id}` - Delete project

### ✅ Tasks

- `POST /api/projects/{project_id}/tasks` - Create a new task
- `GET /api/projects/{project_id}/tasks` - Get all tasks for project
- `GET /api/tasks/{task_id}` - Get specific task
- `PUT /api/tasks/{task_id}` - Update task
- `DELETE /api/tasks/{task_id}` - Delete task

### 🤝 Collaboration

- `POST /api/projects/{project_id}/invite` - Invite user to project
- `GET /api/invitations` - Get pending invitations
- `POST /api/invitations/{invitation_id}/respond` - Accept/decline invitation
- `GET /api/projects/{project_id}/collaborators` - Get project collaborators
- `DELETE /api/projects/{project_id}/collaborators/{user_id}` - Remove collaborator

### ⚡ Real-time

- `GET /api/projects/{project_id}/connected-users` - Get connected users
- `GET /api/users/{user_id}/presence` - Get user presence
- `GET /api/admin/realtime-stats` - Get real-time statistics

## Authentication

Most endpoints require authentication using JWT Bearer tokens. In the Swagger UI:

1. Click the **"Authorize"** button at the top right
2. Enter your JWT token in the format: `Bearer your_jwt_token_here`
3. Click **"Authorize"** to apply the token to all requests

## Request/Response Models

The documentation includes detailed schemas for:

- **UserResponse** - User profile information
- **ProjectResponse** - Project details with owner information
- **TaskResponse** - Task details with assignee information
- **AuthResponse** - Authentication response with tokens
- **ErrorResponse** - Standardized error responses
- **SuccessResponse** - Generic success responses

## Testing Endpoints

You can test all endpoints directly from the Swagger UI:

1. **Expand an endpoint** by clicking on it
2. **Click "Try it out"** to enable the test interface
3. **Fill in the required parameters** and request body
4. **Click "Execute"** to send the request
5. **View the response** including status code, headers, and body

## Example Usage

### 1. Register a new user

```bash
curl -X POST "http://localhost:8080/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword"
  }'
```

### 2. Login to get access token

```bash
curl -X POST "http://localhost:8080/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword"
  }'
```

### 3. Create a project (with Bearer token)

```bash
curl -X POST "http://localhost:8080/api/projects" \
  -H "Authorization: Bearer your_jwt_token_here" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Project",
    "description": "Project description"
  }'
```

## Features

✅ **Complete API Coverage** - All endpoints documented
✅ **Interactive Testing** - Test endpoints directly in browser
✅ **Authentication Support** - JWT Bearer token integration
✅ **Request/Response Schemas** - Detailed data models
✅ **Error Handling** - Comprehensive error responses
✅ **Real-time Features** - WebSocket and presence endpoints
✅ **Collaboration** - Project sharing and invitations
✅ **OAuth Support** - Social authentication endpoints

## Swagger Files

The documentation generates three files:

- `docs/docs.go` - Go code for embedding documentation
- `docs/swagger.json` - JSON specification file
- `docs/swagger.yaml` - YAML specification file

## Regenerating Documentation

If you modify the API endpoints or add new ones:

1. **Add Swagger annotations** to your handler functions
2. **Regenerate the documentation:**
   ```bash
   go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/api/main.go
   ```
3. **Restart the server** to see the changes

## Swagger Annotation Examples

Here's how to add documentation to a new endpoint:

```go
// CreateExample handles POST /api/examples
// @Summary Create a new example
// @Description Create a new example resource
// @Tags Examples
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ExampleRequest true "Example details"
// @Success 201 {object} ExampleResponse "Example created successfully"
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 401 {object} ErrorResponse "User not authenticated"
// @Router /api/examples [post]
func CreateExample(c *gin.Context) {
    // Implementation here
}
```

## Additional Resources

- **Swagger/OpenAPI Specification:** https://swagger.io/specification/
- **Swaggo Documentation:** https://github.com/swaggo/swag
- **Gin Framework:** https://gin-gonic.com/

---

🎉 **Happy API Testing!** The Swagger documentation makes it easy to explore, understand, and test all the available endpoints in your Real Task API.
