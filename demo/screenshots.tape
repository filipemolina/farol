# VHS tape for static README screenshots.
# Uses the same seeded throwaway store the demo GIF uses. Record from the
# repo root after seeding:
#
#   ./demo/seed.sh
#   vhs demo/screenshots.tape
#
# VHS Screenshot takes relative paths only, so every PNG lands under demo/.
# The recording starts only after the TUI window appears — all shell setup
# (env vars) happens with Show suppressed, so the shots are 100% app.

Set Shell "bash"
Set FontSize 15
Set Width 1200
Set Height 660
Set Theme "Dracula"
Set CursorBlink false
Set TypingSpeed 45ms

# Set up the throwaway XDG dirs so the TUI reads the seeded store.
# Hide the shell setup — we only want to capture the app.
Hide
Type "export XDG_DATA_HOME=/tmp/farol-demo/data XDG_CONFIG_HOME=/tmp/farol-demo/config && clear"
Enter
Sleep 800ms
Show

# Launch the TUI.
Type "/tmp/farol-demo/farol"
Enter
Sleep 3s

# --- Lists panel: open it so both panels are visible ------------------------
# L (shift+l) toggles the lists panel; the cursor stays on the task tree.
Type "L"
Sleep 600ms

# --- Main TUI: Lists + Tasks panels, pending/complete sections visible ------
Screenshot demo/screenshot-main.png

# --- Help overlay: shows the full keybinding catalog -----------------------
Type "?"
Sleep 1.2s
Screenshot demo/screenshot-help.png
Escape
Sleep 600ms

# Close the lists panel so the task tree fills the width. Closing it also
# hands focus back to the tree, so n below opens the new-TASK input rather
# than the lists panel's new-LIST modal.
Type "L"
Sleep 500ms

# --- Add input: n starts creating, ] makes the level a child ---------------
Type "n"
Sleep 500ms
Type "Add rate limiting"
Sleep 400ms
Type "]"
Sleep 500ms
Screenshot demo/screenshot-add.png
Escape
Sleep 500ms
Escape
Sleep 500ms

# Re-open the lists panel so the remaining shots carry it in frame.
Type "L"
Sleep 500ms

# --- Theme picker: live preview --------------------------------------------
# T opens the picker. seed.sh pins farol-dark, so the picker opens on it and
# the screenshot is the live preview of the palette every other still is
# already in. Pressing Down would move the preview to a different theme,
# which would contradict the rest of the README.
Type "T"
Sleep 1500ms
Screenshot demo/screenshot-theme.png
Escape
Sleep 600ms

# --- Global search picker --------------------------------------------------
# "timeout" is the query the seed is built for: "Raise the gateway timeout"
# (infra) matches on title and ranks above "Cut the p95 on /search" (api),
# which matches on notes only. One shot, two lists, and the documented
# ranking rule. The old "trellis" query matched a task the tape types but
# never saves, so this shot shipped reading "No matches".
Type "F"
Sleep 800ms
Type "timeout"
Sleep 1.5s
Screenshot demo/screenshot-search.png
Escape
Sleep 500ms

# --- Archive page: the seeded archived list + its task preview -------------
# 2 opens the Archived Lists page, the top nav's second tab. seed.sh archives
# "billing-migration" ahead of time, so this captures a real list + preview,
# not an empty state.
Type "2"
Sleep 1200ms
Screenshot demo/screenshot-archive.png
Escape
Sleep 500ms
# 1 returns to the Active page for the last shot.
Type "1"
Sleep 800ms

# --- Complete section: watch the cascade move items to the Complete section --
# The cursor is on "Ship auth v2" (first root, seeded with a 3-level subtree).
# space completes it and every descendant at once; they move to the Complete
# section below.
#
# This shot comes last because it is the only destructive one, and completing
# a root does NOT un-cascade: a following Space un-completes the root but
# leaves its descendants complete. Run before the other shots, it left them
# all photographing a pending parent over three completed children.
Type "L"
Sleep 500ms
Sleep 300ms
Space
Sleep 1500ms
Screenshot demo/screenshot-complete.png
Sleep 500ms

Ctrl+C
Sleep 600ms
