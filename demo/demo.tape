# VHS tape for the farol README demo.
# Install VHS first: https://github.com/charmbracelet/vhs (needs ttyd + ffmpeg too).
#
# seed.sh compiles ./farol into /tmp/farol-demo/farol and seeds the
# throwaway store there, so the recording never touches the recorder's real
# data. Re-record from the repo root with:
#
#   make demo
#
# That runs seed.sh, records, and compresses via demo/compress.sh (mpdecimate
# for duplicate frames, 10fps, 900px wide, 48-colour palette — same recipe
# stack-stitcher uses).
#
# The recording starts the instant the TUI window appears — all shell setup
# (build, seed, env vars) happens with Show suppressed, so the GIF is 100%
# TUI, and it ends the same way: q is pressed under a fresh Hide so the last
# frame is the app, not the shell prompt that follows it. The demo walks the
# v view-mode cycle on the seeded list, opens the lists panel, creates a
# list, adds a root task and a nested subtask (one level deep via ]),
# previews the theme picker, then opens the Archived Lists page (2) to show
# the seeded archived list and its task preview. seed.sh pins farol-dark, the
# brand navy/amber palette the landing page is built from, so the whole
# walkthrough runs in it.

Output demo/demo.gif

Set Shell "bash"
Set FontSize 15
Set Width 1200
Set Height 660
Set Theme "Dracula"
# A blinking cursor changes pixels twice a second; every tick is a frame the
# GIF has to carry. Holding it still saves megabytes on a recording this long.
Set CursorBlink false
# 150ms per character is a comfortable human typist (~80 wpm). The old 45ms
# read as pasted text. This default governs every VISIBLE typed string — the
# list name and the two task titles — which is the only typing the GIF shows.
# Two kinds of line keep an @1ms override: single-key walkthrough steps (a
# lone keypress is instant for a person too) and the hidden setup commands
# above the Show, where slow typing would only pad the render.
Set TypingSpeed 150ms

# Build the binary and seed the throwaway store. seed.sh compiles ./farol into
# /tmp/farol-demo/farol itself, so this is self-contained from a clean
# checkout. Hide every frame of setup — the GIF starts at the TUI on the
# seeded "Home" list.
Hide
Type@1ms "cd /home/filipe/Documents/projects/farol && ./demo/seed.sh && clear"
Enter
Sleep 3s
# Export the same throwaway XDG dirs inside the visible shell so the TUI reads
# them (also hidden — we only want to show the app).
Type@1ms "export XDG_DATA_HOME=/tmp/farol-demo/data XDG_CONFIG_HOME=/tmp/farol-demo/config && clear"
Enter
Sleep 800ms

# Launch the TUI. Show only after it has had time to render.
Type@1ms "/tmp/farol-demo/farol"
Enter
Sleep 3s
Show

# --- View-mode cycle: v -> Pending only, v -> Complete only, v -> All ------
# The seeded "api" list already has both pending and complete tasks (Drop
# the legacy /v1 routes cascaded to Complete in seed.sh), so filtering here
# is visibly meaningful before the walkthrough moves on to lists/creation.
Type@1ms "v"
Sleep 1200ms
Type@1ms "v"
Sleep 1200ms
Type@1ms "v"
Sleep 1200ms

# --- Open the lists side panel ----------------------------------------------
# L (shift+l) toggles the lists panel. Opening it moves focus onto it, so the
# next keystrokes act on the lists, not the task tree.
Type@1ms "L"
Sleep 800ms

# --- Navigate up and down once in the lists panel ---------------------------
Up
Sleep 600ms
Down
Sleep 800ms

# --- Create a new list -------------------------------------------------------
# n opens the new-list name modal; type the name, Enter submits. The app then
# selects the new list AND auto-closes the lists panel (ListCreatedMsg calls
# closeListsPanel), so no separate "L to close" is needed — pressing L again
# would re-open it. Focus lands on the now-empty task tree.
Type@1ms "n"
Sleep 700ms
Type "server-maint"
Sleep 600ms
Enter
Sleep 1500ms

# --- Add a root task ---------------------------------------------------------
# The new list is empty, so its inline create input auto-opened and is live —
# no n is needed for the first task (n would just type the letter into the
# title). Type the name, Enter saves it. Rapid entry keeps a fresh, focused
# input anchored on the just-created task, so the next task can be indented
# straight away.
Type "Rotate the TLS certs"
Sleep 500ms
Enter
Sleep 1200ms

# --- Add a nested subtask (one level deep) -----------------------------------
# The input is still live and anchored on "Rotate the TLS certs". ]
# (Tree.Indent) sets the create level to +1, so the new task becomes a child
# of "Rotate the TLS certs".
Type@1ms "]"
Sleep 800ms
Type "Reload nginx on the edge nodes"
Sleep 500ms
Enter
Sleep 1500ms

# --- Close the create input --------------------------------------------------
# Esc discards the still-live rapid-entry input and returns the keyboard to
# the tree so the theme picker can open.
Escape
Sleep 800ms

# --- Theme picker: live preview ----------------------------------------------
# T opens the picker. seed.sh pins farol-dark, so the picker opens on it and
# holding here previews the palette the rest of the recording is already in.
# Walking the cursor would re-paint the app in unrelated themes, breaking
# that. Esc discards the preview without persisting.
Type@1ms "T"
Sleep 1500ms
Escape
Sleep 2000ms

# --- Archive page: browse the archived list + its task preview, then back --
# 2 opens the Archived Lists page, the top nav's second tab
# (src/keys/Keys.go). seed.sh archives "billing-migration" ahead of time, so
# this shows a real list + preview, not an empty state. 1 returns to Active.
Type@1ms "2"
Sleep 1500ms
Type@1ms "1"
Sleep 1200ms

# --- Quit ---------------------------------------------------------------------
# q exits to the shell, so Hide is re-engaged first: the recording ends on the
# TUI's last frame instead of the prompt that follows the exit.
Hide
Type@1ms "q"
Sleep 500ms
