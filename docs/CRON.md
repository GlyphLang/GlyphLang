# Cron Tasks

GlyphLang supports scheduled tasks using cron expressions.

```glyph
# Daily cleanup at midnight
* "0 0 * * *" daily_cleanup {
  % db: Database
  > {task: "cleanup", timestamp: now()}
}

# Every 5 minutes health check
* "*/5 * * * *" health_check {
  > {status: "healthy", checked_at: now()}
}

# Weekly report on Sundays at 9am with retries
* "0 9 * * 0" weekly_report {
  + retries(3)
  % db: Database
  > {week: "current", generated_at: now()}
}
```