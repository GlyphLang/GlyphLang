# Event Handlers

GlyphLang supports event-based handlers for reactive systems.
The examples below demonstrate synchronous and asynchronous event handlers.

```glyph
# Handle user creation event
~ "user.created" {
  $ user_id = event.id
  $ email = event.email
  > {handled: true, user_id: user_id}
}

# Async event handler
~ "order.completed" async {
  $ order_id = event.order_id
  $ total = event.total
  > {processed: true}
}

