# PurePlay × CourseForge — Multi-Tier Builder Platform Use Case

**Status:** target use case (production vision)  
**Target customer:** **PurePlay**  
**Related:** [Courseforge × KBL integration](courseforge-integration.md) · [ADR 0036](../adr/0036-courseforge-integration-exploration.md)

## Summary

PurePlay course designers use **CourseForge builder tools** to turn design bundles into playable golf courses. Access is gated by **login and team membership**. Compute runs on a shared **Blender + Unreal worker pool** scheduled by the CourseForge orchestrator and, in hosted mode, **KBL + Volcano** for queue-aware batch scheduling.

The platform serves three overlapping audiences:

| Audience | What they need | Service expectation |
|----------|----------------|---------------------|
| **PurePlay design teams** | Reliable builds during working sessions | Predictable turnaround; optional **speed priority** |
| **Developer accounts** | API keys, CI hooks, custom integrations | Tiered quotas; burst or dedicated capacity |
| **Community pool** | Broader access to builder tools for learning and experimentation | **Low SLA** — long queue waits acceptable |

One React GUI and one FastAPI backend ([MILESTONE-A6](https://github.com/courseforge/course-builder/blob/main/milestones/specs/MILESTONE-A6.md)) serve all modes; differences are **identity, tenant policy, and compute queue tier** — not separate products.

---

## Actors and teams

```mermaid
flowchart TB
  subgraph pureplay [PurePlay organization]
    TL[Team lead]
    D1[Designer — East region]
    D2[Designer — West region]
    DEV[Developer account]
  end

  subgraph access [Access layer]
    IDP[Cognito / SSO login]
    RBAC[Roles and team scopes]
  end

  subgraph product [CourseForge]
    GUI[Builder GUI]
    API[FastAPI backend]
    ORCH[Orchestrator]
  end

  subgraph compute [Compute fabric]
    Q1[Volcano queue — team priority]
    Q2[Volcano queue — team standard]
    Q3[Volcano queue — community pool]
    BW[Blender worker]
    UW[Unreal worker]
  end

  TL --> IDP
  D1 --> IDP
  D2 --> IDP
  DEV --> IDP
  IDP --> RBAC
  RBAC --> GUI
  RBAC --> API
  GUI --> API
  API --> ORCH
  ORCH --> Q1
  ORCH --> Q2
  ORCH --> Q3
  Q1 --> BW
  Q1 --> UW
  Q2 --> BW
  Q2 --> UW
  Q3 --> BW
  Q3 --> UW
```

### Team model

| Concept | Meaning for PurePlay |
|---------|---------------------|
| **Organization tenant** | `PurePlay` — top-level billing, data segregation, org-wide policy |
| **Design team** | e.g. `pureplay-east`, `pureplay-prototype` — shared artifact prefix, shared quota pool |
| **Team lead** | Manages members, approves community publication, can purchase priority credits |
| **Designer** | Submits builds, views team artifacts; cannot cross team boundaries |
| **Developer account** | Machine identity (API key / OAuth client) for automation; separate quota and SLA tier |

Team membership is enforced at the API ([MILESTONE-15](https://github.com/courseforge/course-builder/blob/main/milestones/specs/MILESTONE-15.md) tenancy model): every job, artifact URI, and orchestrator `courseId` carries a **tenant id** and **team id** partition key.

---

## Identity: login screen and user management (AWS Cognito)

PurePlay hosted builder access requires a **login screen**, **session management**, and **user administration** before designers reach the CourseForge GUI. **AWS Cognito** is the reference identity provider — aligned with CourseForge [MILESTONE-15](https://github.com/courseforge/course-builder/blob/main/milestones/specs/MILESTONE-15.md) and the platform default for hosted/subscription operation.

### Why Cognito

| Requirement | Cognito feature |
|-------------|-----------------|
| Login screen for browser users | **Hosted UI** or **Amplify Auth** embedded in the React SPA |
| PurePlay org + team membership | **User Pool groups** + **custom attributes** (`custom:tenant_id`, `custom:team_id`) |
| Community vs internal users | Separate groups (`community`, `designer`, `team-lead`, `developer`) |
| Developer / CI accounts | **App client** with client credentials or machine-to-machine flow |
| Future enterprise SSO | **SAML / OIDC federation** into the same User Pool |
| API authorization | **JWT access tokens** validated by FastAPI on every request |

Local desktop (Tauri) and `run.sh` dev mode remain **unauthenticated or license-gated**; Cognito applies to **hosted cloud** and **community Kind** deployments where multi-tenant policy is enforced.

### Cognito resources (PurePlay production)

```mermaid
flowchart TB
  subgraph users [Users]
    PP[PurePlay designers]
    CM[Community users]
    DEV[Developer CI clients]
  end

  subgraph cognito [AWS Cognito]
    UP[User Pool<br/>pureplay-courseforge]
    HC[Hosted UI / OAuth]
    GR[Groups and custom attributes]
    AC[App clients<br/>spa + api + m2m]
  end

  subgraph app [CourseForge hosted]
    LOGIN[Login screen]
    SPA[React builder SPA]
    API[FastAPI backend]
  end

  PP --> LOGIN
  CM --> LOGIN
  LOGIN --> HC
  HC --> UP
  UP --> GR
  GR --> SPA
  SPA -->|Bearer JWT| API
  DEV --> AC
  AC -->|client credentials| API
  API -->|validate JWT| UP
```

| Resource | Purpose |
|----------|---------|
| **User Pool** | `pureplay-courseforge` (or shared CourseForge pool with tenant attribute) |
| **App client — SPA** | Public client with PKCE for `tools/courseforge/frontend` |
| **App client — API** | Optional confidential client for server-side token exchange |
| **App client — M2M** | Developer accounts; client id + secret → access token with `developer` scope |
| **Hosted UI domain** | `auth.build.pureplay.example.com` — branded login, signup, forgot password |
| **Custom attributes** | `tenant_id`, `team_id`, `service_tier` (mutable by admins only) |
| **Groups** | `pureplay-designer`, `pureplay-team-lead`, `pureplay-developer`, `community` |

Optional **Cognito Identity Pool** if the SPA needs direct AWS credentials (e.g. short-lived S3 upload from browser). Prefer **presigned URLs from the API** when possible so the backend remains the tenancy gatekeeper.

### Login screen — user experience

The builder SPA shows a **dedicated login route** (`/login`) before any build tools load. Unauthenticated users are redirected; expired sessions show a re-login prompt with return URL preserved.

**Phase 1 (fastest): Cognito Hosted UI**

- Redirect to Cognito Hosted UI (PurePlay logo, colors via CSS customization).
- Supports email/password, MFA (optional TOTP), forgot password, email verification.
- On success, redirect back to SPA with authorization code; SPA exchanges for tokens (PKCE).
- **Pros:** no custom auth UI to maintain; MFA and compliance features built in.
- **Cons:** leaves CourseForge domain briefly during login.

**Phase 2 (optional): Embedded login in SPA**

- **Amplify Auth** or **oidc-client-ts** renders login form inside CourseForge chrome.
- Same User Pool; better brand continuity for PurePlay.
- Hosted UI remains fallback for password reset and MFA enrollment.

```mermaid
sequenceDiagram
  participant U as User
  participant SPA as CourseForge SPA
  participant CO as Cognito Hosted UI
  participant API as FastAPI backend

  U->>SPA: Open builder URL
  SPA->>SPA: No valid session → /login
  U->>SPA: Sign in
  SPA->>CO: OAuth authorize (PKCE)
  U->>CO: Email + password (+ MFA)
  CO-->>SPA: Authorization code
  SPA->>CO: Token exchange
  CO-->>SPA: ID + access + refresh tokens
  SPA->>SPA: Store tokens (memory / secure storage)
  SPA->>API: GET /api/me (Authorization Bearer)
  API->>API: Validate JWT, map claims → user context
  API-->>SPA: Profile, team, tier, quotas
  SPA-->>U: Builder home (team-scoped)
```

**Session behavior**

- Access token TTL: 1 hour (typical); refresh token: 30 days with rotation.
- SPA refreshes silently before expiry; on refresh failure → `/login`.
- Logout: revoke refresh token + clear local storage + Cognito global sign-out (optional).

### User management

User lifecycle spans **Cognito (identity)** and **CourseForge backend (authorization + quotas)**. Cognito owns credentials; the API owns team membership, entitlements, and audit.

| Action | Who performs it | Where |
|--------|-----------------|-------|
| Create PurePlay designer | Team lead or org admin | Admin UI → API → `AdminCreateUser` / invite email |
| Invite to team | Team lead | API updates `custom:team_id` + group membership |
| Remove from team | Team lead | API removes group; may disable Cognito user if leaving org |
| Community self-signup | End user | Hosted UI signup → auto `community` group + `tenant_id=community` |
| Promote to team lead | Org admin | API + Cognito group change |
| Developer API client | Org admin | Create M2M app client; map to service principal in API DB |
| Reset password / MFA | User | Hosted UI self-service |
| Disable account | Org admin | Cognito `AdminDisableUser` + API mark inactive |

**Admin surfaces (hosted GUI)**

1. **Org admin** (PurePlay IT) — list users, assign teams, set service tier, view usage.
2. **Team lead** — invite/remove designers on their team; cannot see other teams’ artifacts.
3. **Self-service profile** — name, password, MFA; read-only team and tier display.

The React SPA adds an **Account** / **Team settings** panel (M15 “hosted GUI policy surfaces”): show remaining quota, priority credits, and queue tier before submit.

**Provisioning flow — new PurePlay designer**

1. Team lead enters email in **Invite member** form.
2. API calls Cognito `AdminCreateUser` (invite message) with `custom:tenant_id=pureplay`, `custom:team_id=pureplay-east`.
3. API adds user to group `pureplay-designer`.
4. API inserts row in tenant DB (user id = `sub`, team, default tier `team-standard`).
5. User clicks invite link → sets password on Hosted UI → lands in builder.

**Community signup**

1. User opens community portal (`community.build.pureplay.example.com`).
2. Hosted UI **Sign up** enabled for this app client only.
3. Post-confirmation Lambda (or API webhook) assigns `community` group and `service_tier=community-pool`.
4. No team lead approval required; optional email domain blocklist for abuse.

### JWT claims and API enforcement

Every authenticated API call carries a **Bearer access token**. FastAPI middleware validates signature (Cognito JWKS), issuer, audience, and expiry, then builds request context:

| Claim | Maps to |
|-------|---------|
| `sub` | User id (stable) |
| `email` | Display + audit |
| `custom:tenant_id` | Storage partition (`pureplay`, `community`) |
| `custom:team_id` | Team RBAC + Volcano queue selection |
| `custom:service_tier` | `community-pool`, `team-standard`, `team-priority`, `developer-pro` |
| `cognito:groups` | Role checks (`team-lead` may invite) |

**Fail closed:** missing `tenant_id`, unknown team, or group/endpoint mismatch → `403` with no partial data leak. Job submit, artifact download, and trace endpoints all re-check tenant ([M15 segregation enforcement points](https://github.com/courseforge/course-builder/blob/main/milestones/specs/MILESTONE-15.md)).

Developer M2M tokens use a **service principal** record linked to an app client id; claims include `client_id` and scoped `tenant_id` / `team_id` — no human `email`.

### Developer accounts (separate from human login)

Developer accounts do not use the login screen. They authenticate with **client credentials**:

```http
POST https://auth.build.pureplay.example.com/oauth2/token
Content-Type: application/x-www-form-urlencoded

grant_type=client_credentials&client_id=...&client_secret=...&scope=build:submit build:read
```

The API maps the client to a **developer-pro** tier queue and webhook configuration. Org admins rotate secrets via admin UI; old secrets revoked in Cognito.

### PurePlay + community on one pool vs split pools

| Approach | When to use |
|----------|-------------|
| **Single User Pool**, tenant attribute distinguishes PurePlay vs community | Recommended start — one Hosted UI, simpler ops |
| **Separate User Pools** | Hard regulatory isolation between PurePlay employees and public community |

PurePlay production likely starts with **one pool** and group-based separation; split only if compliance requires it.

### Implementation checklist (CourseForge + infrastructure)

- [ ] Cognito User Pool + SPA app client (PKCE) in `courseforge/infrastructure` CDK/Terraform
- [ ] Hosted UI branding (PurePlay logo, callback URLs for staging + prod)
- [ ] Custom attributes and groups defined in IaC
- [ ] FastAPI JWT middleware + `/api/me`, `/api/admin/users`, `/api/teams/{id}/members`
- [ ] React `/login` route, auth guard on builder routes, token refresh
- [ ] Post-confirmation trigger for community tier assignment
- [ ] M2M app client for developer pro tier
- [ ] Audit log: actor `sub`, tenant, team, action ([M15 baseline](https://github.com/courseforge/course-builder/blob/main/milestones/specs/MILESTONE-15.md))

Local dev bypass: `AUTH_MODE=none` or static dev JWT — never enabled in hosted environments.

---

## Login and access to builder tools (summary)

Hosted PurePlay deployments gate the builder through **Cognito** (see previous section). Claims drive team isolation and queue tier:

| Claim / field | Use |
|---------------|-----|
| `sub` | Stable user id |
| `custom:tenant_id` | PurePlay org partition |
| `custom:team_id` | Design team for RBAC and artifact paths |
| `custom:service_tier` | Maps to Volcano queue / SLA tier |
| `cognito:groups` | Roles: `designer`, `team-lead`, `developer`, `community` |
| Scopes | API actions: `build:submit`, `build:read`, `artifact:download`, `dev:webhook` |

**Login surfaces:**

1. **Browser (hosted cloud / community Kind)** — Login screen → Cognito Hosted UI (or embedded Amplify) → CourseForge SPA.
2. **Pro desktop (optional)** — Tauri shell with CF1/CFS1 license gate for offline-capable local Blender; Cognito when online for cloud Unreal builds.
3. **Developer accounts** — OAuth2 client credentials; no interactive login screen.

Authorization **fails closed** when tenant or team claims are missing. Community users receive the `community` group only; they cannot read PurePlay team artifacts.

---

## Service tiers and SLA

PurePlay offers **different levels of service** on the same worker images. KBL **Volcano queues** (or orchestrator priority fields that map to them) implement the split.

| Tier | Who | Target turnaround | Queue behavior | SLA posture |
|------|-----|-------------------|----------------|-------------|
| **Community pool** | Public / invited community designers | Hours to overnight | Lowest `weight`, no preemption, shared cap | **Best-effort** — long waits acceptable; no uptime guarantee on results |
| **Team standard** | PurePlay designers (default) | Typical < 30 min in business hours | Per-team queue, fair share across members | Business-hours support; retry on worker failure |
| **Team priority** | Teams that opt in to speed | Typical < 5 min when capacity exists | Higher weight, optional preemption of community jobs | Paid or credit-based; SLA credits if exceeded |
| **Developer pro** | CI, partner integrations, batch tooling | Configurable (burst vs dedicated) | Dedicated queue or guaranteed minimum `capability` | Contractual; webhook on completion; audit log export |

### Community pool — low SLA by design

The **community pool** opens builder tools to a wider audience (students, partner clubs, open beta):

- Jobs may sit in queue **during peak PurePlay production hours** — this is expected.
- No SLA on queue depth or completion time; status UI shows position and estimated range only.
- Artifact retention is shorter (e.g. 7 days vs 90 days for team tiers).
- Community submissions do **not** consume team priority credits.

PurePlay production teams are isolated in **team-* queues** so community load cannot starve paid design work beyond configured fair-share caps.

### Speed priority (team and developer tiers)

Teams and developer accounts may **opt into priority scheduling**:

- Per-submit flag: `priority: standard | expedited`
- Expedited jobs route to `team-{id}-priority` Volcano queue or receive higher `priorityClassName`
- Billing: priority credits, monthly allotment, or pay-per-build
- KBL mechanism: separate `Queue` CR with higher `weight` and optional `preemptable: false` on team jobs ([ADR 0031](../adr/0031-computewheel-volcano-queue.md))

---

## Current course build pipeline (three steps)

Today the Alex / CourseForge automation path is a **three-step design → Blender → Unreal** flow. The orchestrator expresses steps 2–3 as versioned workflow templates in [`courseforge/course-builder`](https://github.com/courseforge/course-builder) (`tools/automation-workers/orchestrator/workflows/`).

```mermaid
flowchart LR
  S1["Step 1 — Design bundle<br/>A2 SVG + routing"]
  S2["Step 2 — Blender build<br/>A3 GCD recipe"]
  S3["Step 3 — Unreal import<br/>A4 FBX handoff"]
  S1 --> S2 --> S3
```

| Step | Milestone | Orchestrator stage | Worker | Output |
|------|-----------|-------------------|--------|--------|
| **1. Design bundle** | A2 | *(CourseForge GUI — not an orchestrator stage)* | — | `course.svg`, routing, optional terrain meta |
| **2. Blender build** | A3 | `blender-build` → package `blender-courseforge-build` | `blender-worker` | `course.blend`, `course.fbx` |
| **3. Unreal import** | A4 | `unreal-import` → package `unreal-import-fbx` | `unreal-worker` | Unreal assets under project path |

**Workflow templates today:**

- [`course-build-a3.yaml`](https://github.com/courseforge/course-builder/blob/main/tools/automation-workers/orchestrator/workflows/course-build-a3.yaml) — Blender only (single stage).
- [`course-build-a4.yaml`](https://github.com/courseforge/course-builder/blob/main/tools/automation-workers/orchestrator/workflows/course-build-a4.yaml) — Blender then Unreal with `dependsOn` and `inputFrom` FBX handoff.

The GUI and API treat the full path as one **“Build course”** action; designers see three logical phases in progress UI even when only two orchestrator stages run.

### Alex — planned Unreal expansion

Alex owns the GCD Blender add-on and Unreal tooling. **Additional Unreal stages** are planned after the initial FBX import slice (A4):

| Future stage (draft) | Purpose | Notes |
|---------------------|---------|-------|
| Landscape from heightmap | Terrain from A1 sidecars | `landscape_from_heightmap.py` scoped in A4 spec |
| Game logic placement | Tees, pins, hazards in-level | A5 hook; data plumbing in A4 |
| Packaging / cook | Playable build artifact | New job packages on stable `unreal-worker` image |

New stages append to the orchestrator DAG as **separate job packages** — heavy Unreal images are not rebuilt per workflow change ([stable worker spec](https://github.com/courseforge/infrastructure/blob/kind/courseforge-suite-2026-05/docs/stable-worker-job-package-pattern/stable-worker-spec.md)).

When KBL backs the pool, each stage may map to a **DominoChain** step with Volcano scheduling and snapshot handoff between Blender and Unreal dominos ([integration options](courseforge-integration.md#integration-options-in-increasing-invasiveness)).

---

## End-to-end request flow (hosted PurePlay)

```mermaid
sequenceDiagram
  participant U as Designer
  participant CF as CourseForge API
  participant O as Orchestrator
  participant V as Volcano / KBL
  participant B as blender-worker
  participant UE as unreal-worker

  U->>CF: Login (Cognito JWT)
  U->>CF: POST /api/courseforge/build (bundle + priority tier)
  CF->>CF: Enforce team quota and RBAC
  CF->>O: POST /course-jobs (template course-build-a4)
  O->>V: Schedule blender-build (queue = team or community)
  V->>B: POST /jobs (job package)
  B-->>O: stage complete + fbx artifact
  O->>V: Schedule unreal-import (dependsOn blender-build)
  V->>UE: POST /jobs (unreal-import-fbx)
  UE-->>O: stage complete
  O-->>CF: course job Succeeded
  CF-->>U: Build result + artifact links
```

**Developer account flow** is identical except authentication uses API key / client credentials, and jobs may specify `webhookUrl` for CI completion.

---

## Mapping tiers to KBL constructs

| Platform concept | KBL / Kubernetes artifact |
|------------------|---------------------------|
| Community pool | Volcano `Queue/community-pool` — low weight, large `capability` cap |
| Team standard | Volcano `Queue/team-{teamId}` |
| Team priority | Volcano `Queue/team-{teamId}-priority` or higher priority class |
| Developer pro | Dedicated `PluggableUniverse` or minimum guaranteed queue capability |
| Blender stage | `DominoChain` step → `blender-worker` `runnerImage` |
| Unreal stage | `DominoChain` step → `unreal-worker` `runnerImage` |
| Module release windows | Optional `ComputeWheel` time slice per PurePlay product line |
| Audit / regrade | Snapshot IDs + replay log per build |

See [Courseforge × KBL integration](courseforge-integration.md) for adapter and shim options between orchestrator HTTP dispatch and Workflow / DominoChain CRs.

---

## PurePlay deployment modes

| Mode | Shell | Compute | Typical tier |
|------|-------|---------|--------------|
| **Hosted cloud (EKS)** | Browser + Cognito | Orchestrator + KBL + Volcano on EKS | All tiers including community pool |
| **Community Kind** | Browser on lab ingress | Same stack at smaller scale | Community + team dev |
| **Pro desktop** | Tauri + local license | Local Blender; cloud optional for Unreal | Team standard locally; cloud for UE |
| **Developer CI** | API only | Webhook-driven orchestrator jobs | Developer pro |

PurePlay production is expected to run **hosted cloud** for team and community tiers, with **home i9 / Kind lab** ([lab/HOME-LAB.md](../../lab/HOME-LAB.md)) for integration testing before EKS rollout.

---

## Open decisions

1. **Priority pricing** — credits per expedited build vs monthly team allotment.
2. **Community eligibility** — open signup vs invite-only vs PurePlay-branded subdomain.
3. **Login UI** — Hosted UI only (phase 1) vs embedded Amplify form (phase 2).
4. **Cross-team sharing** — whether a finished course can be published from team A to community without re-build.
5. **GPU queue** — separate Volcano queue for GPU-heavy Unreal cooks when Alex adds them (i9 `kbl.io/gpu=present` label).
6. **KBL adapter timing** — orchestrator-native queues first vs full DominoChain mapping in Phase 32.

---

## References

- [courseforge-integration.md](courseforge-integration.md) — KBL scheduler behind CourseForge workers
- [ADR 0036: Courseforge integration exploration](../adr/0036-courseforge-integration-exploration.md)
- [ADR 0031: ComputeWheel Volcano queue](../adr/0031-computewheel-volcano-queue.md)
- CourseForge [MILESTONE-A3](https://github.com/courseforge/course-builder/blob/main/milestones/specs/MILESTONE-A3.md) (Blender / GCD)
- CourseForge [MILESTONE-A4](https://github.com/courseforge/course-builder/blob/main/milestones/specs/MILESTONE-A4.md) (Unreal import)
- CourseForge [MILESTONE-A6](https://github.com/courseforge/course-builder/blob/main/milestones/specs/MILESTONE-A6.md) (multi-mode distribution)
- CourseForge [MILESTONE-15](https://github.com/courseforge/course-builder/blob/main/milestones/specs/MILESTONE-15.md) (hosted auth, tenancy, quotas)
- [Stable worker job package pattern](https://github.com/courseforge/infrastructure/blob/kind/courseforge-suite-2026-05/docs/stable-worker-job-package-pattern/stable-worker-spec.md)
