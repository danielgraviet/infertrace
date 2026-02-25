# Struct Design — How to Build a Data Model

## The Core Question

A struct should contain exactly what you need to describe **one thing, at one point in time.**

---

## Step 1: Ask "what questions will I need to answer about this thing?"

For a `Span`, the consumers of your data are:
- **The storage layer** — what do I need to save and query later?
- **The UI** — what do I need to display in a trace timeline?
- **The analysis engine** — what do I need to detect anomalies?

Write those questions down, then each answer becomes a field:

| Question | Field |
|---|---|
| Which request does this belong to? | `TraceID` |
| What is this specific step? | `SpanID` |
| Who called this step? | `ParentSpanID` |
| Which service did the work? | `ServiceName` |
| What operation was performed? | `OperationName` |
| When did it start? | `StartTimeUnixNano` |
| How long did it take? | `DurationNanos` |
| Did it succeed or fail? | `Status` |

---

## Three Rules for Clean Structs

**1. Only include what this type *owns*.**
A `Span` owns its timing and identity. It doesn't own the database connection that saved it, or the HTTP client that sent it. Those belong elsewhere.

**2. Don't add fields speculatively.**
Only add a field when something actually needs it. "We might need this later" is how structs become unmanageable.

**3. Follow the data flow.**
Trace where your data comes from and where it goes:

```
Instrumented service          →  Collector  →  Storage  →  Query API
creates Span with real values     receives      persists     returns
```

The struct needs to survive that entire journey, so it needs every field required at each stage.

---

## Pointer Fields vs Value Fields

When a struct has nested data that is only *sometimes* present, use a pointer:

```go
type Span struct {
    // Always present — plain value
    TraceID    string
    SpanID     string
    DurationNanos int64

    // Only present for ML spans — pointer (can be nil)
    Inference  *InferenceContext
    Resources  *ResourceMetrics
}
```

**Pointer (`*T`) = optional.** The field can be `nil` if the data doesn't exist.
**Value (`T`) = always present.** Go initializes it to the zero value even if you don't set it.

Rule of thumb: if a non-ML service sends a span, it won't have inference data. Making it a pointer avoids wasting memory on empty structs and lets you check `if span.Inference != nil` cleanly.
