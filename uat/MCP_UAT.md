# MCP UAT runbook — YouTube Uploader

Use this document when you want a full end-to-end **user acceptance** check of the **YouTube Uploader MCP** against real YouTube (not CI mocks).

**How to run:** Attach or `@`-reference this file and say: **“Run MCP UAT”**. Before Step 1, the **human user** must provide the values in **Run configuration** (below).

The agent must execute all steps in order, record pass/fail, and produce the **UAT report** at the end.

---

## Agent rules (mandatory — read first)

The agent acts as an **external MCP user**, not as a developer on this repository.

**Allowed**

- Read **only** this runbook (`uat/MCP_UAT.md`) and paths derived from **Run configuration** (all under `{REPO_ROOT}/uat/`).
- Call **YouTube Uploader MCP tools** (`channels`, `authenticate`, `accesstoken`, `refreshtoken`, `upload_video`, `check_job_status`, `list_videos`, `update_video`, and any other tools the client lists for the YouTube Uploader server).
- Ask the **human user** for OAuth codes, `REPO_ROOT`, `YOUR_CHANNEL_ID`, channel confirmation, or environment fixes when a step fails.

**Forbidden**

- **No repository access** beyond resolving paths under `{REPO_ROOT}/uat/` — do not open, search, or cite source code, tests, or docs outside `uat/`.
- **No shell or terminal** — do not run `ls`, `ffprobe`, `ffmpeg`, scripts, or any other commands.
- **No writing** — do not create or edit files; use the fixtures already under `uat/`.
- **No implementation knowledge** — do not infer behavior by reading server source; rely on this runbook and MCP tool responses only.

If a step appears to require a forbidden action, mark it **FAIL** or **BLOCKED**, note why, and ask the user to fix the environment or fixtures—not the codebase.

---

## Run configuration (human provides before UAT)

These values are **not** stored in the public repo. The user states them in chat (or keeps a private local note).

| Symbol | Meaning | Example shape (not real data) |
|--------|---------|-------------------------------|
| **`REPO_ROOT`** | Absolute path to **this repository** on the machine where the MCP server runs | `/home/you/dev/youtube-uploader-mcp` |
| **`YOUR_CHANNEL_ID`** | YouTube channel ID to use for all uploads and listings | `UCxxxxxxxxxxxxxxxxxxxxx` |
| **`YOUR_CHANNEL_NAME`** | (Optional) Expected `name` from **`channels`** for sanity-check only | `My UAT Test Channel` |
| **`YOUR_VIDEO_LANG`** | ISO 639-1 language for `video_language` on upload | `en` |
| **`YOUR_SUBTITLE_LANG`** | ISO 639-1 language for subtitle track in Step 8 | `en` |
| **`YOUR_TZ_LABEL`** | Short label for upload titles (your local timezone) | `PST`, `IST`, `UTC` |

### Derived fixture paths (always under `uat/`)

Replace `{REPO_ROOT}` with the user’s path:

| Fixture | Path |
|---------|------|
| Small video | `{REPO_ROOT}/uat/uat_small.mp4` |
| Large video | `{REPO_ROOT}/uat/uat_large.mp4` |
| Thumbnail (Step 5) | `{REPO_ROOT}/uat/uat_small_thumbnail.jpg` |
| Subtitles (Step 8) | `{REPO_ROOT}/uat/uat_small.srt` |

Maintainer: generate `uat_small_thumbnail.jpg` locally (1280×720 JPEG, &lt;2MB); it is gitignored. MP4s are gitignored too.

### Title timestamp format

Use the user’s local time in YouTube upload titles:

```text
Test 001 YYYY-MM-DD HH:mm YOUR_TZ_LABEL
Test 002 YYYY-MM-DD HH:mm YOUR_TZ_LABEL
Test 003 YYYY-MM-DD HH:mm YOUR_TZ_LABEL
```

Pattern: `Test NNN YYYY-MM-DD HH:mm <YOUR_TZ_LABEL>` (24-hour clock).

---

## Prerequisites

- YouTube Uploader MCP **ready** in the client (includes `check_job_status` for long uploads).
- User has supplied **`REPO_ROOT`** and **`YOUR_CHANNEL_ID`**.
- Fixtures exist at the derived paths above on the MCP host.
- Step 3 must not return `video file path does not exist` or `operation not permitted`.

---

## Step 1 — Connect UAT channel (OAuth)

**Goal:** Valid OAuth for **`YOUR_CHANNEL_ID`** before any upload.

1. Call **`channels`**.
2. Confirm an entry whose **`id`** equals **`YOUR_CHANNEL_ID`** (and optionally whose **`name`** matches **`YOUR_CHANNEL_NAME`** if the user gave one).
3. Confirm a usable token (`access_token`, `refresh_token`, non-expired or refreshable via **`refreshtoken`**).
4. If missing or bad: **ask the user** to OAuth the Google account that **owns** that channel:
   - **`authenticate`** → user opens URL and consents.
   - User pastes only the **`code`** from the redirect → **`accesstoken`**.
   - **`channels`** again; confirm **`YOUR_CHANNEL_ID`**.

**Pass:** **`YOUR_CHANNEL_ID`** present and usable.

**Fail:** User must complete OAuth; do not upload until Step 1 passes.

---

## Step 2 — Verify MCP file access

**Goal:** MCP server can read UAT video paths.

Use MCP only — **do not** run terminal commands.

Proceed to Step 3, or ask the user to confirm fixtures exist at the derived paths.

**Pass:** Step 3 **`upload_video`** passes filesystem checks (reaches YouTube, not `does not exist` / `operation not permitted`).

---

## Step 3 — Upload small video (private)

Call **`upload_video`**:

| Parameter | Value |
|-----------|--------|
| `channel_id` | **`YOUR_CHANNEL_ID`** |
| `file_path` | `{REPO_ROOT}/uat/uat_small.mp4` |
| `title` | `Test 001 <time>` (see format) |
| `description` | Short test text (include run timestamp) |
| `tags` | e.g. `mcp-uat,uat,youtube-uploader-mcp` |
| `category_id` | `19` or `22` (or another valid YouTube category ID) |
| `status` | `private` |
| `made_for_kids` | `false` |
| `video_language` | **`YOUR_VIDEO_LANG`** |

Save returned **`id`** as **`SMALL_VIDEO_ID`**.

Small uploads may finish inline; large uploads use Step 10.

**Pass:** Success JSON with YouTube video id.

---

## Step 4 — Verify upload via `list_videos`

- `channel_id`: **`YOUR_CHANNEL_ID`**
- `query`: `Test 001`
- `privacy_status`: `private`

**Pass:** Entry matches **`SMALL_VIDEO_ID`**, title contains `Test 001`, `privacy_status` is `private`.

**Note:** `list_videos` does not return language or made-for-kids; use upload / `update_video` responses.

---

## Step 5 — First metadata update + re-verify

Use **at least two** of:

1. **`thumbnail_path`**: `{REPO_ROOT}/uat/uat_small_thumbnail.jpg`
2. **`made_for_kids`**: opposite of Step 3 default
3. **`subtitle_path`**: skip if Step 8 will cover subtitles

**Pass:** Requested fields show success in `update_video` JSON (thumbnail may **fail** with YouTube **403** if the channel lacks custom-thumbnail eligibility — record **PARTIAL**, not MCP failure).

Re-run **`list_videos`** with `query` matching the current title.

---

## Step 6 — Second update: title `Test 002`, kids, language

**Intended:** Rename to **`Test 002 <time>`**, adjust **`made_for_kids`**, confirm upload language.

**Current MCP behavior (from tool schema + responses only):**

| Field | On `upload_video` | On `update_video` |
|--------|-------------------|-------------------|
| Title / description / tags | Yes | **No** (unless tool adds snippet fields) |
| `made_for_kids` | Yes | **Yes** |
| `video_language` | Yes | **No** (upload-only today) |
| Subtitles | — | **Yes** |
| Thumbnail | — | **Yes** |
| Privacy / schedule | Yes | Partial (`publish_at`) |

1. **`update_video`** with `made_for_kids` as needed.
2. Title **`Test 002`**: if not supported, mark **BLOCKED — snippet update not exposed** and document.
3. **`list_videos`** with `query` `Test 002` or `Test 001`.

**Pass (full):** List shows `Test 002`. **Pass (partial):** Kids update OK; title unchanged — document gap.

---

## Step 7 — Verify Step 6

**`list_videos`** by **`SMALL_VIDEO_ID`**; record **`made_for_kids_status`** from Step 6.

---

## Step 8 — Subtitles

1. Use **`{REPO_ROOT}/uat/uat_small.srt`** (do not create or edit).
2. **`update_video`**: `subtitle_path` as above, `subtitle_language`: **`YOUR_SUBTITLE_LANG`**.

**Pass:** `subtitles_status`: `success`.

---

## Step 9 — Verify subtitles

Confirm Step 8 JSON; captions are not listed by **`list_videos`**.

---

## Step 10 — Large video upload + poll (>30s)

**Goal:** Background upload + **`check_job_status`**.

**`upload_video`**:

| Parameter | Value |
|-----------|--------|
| `file_path` | `{REPO_ROOT}/uat/uat_large.mp4` |
| `channel_id` | **`YOUR_CHANNEL_ID`** |
| `title` | `Test 003 <time>` |
| `description` / `tags` / `category_id` | Same style as Step 3 |
| `status` | `private` |

If the response includes **`status`: `running`** and a **`job_key`**, poll **`check_job_status`** with that **`job_key`** until success, failed, or ~10 minute timeout.

**Pass:** Success payload with video id → **`LARGE_VIDEO_ID`**.

**Note:** Older server builds used `verify_upload` with `channel_id` + `file_path`; if the client only offers **`check_job_status`**, use **`job_key`** only.

---

## Step 11 — Verify large upload

**`list_videos`**: `query` `Test 003`, `privacy_status` `private`, channel **`YOUR_CHANNEL_ID`**.

---

## Step 12 — All test uploads private

**`list_videos`** for titles matching `Test 00` on **`YOUR_CHANNEL_ID`**; each **`privacy_status`** must be **`private`**.

If not private and MCP cannot set privacy via **`update_video`**, report **BLOCKED** and list video ids.

---

## UAT report template

```markdown
## MCP UAT — YYYY-MM-DD HH:mm YOUR_TZ_LABEL

| Step | Description | Result | Notes |
|------|-------------|--------|-------|
| 1 | UAT channel OAuth | PASS/FAIL | YOUR_CHANNEL_ID= |
| 2 | MCP file access (via Step 3) | PASS/FAIL | |
| 3 | Upload small Test 001 | PASS/FAIL | SMALL_VIDEO_ID= |
| 4 | list_videos verify 001 | PASS/FAIL | |
| 5 | update_video (round 1) | PASS/FAIL/PARTIAL | |
| 6 | update Test 002 / kids / language | PASS/FAIL/BLOCKED | |
| 7 | Verify step 6 | PASS/FAIL | |
| 8 | Subtitles | PASS/FAIL | |
| 9 | Verify subtitles | PASS/FAIL | |
| 10 | Large upload + check_job_status | PASS/FAIL | LARGE_VIDEO_ID= |
| 11 | list_videos Test 003 | PASS/FAIL | |
| 12 | All private | PASS/FAIL/BLOCKED | |

### Video IDs this run
- Small:
- Large:

### MCP / product gaps found
-
```

---

## Optional: Cursor rule

When the user asks for **MCP UAT**, read **`uat/MCP_UAT.md` only**, obey **Agent rules**, and ask for **`REPO_ROOT`** + **`YOUR_CHANNEL_ID`** if not provided.

---

## Reference — fixture hints (maintainers)

| File | Typical duration | Typical size |
|------|------------------|--------------|
| `uat_small.mp4` | ~14 min | ~84 MB |
| `uat_large.mp4` | ~54 min | ~1.4 GB |
| `uat_small_thumbnail.jpg` | — | ~90 KB (1280×720 JPEG) |

Keep binaries gitignored; commit this runbook and `uat_small.srt` only.
