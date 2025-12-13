# Production-Grade Pagination Strategies

## Overview
Fetching all records without pagination is a critical anti-pattern in production systems. This guide explains the strategies used in real-world applications.

---

## The Problem: Unpaginated Queries

```go
// ❌ BAD - Don't do this in production
func GetAllStudents() ([]Student, error) {
    // Loads ALL records into memory
    return db.Query("SELECT * FROM students")
}
```

### Consequences:
- **Memory exhaustion**: 1M records × 1KB each = ~1GB RAM
- **Slow queries**: Full table scans, no early termination
- **Network bottleneck**: Huge JSON payloads
- **Database locks**: Long-running queries block other operations
- **Poor UX**: Frontend can't render millions of items anyway
- **Timeout risk**: Requests may timeout before completion

---

## Solution 1: Offset-Based Pagination

### How it Works
```
Page 1: LIMIT 20 OFFSET 0   → Records 1-20
Page 2: LIMIT 20 OFFSET 20  → Records 21-40
Page 3: LIMIT 20 OFFSET 40  → Records 41-60
```

### API Usage
```bash
# Get first page (default: 20 items)
GET /api/students

# Get page 2 with 50 items per page
GET /api/students?page=2&limit=50

# Maximum limit enforced (100 in our implementation)
GET /api/students?limit=1000  # Returns max 100 items
```

### Response Format
```json
{
  "data": [
    {"id": 21, "name": "John", "email": "john@example.com", "age": 20},
    {"id": 22, "name": "Jane", "email": "jane@example.com", "age": 22}
  ],
  "page": 2,
  "limit": 20,
  "total_items": 1500,
  "total_pages": 75,
  "has_next": true,
  "has_prev": true
}
```

### Pros
- ✅ Easy to implement
- ✅ Familiar UX (page numbers)
- ✅ Can jump to any page
- ✅ Good for small-medium datasets

### Cons
- ❌ Performance degrades with large offsets
  - `OFFSET 1000000` still scans 1M rows
- ❌ Inconsistent results if data changes between pages
  - Adding/deleting records shifts pages
- ❌ Not suitable for infinite scroll

### When to Use
- Admin panels with page numbers
- Small-medium datasets (< 1M records)
- When users need to jump to specific pages

---

## Solution 2: Cursor-Based Pagination (Advanced)

### How it Works
Uses the last item's ID as a cursor for the next page.

```sql
-- Page 1
SELECT * FROM students WHERE id > 0 ORDER BY id LIMIT 20

-- Page 2 (last ID was 20)
SELECT * FROM students WHERE id > 20 ORDER BY id LIMIT 20

-- Page 3 (last ID was 40)
SELECT * FROM students WHERE id > 40 ORDER BY id LIMIT 20
```

### API Usage
```bash
# First page
GET /api/students?limit=20

# Next page (cursor from previous response)
GET /api/students?cursor=eyJpZCI6MjB9&limit=20
```

### Response Format
```json
{
  "data": [...],
  "next_cursor": "eyJpZCI6NDB9",  // Base64 encoded cursor
  "has_more": true
}
```

### Implementation Example
```go
func GetStudentsCursor(afterID int64, limit int) ([]Student, int64, error) {
    query := "SELECT * FROM students WHERE id > ? ORDER BY id LIMIT ?"
    rows := db.Query(query, afterID, limit+1)
    
    var students []Student
    for rows.Next() {
        // Scan into students
    }
    
    hasMore := len(students) > limit
    if hasMore {
        students = students[:limit]  // Remove extra record
    }
    
    var nextCursor int64
    if len(students) > 0 {
        nextCursor = students[len(students)-1].ID
    }
    
    return students, nextCursor, nil
}
```

### Pros
- ✅ Consistent performance (always uses index)
- ✅ Stable results (no shifting pages)
- ✅ Perfect for infinite scroll
- ✅ Scales to billions of records

### Cons
- ❌ Can't jump to arbitrary pages
- ❌ Can't show total page count
- ❌ More complex implementation

### When to Use
- Social media feeds (infinite scroll)
- Large datasets (millions of records)
- Real-time data with frequent inserts
- APIs where page jumping isn't needed

---

## Solution 3: Keyset Pagination (Multi-column)

For complex sorting (e.g., by name then ID).

```sql
-- Sort by name, then ID
SELECT * FROM students 
WHERE (name, id) > ('John', 100)
ORDER BY name, id 
LIMIT 20
```

### When to Use
- Sorting by non-unique columns
- Complex multi-column ordering
- When cursor-based isn't enough

---

## Solution 4: Search/Filtering First

Reduce the dataset before paginating.

```bash
# Filter first, then paginate
GET /api/students?age_min=18&age_max=25&major=CS&page=1&limit=20
```

```go
func GetStudentsFiltered(filters Filters, page, limit int) {
    query := "SELECT * FROM students WHERE age BETWEEN ? AND ? AND major = ?"
    // Add pagination
    query += " LIMIT ? OFFSET ?"
}
```

---

## Additional Production Strategies

### 1. **Result Caching**
```go
// Cache total count (it changes rarely)
func (s *Store) GetStudentsCount() (int64, error) {
    // Check cache first
    if count := cache.Get("students:count"); count != nil {
        return count.(int64), nil
    }
    
    // Query database
    count, err := s.db.QueryCount()
    
    // Cache for 5 minutes
    cache.Set("students:count", count, 5*time.Minute)
    return count, err
}
```

### 2. **Database Indexing**
```sql
-- Essential for pagination performance
CREATE INDEX idx_students_id ON students(id);
CREATE INDEX idx_students_created_at ON students(created_at);

-- For filtered queries
CREATE INDEX idx_students_age_name ON students(age, name);
```

### 3. **Rate Limiting**
```go
// Prevent abuse
func RateLimitMiddleware(maxRequests int, window time.Duration) {
    // Limit to 100 requests per minute
}
```

### 4. **Streaming for Exports**
For large exports (CSV, Excel):

```go
func ExportStudents(w http.ResponseWriter) {
    w.Header().Set("Content-Type", "text/csv")
    writer := csv.NewWriter(w)
    
    // Stream in batches
    offset := 0
    batchSize := 1000
    
    for {
        students, err := store.GetStudentsList(offset, batchSize)
        if len(students) == 0 {
            break
        }
        
        for _, s := range students {
            writer.Write([]string{s.Name, s.Email})
        }
        writer.Flush()  // Stream to client
        
        offset += batchSize
    }
}
```

---

## Comparison Table

| Strategy | Performance | Consistency | UX | Complexity |
|----------|-------------|-------------|-----|-----------|
| Offset-based | ⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐ |
| Cursor-based | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐ |
| No pagination | ⭐ | ⭐⭐⭐⭐⭐ | ⭐ | ⭐ |

---

## Best Practices

1. **Always set a maximum limit** (we use 100)
2. **Provide sensible defaults** (we use page=1, limit=20)
3. **Return metadata** (total count, has_next, etc.)
4. **Use database indexes** on sort columns
5. **Consider caching** for expensive counts
6. **Monitor query performance** (slow query logs)
7. **Document your pagination** in API docs
8. **Handle edge cases** (page > total_pages, negative values)

---

## Real-World Examples

### Twitter/X
- Cursor-based pagination
- No page numbers
- Infinite scroll

### GitHub
- Offset-based for repos list
- Cursor-based for activity feed
- Max 100 items per page

### Stripe API
- Cursor-based for all resources
- Consistent across all endpoints
- `starting_after` and `ending_before` cursors

### Google Search
- Offset-based with page numbers
- Max ~1000 results total (performance limit)

---

## Testing Your Pagination

```bash
# Test default behavior
curl http://localhost:8080/api/students

# Test pagination
curl "http://localhost:8080/api/students?page=2&limit=10"

# Test edge cases
curl "http://localhost:8080/api/students?page=99999"  # Empty page
curl "http://localhost:8080/api/students?limit=1000"  # Should cap at 100
curl "http://localhost:8080/api/students?page=-1"     # Should default to 1

# Test with load (use tools like Apache Bench)
ab -n 1000 -c 10 "http://localhost:8080/api/students?page=1&limit=20"
```

---

## Conclusion

Your implementation includes **offset-based pagination**, which is perfect for:
- ✅ Admin panels
- ✅ Small-medium datasets
- ✅ Situations where users need page numbers

For future enhancements, consider:
- Adding cursor-based pagination for large datasets
- Implementing filtering/search
- Adding caching for counts
- Creating database indexes

The key takeaway: **Never expose unpaginated endpoints in production!**