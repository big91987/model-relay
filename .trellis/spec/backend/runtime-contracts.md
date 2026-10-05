# Runtime and persistence contracts

## Ownership and signatures
cmd/model-relay/main.go owns init, serve and reset-password operational entry points. internal/relay/server.go owns New(store *Store, config Config) *Server, ServeHTTP, reply and failure. internal/relay/store.go owns Init(directory, keyFile string) (string, error) and Store.DB; SQLite access uses database/sql and bound parameters, not an ORM. proxy.go owns outbound protocol handling; web assets are served by Go.

## Persistence and secrets
Follow store.go: separate key file, 0700 directory, 0600 key, exclusive installation lock, SQLite WAL, FULL synchronous and one connection. Admin passwords are salted Argon2 hashes; upstream credentials are encrypted; application keys use digests. Tests own temporary installations. Never edit user databases to pass tests; use README operational commands. No schema change belongs in a UI visibility feature.

## Input and failure boundary
server.go jsonInput requires application/json, a JSON object, bounded input, unique fields and DisallowUnknownFields. Example: failure(writer, http.StatusUnauthorized, "unauthorized", "") emits an error envelope with code/type/message and request_id; never include raw credentials or provider diagnostics. storageFailure logs only storage_write_failed and marks degraded state. UI api treats management 401 as signedOut; preserve that behavior.

## Health and version contract
/healthz supports GET and HEAD; HEAD has no body. Other methods are 405 with Allow: GET, HEAD. Healthy GET carries status and the raw Config.Version; degraded/closed database is unavailable. internal/relay/health_test.go covers handler and real HTTP method/state matrix; version_test.go covers raw version and real HTTP consistency. Do not bypass health-history checks or turn earlier receipts into current passes.

## Required verification
Makefile verify runs evidence, health-history, formatting, go mod verify, vet, Node UI tests, Go race/HTTP tests, build and real-service browser journeys. tests/browser/main.go starts isolated Go services and controlled upstreams; these are not real supplier evidence. Preserve failed attempts. Listening restrictions require host verification, not a weaker replacement gate. Avoid logging request/response bodies; tests/browser/journey.mjs audits logs and records for credential/conversation exclusion.
