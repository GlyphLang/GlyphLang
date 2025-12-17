# Blog API

A complete blog platform API with posts, comments, and author management.

## Features

- Blog post management (CRUD operations)
- Nested comments on posts
- Author-based filtering
- Tag-based filtering
- Pagination support
- View counting
- Comment upvoting
- Authentication and authorization
- Rate limiting

## API Endpoints

### Posts

#### Get All Posts (Paginated)
```
GET /api/posts?page=1&per_page=10
```
Returns a paginated list of blog posts.

**Query Parameters:**
- `page` - Page number (default: 1)
- `per_page` - Items per page (default: 10, max: 50)

**Response:**
```json
{
  "posts": [...],
  "total": 100,
  "page": 1,
  "per_page": 10
}
```

#### Get Single Post with Comments
```
GET /api/posts/:id
```
Returns a post with all its comments. Automatically increments view count.

**Response:**
```json
{
  "post": {
    "id": 1,
    "title": "My First Post",
    "content": "Post content...",
    "author": "john_doe",
    "tags": ["tutorial", "glyph"],
    "published": true,
    "views": 42,
    "created_at": 1234567890,
    "updated_at": 1234567890
  },
  "comments": [...],
  "comments_count": 5
}
```

#### Create Post
```
POST /api/posts
```
Creates a new blog post. Requires authentication.

**Request Body:**
```json
{
  "title": "Post Title",
  "content": "Post content goes here...",
  "tags": ["tag1", "tag2"]
}
```

**Response:**
```json
{
  "success": true,
  "message": "Post created successfully",
  "data": {...}
}
```

#### Update Post
```
PUT /api/posts/:id
```
Updates an existing post. User must be the author or admin.

**Request Body:**
```json
{
  "title": "Updated Title",
  "content": "Updated content",
  "tags": ["new", "tags"],
  "published": true
}
```

#### Delete Post
```
DELETE /api/posts/:id
```
Deletes a post and all associated comments. Admin only.

#### Get Published Posts
```
GET /api/posts/published
```
Returns all published posts sorted by views (most popular first).

#### Get Posts by Tag
```
GET /api/posts/tags/:tag
```
Returns all posts containing the specified tag.

**Example:**
```
GET /api/posts/tags/tutorial
```

#### Get Posts by Author
```
GET /api/authors/:author/posts
```
Returns all posts by a specific author.

**Example:**
```
GET /api/authors/john_doe/posts
```

### Comments

#### Add Comment to Post
```
POST /api/posts/:id/comments
```
Adds a comment to a specific post. Requires authentication.

**Request Body:**
```json
{
  "content": "Great post! Thanks for sharing."
}
```

**Response:**
```json
{
  "success": true,
  "message": "Comment added successfully",
  "data": {
    "id": 1,
    "post_id": 5,
    "author": "jane_doe",
    "content": "Great post!",
    "upvotes": 0,
    "created_at": 1234567890
  }
}
```

#### Upvote Comment
```
PATCH /api/comments/:id/upvote
```
Increments the upvote count for a comment. Requires authentication.

## Type Definitions

### Post
```
: Post {
  id: int!
  title: str!
  content: str!
  author: str!
  tags: List[str]
  published: bool!
  views: int!
  created_at: timestamp
  updated_at: timestamp
}
```

### Comment
```
: Comment {
  id: int!
  post_id: int!
  author: str!
  content: str!
  upvotes: int!
  created_at: timestamp
}
```

### Author
```
: Author {
  id: int!
  name: str!
  email: str!
  bio: str
  posts_count: int!
  joined_at: timestamp
}
```

### PostWithComments
```
: PostWithComments {
  post: Post
  comments: List[Comment]
  comments_count: int!
}
```

## Authentication & Authorization

### Required for:
- Creating posts (any authenticated user)
- Updating posts (author or admin)
- Deleting posts (admin only)
- Adding comments (any authenticated user)
- Upvoting comments (any authenticated user)

### Public endpoints:
- Viewing posts
- Browsing published posts
- Filtering by tags/authors

## Rate Limiting

- GET endpoints: 100-200 requests/minute
- POST endpoints: 20-50 requests/minute
- PUT endpoints: 30 requests/minute
- DELETE endpoints: 10 requests/minute
- PATCH endpoints: 100 requests/minute

## Workflow Example

1. **Create a post:**
   ```
   POST /api/posts
   {
     "title": "Getting Started with Glyph",
     "content": "Glyph is a language for building APIs...",
     "tags": ["tutorial", "beginner"]
   }
   ```

2. **Publish the post:**
   ```
   PUT /api/posts/1
   {
     "published": true
   }
   ```

3. **Users view and comment:**
   ```
   GET /api/posts/1
   POST /api/posts/1/comments
   {
     "content": "Very helpful tutorial!"
   }
   ```

4. **Upvote helpful comments:**
   ```
   PATCH /api/comments/1/upvote
   ```

5. **Find related posts:**
   ```
   GET /api/posts/tags/tutorial
   ```

## Error Responses

```json
{
  "success": false,
  "message": "Error description",
  "code": "ERROR_CODE"
}
```

Common error codes:
- `NOT_FOUND` - Resource not found
- `VALIDATION_ERROR` - Invalid input
- `FORBIDDEN` - Insufficient permissions
