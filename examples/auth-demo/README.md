# Authentication Demo API

A complete authentication system demonstrating JWT tokens, protected routes, and role-based access control (RBAC).

## Features

- User registration and login
- JWT token authentication
- Token refresh mechanism
- Password hashing and verification
- Protected routes requiring authentication
- Role-based access control (User, Moderator, Admin)
- Profile management
- Password change functionality
- Rate limiting
- Secure error handling

## API Endpoints

### Public Endpoints (No Authentication Required)

#### Health Check
```
GET /api/health
```
Check API status and version.

**Response:**
```json
{
  "status": "ok",
  "timestamp": 1234567890,
  "version": "1.0.0"
}
```

#### Register
```
POST /api/auth/register
```
Create a new user account.

**Request Body:**
```json
{
  "username": "johndoe",
  "email": "john@example.com",
  "password": "securepass123",
  "full_name": "John Doe"
}
```

**Response:**
```json
{
  "success": true,
  "message": "User registered successfully",
  "token": "eyJhbGc...",
  "user": {
    "id": 1,
    "username": "johndoe",
    "email": "john@example.com",
    "role": "user",
    "created_at": 1234567890,
    "last_login": 1234567890
  }
}
```

**Validations:**
- Username required and must be unique
- Email required and must be unique
- Password must be at least 8 characters
- Full name is optional

#### Login
```
POST /api/auth/login
```
Authenticate user and receive JWT token.

**Request Body:**
```json
{
  "username": "johndoe",
  "password": "securepass123"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "token": "eyJhbGc...",
  "user": {...}
}
```

### Protected Endpoints (Authentication Required)

All protected endpoints require the JWT token in the Authorization header:
```
Authorization: Bearer <token>
```

#### Logout
```
POST /api/auth/logout
```
Logout current user (invalidates token on client side).

#### Get Current User
```
GET /api/auth/me
```
Get the authenticated user's profile.

**Response:**
```json
{
  "id": 1,
  "username": "johndoe",
  "email": "john@example.com",
  "role": "user",
  "full_name": "John Doe",
  "bio": "Software developer",
  "avatar_url": "https://...",
  "created_at": 1234567890
}
```

#### Update Profile
```
PUT /api/auth/profile
```
Update user profile information.

**Request Body:**
```json
{
  "full_name": "John Smith",
  "bio": "Full-stack developer",
  "avatar_url": "https://example.com/avatar.jpg"
}
```

#### Change Password
```
PUT /api/auth/password
```
Change user password.

**Request Body:**
```json
{
  "current_password": "oldpass123",
  "new_password": "newpass456"
}
```

**Validations:**
- Current password must be correct
- New password must be at least 8 characters
- New password must be different from current

#### Refresh Token
```
POST /api/auth/refresh
```
Generate a new JWT token before expiration.

**Response:**
```json
{
  "success": true,
  "token": "eyJhbGc...",
  "expires_at": 1234567890
}
```

### Admin Endpoints (Admin Role Required)

#### Get All Users
```
GET /api/admin/users
```
Retrieve list of all users. Admin only.

**Response:**
```json
{
  "users": [...],
  "total": 100
}
```

#### Delete User
```
DELETE /api/admin/users/:id
```
Delete a user account. Admin only.

**Response:**
```json
{
  "success": true,
  "message": "User deleted successfully",
  "data": {...}
}
```

**Rules:**
- Admin cannot delete their own account
- Deletes user profile and all associated data

### Moderator Endpoints (Moderator or Admin Role Required)

#### Update User Role
```
PUT /api/admin/users/:id/role
```
Change a user's role. Moderator or Admin only.

**Request Body:**
```json
{
  "role": "moderator"
}
```

**Valid Roles:**
- `user` - Standard user (default)
- `moderator` - Can manage users
- `admin` - Full administrative access

**Rules:**
- Only admins can assign the admin role
- Moderators can assign user and moderator roles

## Type Definitions

### User
```
: User {
  id: int!
  username: str!
  email: str!
  role: str!
  created_at: timestamp
  last_login: timestamp
}
```

### UserProfile
```
: UserProfile {
  id: int!
  username: str!
  email: str!
  role: str!
  full_name: str
  bio: str
  avatar_url: str
  created_at: timestamp
}
```

### AuthResponse
```
: AuthResponse {
  success: bool!
  message: str!
  token: str
  user: User
}
```

## Authentication Flow

### Registration Flow
1. User submits registration form
2. Server validates input
3. Server checks for duplicate username/email
4. Password is hashed using bcrypt
5. User record is created with role "user"
6. JWT token is generated and returned
7. Client stores token (localStorage/cookies)

### Login Flow
1. User submits credentials
2. Server finds user by username
3. Server verifies password hash
4. Last login timestamp is updated
5. JWT token is generated and returned
6. Client stores token

### Protected Route Access
1. Client includes JWT in Authorization header
2. Server validates token signature
3. Server extracts user info from token
4. Request proceeds with auth context
5. Response is returned

### Role-Based Access
1. Route specifies required role via middleware
2. Server validates token and extracts role
3. Server checks if user role has permission
4. Request proceeds if authorized
5. 403 Forbidden if unauthorized

## JWT Token Structure

Tokens contain the following claims:
```json
{
  "user_id": 1,
  "username": "johndoe",
  "role": "user",
  "iat": 1234567890,
  "exp": 1234567890
}
```

- Default expiration: 7 days
- Use refresh endpoint to get new token

## Security Features

1. **Password Security**
   - Minimum 8 characters required
   - Passwords are hashed using bcrypt
   - Passwords never returned in responses

2. **Token Security**
   - JWT tokens signed with secret key
   - Tokens expire after 7 days
   - Refresh mechanism available

3. **Role-Based Access Control**
   - Three-tier role system
   - Middleware enforces role requirements
   - Admins have full access

4. **Rate Limiting**
   - Login: 20 requests/minute
   - Registration: 10 requests/minute
   - Password change: 10 requests/minute
   - Profile updates: 30 requests/minute
   - Admin operations: 10-50 requests/minute

5. **Input Validation**
   - Required field checking
   - Length validations
   - Format validations
   - Duplicate checking

## Error Responses

```json
{
  "success": false,
  "message": "Error description",
  "code": "ERROR_CODE"
}
```

Common error codes:
- `VALIDATION_ERROR` - Invalid input data
- `USERNAME_EXISTS` - Username already taken
- `EMAIL_EXISTS` - Email already registered
- `INVALID_CREDENTIALS` - Wrong username/password
- `INVALID_PASSWORD` - Incorrect current password
- `NOT_FOUND` - Resource not found
- `FORBIDDEN` - Insufficient permissions

## Example Workflow

### 1. Register a new user
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "email": "alice@example.com",
    "password": "securepass123",
    "full_name": "Alice Smith"
  }'
```

### 2. Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "password": "securepass123"
  }'
```

### 3. Access protected route
```bash
curl -X GET http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer eyJhbGc..."
```

### 4. Update profile
```bash
curl -X PUT http://localhost:8080/api/auth/profile \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "Alice Johnson",
    "bio": "DevOps Engineer"
  }'
```

### 5. Change password
```bash
curl -X PUT http://localhost:8080/api/auth/password \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "Content-Type: application/json" \
  -d '{
    "current_password": "securepass123",
    "new_password": "newsecurepass456"
  }'
```

## Role Hierarchy

```
Admin
  ├── Full access to all endpoints
  ├── Can delete users
  ├── Can assign any role
  └── Cannot delete own account

Moderator
  ├── Can view users
  ├── Can update user roles (except admin)
  └── Can manage content

User
  ├── Can access own profile
  ├── Can update own profile
  └── Can change own password
```

## Best Practices

1. **Token Storage**
   - Store tokens securely (httpOnly cookies preferred)
   - Clear tokens on logout
   - Refresh before expiration

2. **Password Management**
   - Use strong passwords
   - Change passwords regularly
   - Never share passwords

3. **API Usage**
   - Always use HTTPS in production
   - Handle token expiration gracefully
   - Implement proper error handling

4. **Security**
   - Validate all user input
   - Use rate limiting
   - Log authentication events
   - Monitor for suspicious activity
