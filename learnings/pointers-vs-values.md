# Pointers vs Values in Go

## The Core Tradeoff

```go
// Value — every caller gets an independent copy
func NewSpan(...) Span { return Span{...} }

// Pointer — every caller looks at the same Span in memory
func NewSpan(...) *Span { return &Span{...} }
```

## What `*` and `&` mean

| Symbol | Where | Meaning |
|---|---|---|
| `*Span` | in a type | "a pointer to a Span" |
| `&Span{...}` | in code | "give me the address of this Span" |

A pointer is always **8 bytes** regardless of what it points to.

## Why Copying Hurts at Scale

```go
span := NewSpan("svc", "infer")  // one Span in memory

writeToStorage(span)   // copy #1 — ~112 bytes
sendToCollector(span)  // copy #2 — ~112 bytes
logSpan(span)          // copy #3 — ~112 bytes
```

At 100k spans/second, copying 112 bytes vs passing an 8-byte pointer adds up fast.

## The Silent Bug Copies Create

```go
func endSpan(s Span) {
    s.DurationNanos = 12345
    s.Status = "ok"
}

endSpan(span)
fmt.Println(span.Status) // prints "" — original was never updated
```

No error, no panic. The function ran but your original Span is unchanged. Pointers prevent this.

## Rule of Thumb

- **Struct you'll mutate or pass around** → pointer (`*Span`)
- **Small config/filter/ID types** → value (`SpanFilter`)
