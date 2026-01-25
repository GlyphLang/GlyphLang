# Async / Await

This document explains how asynchronous execution works in GlyphLang.
The examples below demonstrate async blocks and await usage.

```glyph
# Basic async block - executes in background, returns Future
@ GET /compute {
  $ future = async {
    $ x = 10
    $ y = 20
    > x + y
  }
  $ result = await future
  > {value: result}
}

# Parallel execution - all requests run concurrently
@ GET /dashboard {
  $ userFuture = async { > db.getUser(userId) }
  $ ordersFuture = async { > db.getOrders(userId) }
  $ statsFuture = async { > db.getStats(userId) }

  # Await blocks until Future resolves
  $ user = await userFuture
  $ orders = await ordersFuture
  $ stats = await statsFuture

  > {user: user, orders: orders, stats: stats}
}

```
