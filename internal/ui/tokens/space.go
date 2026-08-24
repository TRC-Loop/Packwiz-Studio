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

	// ActivityBarWidth is the icon rail on the far left.
	ActivityBarWidth float32 = 44
	// ActivityIconSlot is the square hit target of one rail icon.
	ActivityIconSlot float32 = 40

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
	// LauncherWidth and LauncherHeight size the launcher window.
	LauncherWidth  float32 = 720
	LauncherHeight float32 = 460

	// PackWindowWidth and PackWindowHeight size a new pack window.
	PackWindowWidth  float32 = 1100
	PackWindowHeight float32 = 700
)
