# Task 2: Config constants + SchaleDB client

## Completed
- Created config.go with all constants per plan spec
- Created schaledb.go with FetchManifest() that downloads JSON and caches to ~/.cache/student-picker/students.json
- Build passes successfully

## Key Decisions
- Made CacheDir() a function instead of const to properly expand HOME env var
- FetchManifest() returns []byte (raw JSON) per task requirement - parsing deferred to Task 3
- Used context.WithTimeout(30s) for HTTP request per pattern

## Files Created
- config.go: Constants and CacheDir() function
- schaledb.go: FetchManifest() with HTTP client

## Verification
- go build succeeded
- Files in place: config.go, schaledb.go