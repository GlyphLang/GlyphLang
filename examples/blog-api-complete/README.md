# Blog API - Complete Example

A full-featured blog API demonstrating Glyph's capabilities.

## Features

- ✅ CRUD operations for posts
- ✅ Comments system
- ✅ Pagination & filtering
- ✅ Search functionality
- ✅ Categories & tags
- ✅ Statistics dashboard
- ✅ Input validation
- ✅ XSS protection
- ✅ Like & view tracking

## Quick Start

```bash
# Run with hot-reload
./glyph dev examples/blog-api-complete/main.glyph

# Run with bytecode VM
./glyph run examples/blog-api-complete/main.glyph --bytecode
```

## API Endpoints

### Health & Info
- `GET /` - API information
- `GET /health` - Health check

### Posts
- `GET /posts` - List posts (with pagination & filters)
- `GET /posts/:id` - Get single post
- `POST /posts` - Create post
- `PUT /posts/:id` - Update post
- `DELETE /posts/:id` - Delete post

### Post Actions
- `POST /posts/:id/like` - Like a post
- `POST /posts/:id/view` - Track view

### Comments
- `GET /posts/:id/comments` - List comments
- `POST /posts/:id/comments` - Add comment

### Discovery
- `GET /search?query=keyword` - Search posts
- `GET /categories` - List categories
- `GET /tags` - List tags
- `GET /stats` - Get statistics

## Example Requests

### List Posts with Filters

```bash
# Get published posts by Alice, page 1, 10 per page
curl "http://localhost:3000/posts?author=Alice%20Johnson&published=true&page=1&limit=10"
```

### Create Post

```bash
curl -X POST http://localhost:3000/posts \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My Amazing Post",
    "content": "This is the full content of my post...",
    "excerpt": "Short summary here",
    "author": "Alice Johnson",
    "category": "Tutorial",
    "tags": ["glyph", "tutorial"],
    "published": true,
    "featured": false
  }'
```

### Add Comment

```bash
curl -X POST http://localhost:3000/posts/1/comments \
  -H "Content-Type: application/json" \
  -d '{
    "author": "Bob Smith",
    "content": "Great article!"
  }'
```

### Search

```bash
curl "http://localhost:3000/search?query=database&category=tutorial"
```

## Response Examples

### List Posts Response

```json
{
  "posts": [
    {
      "id": 1,
      "title": "Getting Started with Glyph",
      "slug": "getting-started-with-glyph",
      "excerpt": "Learn the basics...",
      "author": "Alice Johnson",
      "category": "Tutorial",
      "tags": ["glyph", "tutorial"],
      "published": true,
      "views": 1523,
      "likes": 87,
      "comments_count": 12
    }
  ],
  "meta": {
    "pagination": {
      "current_page": 1,
      "per_page": 10,
      "total_items": 5,
      "total_pages": 1,
      "has_next": false,
      "has_prev": false
    }
  }
}
```

## Next Steps

- Add database integration
- Implement authentication
- Add image upload for posts
- Add markdown support
- Add email notifications
