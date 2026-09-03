# Week 8 – Mock Interview + Revision – Study Notes

## Mock Interview Format

### Mock 1: Go Internals (60 min)

**Round 1 – Quick Fire (15 min):**
- Array vs Slice internals
- Map implementation
- Interface iface/eface
- Defer execution order
- Nil interface trap
- Method set rules

**Round 2 – Concurrency Deep Dive (20 min):**
- GMP model — explain each component
- Channel internals (hchan struct)
- Goroutine leak scenario + fix
- Worker pool implementation
- Mutex vs Channel decision

**Round 3 – Runtime (15 min):**
- Escape analysis — when variables escape
- GC tri-color marking
- STW duration and phases
- pprof usage in production

**Round 4 – Coding (10 min):**
- Implement rate limiter using channel
- Fix race condition in given code
- Implement timeout with context

---

### Mock 2: Backend Engineering (60 min)

**Topics:**
1. HTTP Client configuration (timeouts, pool)
2. Middleware chain implementation
3. Graceful shutdown flow
4. Context best practices
5. API idempotency implementation
6. Connection pool exhaustion debugging
7. Database transaction patterns

**Sample Questions:**
- "Your service returns 504 intermittently. How do you debug?"
- "Design a middleware that limits concurrent requests"
- "How do you implement graceful shutdown in Kubernetes?"

---

### Mock 3: System Design (45 min)

**Format:**
```
0-5 min:  Requirements clarification
5-8 min:  Scale estimation
8-12 min: API design
12-15 min: Data model
15-30 min: Architecture + deep dive
30-40 min: Scaling + failure handling
40-45 min: Observability + wrap-up
```

**Sample Problems:**
1. Design a payment system (10K TPS)
2. Design a real-time chat system
3. Design a notification system
4. Design a URL shortener at scale

---

### Mock 4: Production Troubleshooting (30 min)

**Approach Template:**
```
1. Understand symptoms (what changed? when? scope?)
2. Check metrics (latency, error rate, throughput)
3. Check logs (error patterns, timing)
4. Check traces (where is time spent?)
5. Narrow scope (which service? which endpoint?)
6. Form hypothesis
7. Validate (profile, query plan, etc.)
8. Fix
9. Verify fix
10. Post-mortem (root cause + prevention)
```

---

## Senior Interview Response Patterns

### Pattern 1: Technical Explanation

```
What → How → Why → Trade-offs → When to use/not use
```

Example: "What is escape analysis?"
- **What**: Compiler determines at compile time if variable goes to stack or heap
- **How**: Analyzes if variable's address escapes function scope
- **Why**: Stack is cheaper (no GC), want to minimize heap allocations
- **Trade-offs**: Sometimes pointer for large struct forces heap but avoids copy cost
- **When**: Check with `-gcflags="-m"`, optimize hot paths only

### Pattern 2: Problem Solving

```
Understand → Scope → Approach → Trade-offs → Decision
```

Example: "How would you handle cache stampede?"
- **Understand**: Many requests hit expired key simultaneously → DB overloaded
- **Scope**: Affects hot keys, during TTL expiration
- **Approach options**: Mutex, singleflight, probabilistic refresh, never-expire
- **Trade-offs**: Mutex adds latency for waiters, singleflight adds complexity
- **Decision**: singleflight for moderate traffic, never-expire + background refresh for extreme cases

### Pattern 3: Design Decision

```
Context → Options → Criteria → Decision → Validation
```

Example: "Mutex or Channel?"
- **Context**: Protecting shared map accessed by 10 goroutines
- **Options**: sync.Mutex, sync.RWMutex, sync.Map, channel
- **Criteria**: Read-heavy (90%), simple lock/unlock, performance matters
- **Decision**: sync.RWMutex (multiple readers, single writer)
- **Validation**: Benchmark showed 3x improvement over Mutex for read-heavy workload

---

## Common Behavioral Questions

### "Tell me about a difficult production issue"

**STAR Format:**
- **Situation**: What was the system, scale, team?
- **Task**: What was the problem? Impact?
- **Action**: What did YOU do? (Be specific about your contribution)
- **Result**: How was it resolved? What was the outcome? What did you learn?

### "How do you make technical decisions?"

Senior answer framework:
1. Understand requirements and constraints
2. Identify options (at least 2-3)
3. Evaluate trade-offs (performance, complexity, team ability)
4. Consider maintenance and future scalability
5. Document decision and reasoning
6. Get team input for significant decisions
7. Be willing to revisit if assumptions change

### "How do you mentor junior engineers?"

- Code review with explanations (not just approve/reject)
- Pair programming on complex tasks
- Let them struggle (guided, not rescued immediately)
- Share context and reasoning behind decisions
- Create learning opportunities in ticket assignment
- Regular 1:1s for career growth discussion

---

## Production Troubleshooting Scenarios

### Scenario 1: Latency Spike

```
Symptom: P99 latency went from 100ms to 3s at 2pm

Debug flow:
1. Metrics: when exactly? gradual or sudden?
2. Correlation: any deployment? traffic spike?
3. Tracing: which service/endpoint is slow?
4. Profile: CPU? Memory? I/O?
5. Database: connection pool? slow queries? locks?
6. Network: DNS? timeout? packet loss?

Common causes:
• Database lock contention
• GC pressure (heap grew)
• Connection pool exhaustion
• Downstream service degradation
• Memory leak causing swap
• DNS resolution timeout
```

### Scenario 2: Memory Leak

```
Symptom: Service memory grows continuously, OOM after 24h

Debug flow:
1. Heap profile: which objects accumulate?
2. Goroutine profile: goroutine leak?
3. Code review: unbounded caches? unclosed resources?
4. Connection tracking: leaked DB/HTTP connections?

Common causes:
• Goroutine leak (blocked channel, no context cancellation)
• Unbounded in-memory cache
• String/slice references preventing GC
• Unclosed response bodies
• time.Ticker not stopped
```

### Scenario 3: Intermittent Errors

```
Symptom: 0.1% of requests return 500, random pattern

Debug flow:
1. Error logs: what error? which code path?
2. Correlation: same users? same endpoints? time pattern?
3. Dependencies: is one downstream service flaky?
4. Connection pool: exhaustion under load?
5. Race condition: only under concurrency?

Common causes:
• Race condition (run with -race flag)
• Connection pool timeout
• Downstream service intermittent failure
• Resource exhaustion (file descriptors, goroutines)
• Nil pointer from cache miss timing
```

---

## Final Review: Key Concepts to Explain Without Notes

### Go (must be fluent)
1. Slice internals (header, append, sharing)
2. Map internals (buckets, not thread-safe)
3. Interface (iface/eface, nil trap, method set)
4. GMP scheduler (work stealing, preemption)
5. Channel (hchan, blocking rules, nil/closed behavior)
6. Escape analysis (rules, how to check)
7. GC (tri-color, phases, STW duration)
8. pprof (types, how to use in production)
9. Context (cancellation, values, best practices)

### Backend (must explain clearly)
10. HTTP connection pooling (Transport config)
11. Graceful shutdown (signal → drain → close)
12. Middleware pattern (chain, response writer wrapping)
13. Database transactions (isolation, deadlock, N+1)
14. Redis caching (strategies, problems, distributed lock)
15. Kafka (ordering, delivery semantics, consumer groups)

### Distributed Systems (must reason about trade-offs)
16. CAP theorem (real-world choices)
17. Saga pattern (orchestration vs choreography)
18. Outbox pattern (atomic publish + DB update)
19. Idempotency (implementation details)
20. Circuit breaker (states, transitions)
21. Rate limiting (algorithms, distributed)

### System Design (must structure answer)
22. Design framework (requirements → scale → architecture)
23. At least 3 complete designs practiced end-to-end
