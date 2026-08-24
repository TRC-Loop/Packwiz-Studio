package tokens

// Spacing scale. Every gap, pad and inset in the UI comes from here.
// The scale is compact by design: long mod lists are the common case.
const (
	SpaceXS  float32 = 2
	SpaceSM  float32 = 4
	SpaceMD  float32 = 8
	SpaceLG  float32 = 12
	SpaceXL  float32 = 16
	SpaceXXL float32 = 24
)

// Corner radii.
const (
	// RadiusControl applies to buttons, inputs and select boxes.
	RadiusControl float32 = 6
	// RadiusCard applies to mod cards and launcher recents entries.
	RadiusCard float32 = 6
	// RadiusOverlay applies to dialogs, popups and menus.
	RadiusOverlay float32 = 6
	// RadiusPane is zero: structural panes stay square.
	RadiusPane float32 = 0
)

// Component metrics for the pack window shell.
const (
	// RowHeight is one entry in a list: mod, changed file, release.
	RowHeight float32 = 24

	// ActivityBarWidth is the icon rail on the far left. It is wide enough
	// for a section's name to sit under its glyph.
	ActivityBarWidth float32 = 72
	// ActivityIconSlot is the height of one rail entry, glyph plus name.
	ActivityIconSlot float32 = 52

	// SidePanelWidth is the default width of the list pane.
	SidePanelWidth float32 = 220
	// SidePanelMinWidth is how far the pane may be dragged closed.
	SidePanelMinWidth float32 = 160

	// StatusBarHeight is the always-visible bar at the window foot.
	StatusBarHeight float32 = 24
	// LogDrawerHeight is the drawer's height when first opened.
	LogDrawerHeight float32 = 160
	// LogDrawerMinHeight is the smallest useful drawer height.
	LogDrawerMinHeight float32 = 64

	// PaneInset is the padding between a pane's edge and its content.
	PaneInset float32 = 8

	// HairlineThickness is the width of dividers and separators.
	HairlineThickness float32 = 1
	// FocusRingThickness is the width of the keyboard focus outline.
	FocusRingThickness float32 = 1
)

// Window sizes.
const (
	// LauncherWidth and LauncherHeight size the launcher window. It is
	// roomy enough for the new pack form to open inside it: a dialog
	// taller than its window is clipped, buttons and all.
	LauncherWidth  float32 = 900
	LauncherHeight float32 = 620

	// PackWindowWidth and PackWindowHeight size a new pack window.
	PackWindowWidth  float32 = 1100
	PackWindowHeight float32 = 700

	// SettingsWidth and SettingsHeight size the settings dialog.
	SettingsWidth  float32 = 560
	SettingsHeight float32 = 480

	// FormWidth and FormHeight size the new pack dialog.
	FormWidth  float32 = 520
	FormHeight float32 = 460

	// ModCardWidth and ModCardHeight size one card in the browser grid.
	ModCardWidth  float32 = 240
	ModCardHeight float32 = 148

	// DialogChrome is the vertical room a dialog spends on its own title
	// row and button row, which scrolling content has to leave free.
	DialogChrome float32 = 96

	// ExportHeight sizes the export dialog.
	ExportHeight float32 = 340

	// CloneHeight sizes the clone dialog.
	CloneHeight float32 = 400

	// PropertiesHeight sizes the pack properties dialog.
	PropertiesHeight float32 = 480

	// AddModHeight sizes the manual add mod dialogs.
	AddModHeight float32 = 420

	// ReleaseWidth and ReleaseHeight size the release dialog, which holds
	// a changelog field and so needs more room than a plain form.
	ReleaseWidth  float32 = 640
	ReleaseHeight float32 = 620
)
