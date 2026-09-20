# MCP UAT runbook — YouTube Uploader

Use this document when you want a full end-to-end **user acceptance** check of the **YouTube Uploader MCP** against real YouTube (not CI mocks).

**How to run:** Attach or `@`-reference this file and say: **“Run MCP UAT”** or **“Run MCP UAT from uat/MCP_UAT.md.”**

The agent must execute all steps in order, record pass/fail, and produce the **UAT report** at the end.

---

## Agent rules (mandatory — read first)

The agent acts as an **external MCP user**, not as a developer on this repository.

**Allowed**

- Read **only** this runbook (`uat/MCP_UAT.md`) and the fixture paths listed in **Constants** (all under `uat/`).
- Call **YouTube Uploader MCP tools** (`channels`, `authenticate`, `accesstoken`, `refreshtoken`, `upload_video`, `verify_upload`, `list_videos`, `update_video`, and related tools exposed by the configured MCP server).
- Ask the **human user** for OAuth codes, channel confirmation, or environment fixes when a step fails.

**Forbidden**

- **No repository access** beyond the `uat/` folder — do not open, search, or cite source code, tests, `CONTRIBUTING.md`, `README.md`, or any path outside `uat/`.
- **No shell or terminal** — do not run `ls`, `ffprobe`, `ffmpeg`, scripts, or any other commands.
- **No writing** — do not create or edit files, generate subtitle files, or patch the codebase; use the fixtures and paths already provided in **Constants**.
- **No implementation knowledge** — do not infer behavior by reading Go/MCP server code; rely on this runbook and MCP tool responses only.

If a step appears to require a forbidden action, mark it **FAIL** or **BLOCKED**, note why, and ask the user to fix the environment or fixtures—not the codebase.

---

## Constants (do not guess)

| Item | Value |
|------|--------|
| **Target channel (YouTube)** | Relaxing Routes — handle `@relaxingroutesyoutube` |
| **Google account label** | **“Random”** (user’s name for this Google login) |
| **Channel ID** | `UCIyTe9SBulUwvGgAMA3cARQ` |
| **Small video path** | `/Users/itclear/Projects/mcp/youtube-uploader-mcp/uat/uat_small.mp4` |
| **Large video path** | `/Users/itclear/Projects/mcp/youtube-uploader-mcp/uat/uat_large.mp4` |
| **Custom thumbnail (Step 5)** | `/Users/itclear/Projects/mcp/youtube-uploader-mcp/uat/uat_small_thumbnail.jpg` (1280×720 JPEG, &lt;2MB) |
| **Subtitles fixture (Step 8)** | `/Users/itclear/Projects/mcp/youtube-uploader-mcp/uat/uat_small.srt` |
| **MCP namespace (Cursor)** | `user-youtube-uploader-mcp-v0.1.3.alpha` (or current configured name; discover via client, not repo files) |

Use **only** the thumbnail path above for Step 5 — do not substitute paths from elsewhere in the repo.

### Title timestamp format

Use **local IST** in YouTube upload titles:

```text
Test 001 2026-09-20 11:40 IST
Test 002 2026-09-20 11:45 IST
Test 003 2026-09-20 12:30 IST
```

Pattern: `Test NNN YYYY-MM-DD HH:mm IST` (24-hour clock, space before `IST`).

---

## Prerequisites

- YouTube Uploader MCP status **ready** in the client (tools listed in **Agent rules**).
- Human maintainer has placed UAT fixtures under `uat/` (`uat_small.mp4`, `uat_large.mp4`, `uat_small_thumbnail.jpg`, `uat_small.srt`) at the absolute paths in **Constants**.
- MCP `upload_video` can read those paths (Step 3 must not return `video file path does not exist` or `operation not permitted`).

---

## Step 1 — Connect Relaxing Routes (“Random”)

**Goal:** Valid OAuth for the correct channel before any upload.

1. Call **`channels`**.
2. Confirm an entry with:
   - `name`: **Relaxing Routes**
   - `id`: `UCIyTe9SBulUwvGgAMA3cARQ`
   - Valid token (`access_token`, `refresh_token`, non-expired or refreshable).
3. If missing or token bad: **ask the user** to connect **Relaxing Routes** using the Google account they call **“Random”**:
   - Run **`authenticate`**, user opens URL, completes consent.
   - User pastes the **`code`** from the redirect; run **`accesstoken`** with that code only.
   - Call **`channels`** again and confirm Relaxing Routes.

**Pass:** Relaxing Routes channel ID present and usable (optional: **`refreshtoken`** succeeds if near expiry).

**Fail:** User must complete OAuth; do not upload until Step 1 passes.

---

## Step 2 — Verify MCP file access

**Goal:** Confirm the MCP server process can read both UAT video paths (external users have no shell access).

Use MCP only — **do not** run terminal commands.

1. Proceed to Step 3 with **`upload_video`** on the **small** fixture, **or**
2. If you must check access without uploading, ask the user to confirm fixtures exist at the paths in **Constants**.

**Pass:** Step 3 **`upload_video`** gets past filesystem errors (`video file path does not exist`, `operation not permitted`) and reaches YouTube (success or API/auth error, not local open failure).

**Fail:** Filesystem denied — ask the user to fix paths, permissions, or MCP host access; do not inspect the codebase.

---

## Step 3 — Upload small video (private)

**Goal:** First test upload with generated metadata.

Call **`upload_video`**:

| Parameter | Value |
|-----------|--------|
| `channel_id` | `UCIyTe9SBulUwvGgAMA3cARQ` |
| `file_path` | small video path (above) |
| `title` | `Test 001 <time>` (see format) |
| `description` | Short test text (include run id / timestamp) |
| `tags` | Comma-separated tags, e.g. `mcp-uat,uat,relaxing-routes` |
| `category_id` | `19` (Travel & Events) or `22` (People & Blogs) |
| `status` | `private` |
| `made_for_kids` | `false` |
| `video_language` | `hi` (Hindi) — optional but set for this run |

Save returned **`id`** / `video_id` as **`SMALL_VIDEO_ID`** for later steps.

If response indicates upload still **running** (large files only; small may finish inline), use Step 10 pattern with **`verify_upload`**.

**Pass:** Success JSON with YouTube video id.

---

## Step 4 — Verify upload via `list_videos`

Call **`list_videos`**:

- `channel_id`: Relaxing Routes
- `query`: `Test 001` (or full title substring)
- `privacy_status`: `private`
- `max`: enough to include the new video

**Pass:** Entry matches **`SMALL_VIDEO_ID`**, title contains `Test 001`, `privacy_status` is `private`, description/tags reflected as returned by API (title + description at minimum).

**Note:** `list_videos` returns `id`, `title`, `description`, `published_at`, `privacy_status`, `thumbnail_url`. It does **not** return language or made-for-kids; those are verified via upload response / `update_video` result / Studio if needed.

---

## Step 5 — First metadata update + re-verify

**Goal:** Exercise **`update_video`** on the small test video.

Use **at least two** of:

1. **`thumbnail_path`**: `/Users/itclear/Projects/mcp/youtube-uploader-mcp/uat/uat_small_thumbnail.jpg`
2. **`made_for_kids`**: `true` or `false` (opposite of upload default)
3. **`subtitle_path`**: skip here if Step 8 will cover subtitles

Call **`update_video`** with `channel_id`, `video_id` = **`SMALL_VIDEO_ID`**.

Re-run **`list_videos`** with `query` matching the current title.

**Pass:** `update_video` JSON shows `success` for requested fields; list still finds video; title/description unchanged unless API allows (see limitations).

---

## Step 6 — Second update: title `Test 002`, kids, language

**Intended test behavior:**

- Rename title to **`Test 002 <time>`**
- Set **`made_for_kids`** as appropriate
- Confirm **language** (Hindi) on the video

**Current MCP behavior (important):**

| Field | Set on upload | `update_video` |
|--------|----------------|----------------|
| Title / description / tags | Yes | **No** |
| `made_for_kids` | Yes | **Yes** |
| `video_language` | Yes | **No** (upload-only) |
| Subtitles | — | **Yes** (`subtitle_path`) |
| Thumbnail | — | **Yes** |
| Privacy / schedule | Yes | Partial (`publish_at`, not arbitrary privacy) |

**Agent procedure:**

1. Call **`update_video`** with `made_for_kids` (and any other supported fields).
2. For **title `Test 002`**: if the server cannot update snippet metadata, mark step **BLOCKED — needs MCP enhancement** and document in the report. Do **not** silently skip without noting the gap.
3. Re-run **`list_videos`** with `query`: `Test 002` or `Test 001` depending on what the API shows.

**Pass (full):** List shows `Test 002`, kids flag updated per `update_video` result.  
**Pass (partial, documented):** Kids update succeeds; title/language remain as set at upload until the MCP product supports title/snippet updates (document from tool responses and this runbook only).

---

## Step 7 — Verify Step 6 changes

- **`list_videos`**: locate video by id **`SMALL_VIDEO_ID`**; confirm title/description/privacy as API returns.
- Record **`update_video`** JSON for `made_for_kids_status`.

**Pass:** Consistent with Step 6 outcome and limitations table above.

---

## Step 8 — Subtitles for small video

1. Use the existing subtitle fixture at **`uat_small.srt`** (full path in **Constants**). Do **not** create or edit the file.

2. Call **`update_video`**:
   - `subtitle_path`: path from **Constants**
   - `subtitle_language`: `hi`

**Pass:** `subtitles_status`: `success` in tool JSON.

---

## Step 9 — Verify subtitle update

- Confirm **`update_video`** success payload from Step 8.
- **`list_videos`** will **not** list caption tracks; note verification is via update result (optional: user checks YouTube Studio).

**Pass:** Subtitles reported success; no error in `errors` map.

---

## Step 10 — Large video upload + poll (>30s)

**Goal:** Prove background upload + **`verify_upload`**.

Call **`upload_video`**:

| Parameter | Value |
|-----------|--------|
| `file_path` | large video path |
| `channel_id` | Relaxing Routes |
| `title` | `Test 003 <time>` |
| `description` / `tags` / `category_id` | Test metadata (private run) |
| `status` | `private` |

The server returns early after **~30 seconds** if upload continues; message should mention **`verify_upload`**.

Then poll **`verify_upload`** with the **same** `channel_id` and `file_path` until:

- **success** (contains video id → save as **`LARGE_VIDEO_ID`**), or
- **failed** (report error), or
- timeout (10 minutes hard limit in server — report failure).

**Pass:** Poll completes with success and a valid video id.

---

## Step 11 — Verify large upload

**`list_videos`** with `query`: `Test 003`, `privacy_status`: `private`.

**Pass:** **`LARGE_VIDEO_ID`** present with expected title and private status.

---

## Step 12 — Ensure all test uploads are private

For every id collected in this run (**`SMALL_VIDEO_ID`**, **`LARGE_VIDEO_ID`**, and any extra test ids):

1. Call **`list_videos`** (filter by title prefix `Test 00` or by known ids).
2. Confirm each `privacy_status` is **`private`**.

If any video is not private:

- **`update_video` does not support setting privacy to private directly** today (unless using schedule semantics on private drafts).
- Report **BLOCKED** and list ids that need manual Studio fix or future MCP support.

**Pass:** All run videos show `private` in `list_videos`, or gaps explicitly documented.

---

## UAT report template (agent fills in)

```markdown
## MCP UAT — YYYY-MM-DD HH:mm IST

| Step | Description | Result | Notes |
|------|-------------|--------|-------|
| 1 | Relaxing Routes / Random OAuth | PASS/FAIL | |
| 2 | MCP file access (via Step 3) | PASS/FAIL | |
| 3 | Upload small Test 001 | PASS/FAIL | SMALL_VIDEO_ID= |
| 4 | list_videos verify 001 | PASS/FAIL | |
| 5 | update_video (round 1) | PASS/FAIL | |
| 6 | update Test 002 / kids / language | PASS/FAIL/BLOCKED | |
| 7 | Verify step 6 | PASS/FAIL | |
| 8 | Subtitles | PASS/FAIL | |
| 9 | Verify subtitles | PASS/FAIL | |
| 10 | Large upload + verify_upload poll | PASS/FAIL | LARGE_VIDEO_ID= |
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

To auto-load this runbook in Cursor, add a project rule: when the user asks for **MCP UAT**, read and follow **`uat/MCP_UAT.md` only** and obey **Agent rules** (external user; no repo/code/shell).

---

## Reference — fixture hints (for humans maintaining `uat/`)

Approximate sizes (maintainers only; the UAT agent does not verify these with shell commands):

| File | Duration | Size |
|------|----------|------|
| `uat_small.mp4` | ~14 min | ~84 MB |
| `uat_large.mp4` | ~54 min | ~1.4 GB |
| `uat_small_thumbnail.jpg` | — | ~90 KB (1280×720 JPEG) |

Fixtures must stay at the paths in **Constants** so the MCP server can read them on macOS.
