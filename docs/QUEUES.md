# Queue Workers

GlyphLang supports queue-based background workers with concurrency control.

```glyph
# Email sending worker with concurrency and retries
& "email.send" {
  + concurrency(5)
  + retries(3)
  + timeout(30)

  $ to = message.to
  $ subject = message.subject
  > {sent: true, to: to}
}

# Image processing worker
& "image.process" {
  + concurrency(3)
  + timeout(120)

  $ image_id = message.image_id
  > {processed: true, image_id: image_id}
}
```
