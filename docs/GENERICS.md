# Generics

GlyphLang supports generic type parameters with inference and constraints.

```glyph
! identity<T>(x: T): T {
  > x
}

! first<T>(a: T, b: T): T {
  > a
}

! map<T, U>(arr: [T], fn: (T) -> U): [U] {
  $ result = []
  for item in arr {
    $ mapped = fn(item)
    result = append(result, mapped)
  }
  > result
}
```