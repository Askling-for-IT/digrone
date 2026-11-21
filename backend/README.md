# Backend

Go API using chi, pgx connection pooling, and PostgreSQL. Its structure follows
`twogo`: configuration → router → handlers → store → database.

## Run locally

Requires Go 1.25.5 or newer and Docker Compose. From `backend/`:

```sh
docker compose up -d --wait
cp .env.example .env
# Edit .env with your Keycloak realm URL and client secret before continuing.
set -a
. ./.env
set +a
go run ./cmd/api
```

The local `.env` configures the API on port **8083**, leaving port **8080** for
the frontend and **8081** for your existing local Keycloak container.
The application reads process environment variables; it does not load `.env`
automatically. The example file contains local development credentials only.
The PostgreSQL container is available on `localhost:5435`, so it can coexist
with a PostgreSQL server installed on your computer.

| Variable | Required | Default | Purpose |
| --- | --- | --- | --- |
| `DATABASE_URL` | Yes | — | PostgreSQL connection URL, including database and TLS options |
| `PORT` | No | `8081` | API listening port; local `.env` sets `8083` |
| `VERSION` | No | `0.0.1` | Version exposed by `/version` |
| `APP_URL` | No | `http://localhost:8080` | External React origin, without a trailing slash; HTTPS enables Secure cookies |
| `OIDC_ISSUER` | Yes | — | Keycloak realm URL: `https://your-keycloak-host/realms/digrone` |
| `OIDC_CLIENT_ID` | No | `digrone` | Confidential Keycloak client |
| `OIDC_CLIENT_SECRET` | Yes | — | Client secret from Keycloak; keep in `.env` |
| `OIDC_REDIRECT_URL` | No | `APP_URL` + `/auth/callback` | Login callback registered in Keycloak |
| `OIDC_ADMIN_ROLE` | No | `admin` | Required role on the configured client |

Database data persists in the Compose volume. Stop the database while keeping data:

```sh
docker compose down
```

To dispose of this development database **and permanently delete its data**:

```sh
docker compose down --volumes
```

You can instead use an existing PostgreSQL server by creating a dedicated database
and configuring `DATABASE_URL` with its credentials, port, and TLS settings.
Do not use the local credentials or `sslmode=disable` for a remote production
connection; configure the database's required TLS settings.

## Endpoints

```sh
curl http://localhost:8083/health
# {"status":"ok"}
curl http://localhost:8083/db-health
# {"status":"db ok"}
curl http://localhost:8083/version
# {"version":"0.0.1"}
```

`/health` is a process liveness check. `/db-health` checks PostgreSQL with a
2-second timeout and returns HTTP 503 when unavailable. Startup fails if
`DATABASE_URL` is missing, PostgreSQL cannot be reached, or migrations fail.
Port and connection options are passed directly to the HTTP server and database
driver; there is no separate configuration validation layer.
Keycloak discovery must also succeed at startup. SIGINT/SIGTERM trigger graceful
HTTP shutdown before the pool closes.

## Migrations

SQL files in `internal/db/migrations/` are embedded in the API binary and applied
on startup with `golang-migrate`, as in twogo. Its `schema_migrations` table tracks
the version and dirty state; starting the API again skips applied migrations.
Application tables will use PostgreSQL's existing `public` schema, like twogo.
The initial migration records a baseline without creating application tables.
`000002_sessions` adds application sessions. `000003_players_matches_events`
adds players, matches, and statistics/fine events. `000004_event_match_links`
enforces match references for goals/penalties and forbids them for fines/payments.

Add sequential pairs such as `000005_description.up.sql` and
`000005_description.down.sql`. Never edit an already applied migration. Use
transactions for changes that PostgreSQL allows inside a transaction. The API
only migrates forward; down files are available for deliberate manual rollback
with the migration CLI. Back up data before production schema changes.

A failed migration prevents startup. Inspect and fix the database/migration state
before retrying; the API does not automatically force a version or roll back data.
The database user needs permission to create the migration tracking table and
subsequent application objects in `public`.

## Layout

- `cmd/api`: startup, HTTP server, shutdown
- `internal/config`: environment configuration
- `internal/http`: routing and HTTP middleware
- `internal/http/handlers`: HTTP request handlers
- `internal/store`: shared database access; domain queries will be added here
- `internal/db`: migration runner and embedded SQL migrations

## Players, matches, and events

- `players`: a UUID and name for each person. Player statistics are calculated
  from events rather than stored on this table.
- `matches`: kickoff time with timezone, opponent name, home/away flag,
  competition label, status (`scheduled`, `completed`, `cancelled`), and final
  goals for/against. Scores are empty until a result is recorded. Competition
  labels such as `Season 26/27` and `Power Cup 2027` are stored directly on matches.
- `events`: player, optional match, event type, optional assisting player,
  penalty minutes, number of beer cans, comment, and occurrence timestamp.

Event types and their intended fields:

| Type | Fields | Meaning |
| --- | --- | --- |
| `goal` | `match_id`, `player_id`, optional `assisting_player_id` | One goal for the scorer and one assist if specified |
| `penalty` | `match_id`, `player_id`, `penalty_minutes` | Penalty minutes for a player |
| `fine` | `player_id`, positive `num_beers`, optional `comment`, no `match_id` | Beer cans owed and the reason |
| `fine_payment` | `player_id`, positive `num_beers`, optional `comment`, no `match_id` | Beer cans paid, including partial payments |

Outstanding beer balance is the sum of fines minus the sum of payments. Total
paid is the sum of payments. Payments are separate records so paying a fine does
not erase its history. Fines and payments are always independent of matches.
Beer balances use all fine/payment events, regardless of match status. Sporting
totals will use goal and penalty events from completed matches only. Final match
scores are stored independently so results can be recorded without known scorers.

Foreign keys preserve player and match references. Deleting a referenced player
or match requires handling its events explicitly; events are not silently deleted.
No match rosters, appearance counts, team IDs, shirt numbers, or captain fields
are stored. Historical JSON data has not been imported. Statistics aggregation
queries are subsequent work.

## Player API

| Method | Path | Access | Result |
| --- | --- | --- | --- |
| GET | `/api/players` | Public | Player list ordered by name; an empty list is `[]` |
| GET | `/api/players/{id}` | Public | One player |
| POST | `/api/players` | Admin session | Create a player; HTTP 201 with the player and a Location header |
| PUT | `/api/players/{id}` | Admin session | Update a player's name; return the updated player |
| DELETE | `/api/players/{id}` | Admin session | Delete an unreferenced player; HTTP 204 |

Create and update requests contain only the name:

```json
{"name":"Kim Askling"}
```

Player responses contain `id` (a database-generated UUID) and `name`. Missing
players return HTTP 404. Deletion returns HTTP 409 if the player is referenced
as a scorer, assisting player, or recipient of a penalty/fine/payment; their
event history remains intact. Writes use the same admin sessions and cross-origin
protection as the authentication endpoints.

For local browser requests, use `/api/players` through the React development
server on port 8080. The Vite proxy forwards requests to the API on port 8083.

## Match API

| Method | Path | Access | Result |
| --- | --- | --- | --- |
| GET | `/api/matches` | Public | Match summaries ordered by start time |
| GET | `/api/matches/{id}` | Public | Match details including its events |
| POST | `/api/matches` | Admin session | Create a match and optional events; HTTP 201 |
| PUT | `/api/matches/{id}` | Admin session | Update a match and optional events; HTTP 200 |

Filter the list using `?competition=Season%2026%2F27`. Empty lists are `[]`.
There is no match deletion endpoint; deletion is reserved for manual database
maintenance. Existing event foreign keys still apply to manual deletions.

Match save requests contain the complete match fields and may include events:

```json
{
  "start_time": "2026-11-22T13:00:00+01:00",
  "opponent": "Opponent name",
  "home_game": true,
  "competition": "Season 26/27",
  "status": "completed",
  "goals_for": 2,
  "goals_against": 1,
  "events": [
    {
      "event_type": "goal",
      "player_id": "<scorer-player-uuid>",
      "assisting_player_id": "<assisting-player-uuid>"
    },
    {
      "event_type": "penalty",
      "player_id": "<player-uuid>",
      "penalty_minutes": 2
    }
  ]
}
```

Replace the placeholders with existing player UUIDs. Status defaults to
`scheduled`; use `completed` explicitly when saving a final result, and provide
both final scores. Scheduled scores may be `null`. Start times use RFC 3339,
including a timezone offset or `Z`.

The match result and supplied events are saved in **one transaction**. If any
player reference or event fails, the entire save rolls back, including the
previous result and events on updates.

- Supplying `events` replaces the match's complete event list, giving the new
  records fresh event IDs. Use the returned IDs for subsequent individual edits.
- Supplying `events: []` clears the match's events.
- Omitting `events` (or setting it to `null`) preserves existing events on updates.
- Nested events must be goals or penalties; their match ID is assigned by the
  backend. Fines and payments are recorded through the separate event endpoints.

PUT replaces the match fields; it is not a partial update. Scores are independent
of event counts so a result can be saved before all scorers are known. Save
responses include the match and its current `events` array. Single-match reads
use one database snapshot for a consistent result/event response.

## Event API

| Method | Path | Access | Result |
| --- | --- | --- | --- |
| GET | `/api/events` | Public | Event list ordered by occurrence time |
| GET | `/api/events/{id}` | Public | One event |
| POST | `/api/events` | Admin session | Create an event; HTTP 201 |
| PUT | `/api/events/{id}` | Admin session | Update an event; HTTP 200 |
| DELETE | `/api/events/{id}` | Admin session | Delete an event; HTTP 204 |

List filters can be combined: `match_id`, `player_id` (the primary player, not
an assisting player), and `event_type`. Example:
`/api/events?player_id=<uuid>&event_type=fine`.

Goal and penalty requests require `match_id`. Fines and payments must omit it
or send `null`. Only goals accept `assisting_player_id`; penalties require positive
`penalty_minutes`; fines/payments require positive `num_beers`. Other type-specific
fields should be omitted or `null`. The database also enforces the match-link rule.

Example standalone fine:

```json
{
  "event_type": "fine",
  "player_id": "<player-uuid>",
  "num_beers": 6,
  "comment": "Late to training"
}
```

Use `fine_payment` to record a payment with the same fields. `comment` is optional.
Optional `occurred_at` accepts an RFC 3339 timestamp; creation defaults to the
current time and updates preserve the existing timestamp if omitted.
PUT otherwise replaces the event fields. Missing records return HTTP 404;
unknown referenced players or matches return HTTP 409. All writes require an
admin session and are protected against cross-origin browser requests.

## Keycloak setup

Use your existing Keycloak server; the application does not run another instance.
Create the realm **`digrone`**, then:

1. Create an OpenID Connect client with client ID **`digrone`**. Enable **Client
   authentication** and **Standard flow**. Leave implicit flow, direct access
   grants, and service accounts disabled.
2. Register **`http://localhost:8080/auth/callback`** as the valid redirect URI
   for local development. For deployment, register the same path under your
   HTTPS application domain. The URI must match `OIDC_REDIRECT_URL` exactly.
   Set PKCE code challenge method to **S256** if configuring it explicitly.
3. Copy the client secret from the client's **Credentials** tab into
   `backend/.env`. Set `OIDC_ISSUER` to your server URL ending in
   **`/realms/digrone`**.
4. Under the client's **Roles**, create **`admin`**. Assign that **client role**
   to each admin user under **Users → Role mapping → Assign role** (filter by
   clients). A Keycloak realm administrator does not automatically have this role.
5. Include this client's roles in the **ID token**. Under **Client scopes →
   digrone-dedicated → Mappers**, add a **User Client Role** mapper. Set its
   client ID to `digrone`, token claim name to `resource_access.digrone.roles`,
   claim JSON type to `String`, **Multivalued** on, and **Add to ID token** on.

The backend reads the client role from the verified ID token. Keycloak's default
access-token role mapper alone is insufficient. If you change the client ID,
update the mapper's client ID and claim name too.

## Authentication endpoints

| Method | Path | Behavior |
| --- | --- | --- |
| GET | `/auth/login` | Redirect to Keycloak using Authorization Code + PKCE |
| GET | `/auth/callback` | Verify login and admin role, create session, redirect to `/admin` |
| GET | `/auth/me` | Return the current admin's subject, username, email, admin flag, and session expiry; HTTP 401 without a valid session |
| POST | `/auth/logout` | Delete this application's session and clear cookies; HTTP 204 |

Sessions last **8 hours** without sliding renewal or background token refresh.
Only a hash of the random session token is stored in PostgreSQL, along with
identity and expiry information. Passwords and Keycloak tokens are not stored.
Cookies use HttpOnly and SameSite=Lax; `APP_URL=https://...` also enables Secure.
Expired sessions are rejected immediately and cleaned up on subsequent logins.

Logout ends the digröne session only; the browser may still be signed into
Keycloak and other applications. Role changes in Keycloak take effect on the
next login. To revoke an existing application's session immediately, delete the
user's rows from `public.sessions`.

`RequireAdmin` middleware protects `/auth/me` and all player/match/event write routes. It returns
HTTP 401 for missing/expired sessions and HTTP 403 for sessions without admin
access.

Go's built-in cross-origin protection guards requests that change data. Browser
POST/PUT/PATCH/DELETE requests must come from the same origin. The Vite development
server proxies `/auth` and `/api` to Go, so use **http://localhost:8080/auth/login**
for local browser login. Once logged in, React can use relative requests:

```js
const me = await fetch('/auth/me').then(response => response.json());
await fetch('/auth/logout', { method: 'POST' });
```

Production should serve React and proxy `/auth` and `/api` to Go under the same
HTTPS domain, preserving the original Host header. Set `APP_URL` and the Keycloak
redirect URI to that domain. No cross-origin cookie or CORS configuration is
needed with this layout.

The React admin screen still uses its previous login/storage flow. Wiring that
screen to these endpoints belongs to the frontend integration step.

## Verify

```sh
go vet ./...
go build -o bin/api ./cmd/api
```

Authentication was verified using a temporary local OIDC provider issuing signed
ID tokens and a disposable PostgreSQL database. Checks covered admin login,
PKCE, state/nonce validation, token signature/issuer/audience/expiry, non-admin
rejection, session persistence and expiry, logout, and cross-origin POST rejection.
Player endpoints were verified against local PostgreSQL with temporary records:
public reads, admin-only create/update/delete, persisted names, missing players,
and protection against deleting players referenced by events. Temporary checks
ran with the race detector; their records and test files were removed afterward.
Match/event CRUD was verified in a separate temporary PostgreSQL database using
race-enabled temporary checks. These covered all write authorization, cross-origin
rejection, combined-save rollback, event replacement/preservation, standalone fines
and payments, filters, and migration rollback/reapply. The temporary database and
test file were removed; the application database was not modified.
No test files are kept in the repository. A real-server login still requires the
Keycloak realm and client configuration above.

Statistics endpoints, historical import, frontend API integration, and server
deployment are subsequent steps. The frontend continues using its existing storage.
