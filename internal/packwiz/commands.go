package packwiz

import (
	"context"

	"github.com/PalisadeMC/Packwiz-Studio/internal/cmdrun"
)

// The subcommand names below use packwiz's canonical spellings rather
// than its aliases ("modrinth", not "mr"), so the log drawer reads as
// something a user could paste into a terminal without knowing the
// shorthand.

// Refresh rebuilds the pack index. packwiz needs this after any manual
// edit to the pack folder, so the app runs it after raw TOML edits and
// before an export.
func (c *Client) Refresh(ctx context.Context) (cmdrun.Result, error) {
	return c.run(ctx, "refresh")
}

// AddModrinth adds a Modrinth project to the pack. The ref must be
// unambiguous: a project ID or an exact slug from the browser, never a
// free-text search, because packwiz resolves an ambiguous search
// interactively and the app runs non-interactively.
func (c *Client) AddModrinth(ctx context.Context, ref string) (cmdrun.Result, error) {
	return c.run(ctx, "modrinth", "add", ref)
}

// Remove deletes a mod's metadata file and refreshes the index. The name
// is the mod's metadata name as packwiz knows it.
func (c *Client) Remove(ctx context.Context, name string) (cmdrun.Result, error) {
	return c.run(ctx, "remove", name)
}

// Update updates one mod to its newest allowed version.
func (c *Client) Update(ctx context.Context, name string) (cmdrun.Result, error) {
	return c.run(ctx, "update", name)
}

// UpdateAll updates every mod in the pack. This backs the manual
// "Check for Updates" action. The app never runs it on a timer.
func (c *Client) UpdateAll(ctx context.Context) (cmdrun.Result, error) {
	return c.run(ctx, "update", "--all")
}

// Pin freezes a mod at its current version so updates skip it.
func (c *Client) Pin(ctx context.Context, name string) (cmdrun.Result, error) {
	return c.run(ctx, "pin", name)
}

// Unpin lets a pinned mod receive updates again.
func (c *Client) Unpin(ctx context.Context, name string) (cmdrun.Result, error) {
	return c.run(ctx, "unpin", name)
}

// List asks packwiz for the pack's mods. The app parses the pack's TOML
// directly for anything structured; this is for showing the user what
// packwiz itself thinks is installed.
func (c *Client) List(ctx context.Context) (cmdrun.Result, error) {
	return c.run(ctx, "list")
}

// ExportModrinth writes a .mrpack. An empty output path lets packwiz pick
// the name, but the app always supplies one: the export dialog resolves a
// concrete path first so the result can be attached to a release.
func (c *Client) ExportModrinth(ctx context.Context, output string) (cmdrun.Result, error) {
	args := []string{"modrinth", "export"}
	if output != "" {
		args = append(args, "--output", output)
	}
	return c.run(ctx, args...)
}
