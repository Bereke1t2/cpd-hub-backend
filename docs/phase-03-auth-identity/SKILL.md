# Phase 3 — Auth & Identity

**Goal:** Untangle the username/email muddle, load the authenticated user into every protected request, and
round out auth with `GET /api/auth/me`, token refresh, and input validation. After this phase the rest of
the app can rely on "who is calling" being available in the gin context.

**Depends on:** Phases 1, 2.
**Risk:** Medium–High. Touches login/signup and every protected route (adds `loadUser`). Keep the token's
`username` claim authoritative so the client (which stores `username` and builds `/users/profile/:username`)
keeps working.

---

## Checklist
- [x] 3.1 Decide the identity model; signup takes an explicit `username` (fallback: derive from email).
- [x] 3.2 Login by email **or** username; store email in its own column.
- [x] 3.3 `loadUser` middleware → puts the current `*domain.UserProfile` in the context.
- [x] 3.4 `GET /api/auth/me` returns the current user.
- [x] 3.5 Access + refresh tokens; `POST /api/auth/refresh`.
- [x] 3.6 Request validation (struct tags + a binding helper) on auth payloads.
- [x] 3.7 Consistent 401s; never leak whether an email exists.

---

## 3.1 Identity model
Today `users.username` stores the email and login does `WHERE username = req.Email`. Phase 2 added a real
`email` column. Now make username a first-class handle:

- **Signup**: accept optional `username`. If absent, derive a unique slug from `fullName`/email local-part
  (`alice`, `alice2`, …). Store `email` in `email`, the handle in `username`.
- The JWT `username` claim is the handle. The client already reads it and uses it for profile routes — don't
  break that.

> Migration note: existing rows have email-in-username. Backfill: `UPDATE users SET email = username WHERE
> email IS NULL;` then optionally generate handles. Ship this as `migrations/000N_backfill_email.up.sql`.

## 3.2 Login
Accept either an email or a handle in the `email` field the client sends (it only has one field). Query:
```sql
SELECT username, full_name, password_hash FROM users WHERE email = $1 OR username = $1
```
Compare bcrypt, issue tokens. See [`auth_usecase.go`](./auth_usecase.go) for the full flow (it replaces the
logic currently inline in `postgres/auth_repo.go`, moving rules into a usecase and leaving the repo as data
access only).

## 3.3 loadUser middleware
The router already accepts a `loadUser gin.HandlerFunc` but `NewHandler` passes `nil`. Implement it — copy
[`middleware_auth.go`](./middleware_auth.go) to `internal/infrastructure/security/load_user.go`. It reads the
validated claims (set by `AuthMiddleware`), loads the profile via the Auth/Profile repo, and stores it:
```go
c.Set("user", profile)          // *domain.UserProfile
c.Set("username", profile.Username)
```
Add a tiny accessor used by every handler:
```go
func currentUsername(c *gin.Context) string { v, _ := c.Get("username"); s, _ := v.(string); return s }
```
Wire it in `NewHandler`:
```go
RegisterRoutes(g, h, security.AuthMiddleware(), security.LoadUser(repos.Profile))
```

## 3.4 GET /api/auth/me
Add the route (protected) and handler — returns the profile loaded by `loadUser`. The client can call this
right after login to hydrate the session instead of guessing the username. See `auth_usecase.go` + the
handler snippet in the SKILL's §3 wiring.

```go
authProtected := api.Group("/auth"); authProtected.Use(auth, loadUser)
authProtected.GET("/me", h.Me)
authProtected.POST("/refresh", h.Refresh)
```

## 3.5 Refresh tokens
Issue a short-lived **access** token (24h) + a longer **refresh** token (e.g. 30d). `POST /api/auth/refresh`
with a valid refresh token mints a new access token. Keep it simple: sign refresh tokens with the same secret
but a `typ:"refresh"` claim; reject refresh tokens on protected routes and access tokens on `/refresh`. See
[`jwt.go`](./jwt.go) for `GenerateToken`/`GenerateRefreshToken`/`ParseToken` with the `typ` guard.

> If you don't need refresh yet, ship just `/me` and a single 24h token — but design the claim now so adding
> refresh later isn't a breaking change.

## 3.6 Validation
Add binding tags to the request structs and validate on bind:
```go
type SignupRequest struct {
	FullName        string `json:"fullName" binding:"required,min=2"`
	Username        string `json:"username" binding:"omitempty,alphanum,min=3,max=20"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" binding:"required,eqfield=Password"`
}
```
`c.ShouldBindJSON(&req)` now returns a descriptive error; map it to `domain.ErrValidation(...)`. See
[`validation.go`](./validation.go) for a helper that turns validator errors into a clean message.

## 3.7 Don't leak account existence
Login failures (no such user, wrong password) must return the **same** 401 message ("invalid credentials").
Signup on an existing email returns 409 conflict. Never echo the password back. Rate-limit auth in Phase 10.

---

## Definition of Done
- [x] Signup stores email + a distinct handle; login works with either.
- [x] Every protected handler can call `currentUsername(c)` and get the caller's handle.
- [x] `GET /api/auth/me` returns the current profile; client can hydrate session from it.
- [x] `POST /api/auth/refresh` mints a new access token from a refresh token; cross-type use is rejected.
- [x] Invalid signup/login payloads return 400 with a clear message (not a generic "bad json").
- [x] Wrong credentials and unknown account both return the same 401.
