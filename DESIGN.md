---
name: ResizeMe
description: A native, focused desktop utility for repeatable window sizing on Windows and macOS.
colors:
  accent: "#005FB8"
  accent-hover: "#005FB8E6"
  text-primary: "#000000E4"
  text-secondary: "#0000009B"
  text-tertiary: "#00000072"
  text-on-accent: "#FFFFFF"
  app-background: "#F3F3F3"
  layer: "#FFFFFF80"
  card: "#FFFFFFB3"
  divider: "#0000000F"
  control-stroke: "#00000029"
  favorite: "#D48B00"
  danger: "#C42B1C"
  mac-icon-blue-light: "#60A5FA"
  mac-icon-blue: "#3B6EEB"
  mac-icon-blue-deep: "#2240BE"
typography:
  headline:
    fontFamily: "Segoe UI Variable Text, Segoe UI, system-ui, sans-serif"
    fontSize: "20px"
    fontWeight: 600
    lineHeight: 1.4
  title:
    fontFamily: "Segoe UI Variable Text, Segoe UI, system-ui, sans-serif"
    fontSize: "14px"
    fontWeight: 600
    lineHeight: 1.4286
  body:
    fontFamily: "Segoe UI Variable Text, Segoe UI, system-ui, sans-serif"
    fontSize: "14px"
    fontWeight: 400
    lineHeight: 1.4286
  label:
    fontFamily: "Segoe UI Variable Text, Segoe UI, system-ui, sans-serif"
    fontSize: "12px"
    fontWeight: 400
    lineHeight: 1.3333
  icon-fluent:
    fontFamily: "Segoe Fluent Icons"
  icon-mdl2:
    fontFamily: "Segoe MDL2 Assets"
  mac-hero:
    fontFamily: "SF Pro Rounded, SF Pro Display, -apple-system, sans-serif"
    fontSize: "26px"
    fontWeight: 700
  mac-body:
    fontFamily: "SF Pro Text, -apple-system, sans-serif"
    fontSize: "13px"
    fontWeight: 400
  mac-caption:
    fontFamily: "SF Pro Text, -apple-system, sans-serif"
    fontSize: "11px"
    fontWeight: 400
rounded:
  app-icon: "3px"
  control: "4px"
  surface: "8px"
  toggle: "10px"
  mac-card: "12px"
  pill: "999px"
spacing:
  xs: "4px"
  sm: "8px"
  md: "12px"
  lg: "16px"
  xl: "24px"
components:
  button-primary:
    backgroundColor: "{colors.accent}"
    textColor: "{colors.text-on-accent}"
    rounded: "{rounded.control}"
    padding: "4px 12px 6px"
    height: "32px"
  button-primary-hover:
    backgroundColor: "{colors.accent-hover}"
    textColor: "{colors.text-on-accent}"
    rounded: "{rounded.control}"
    padding: "4px 12px 6px"
    height: "32px"
  card:
    backgroundColor: "{colors.card}"
    textColor: "{colors.text-primary}"
    rounded: "{rounded.surface}"
    padding: "12px 16px"
---

# Design System: ResizeMe

## 1. Overview

**Creative North Star: "The Native Utility Drawer"**

ResizeMe is one focused utility expressed through two native platform vocabularies. The Windows app follows WinUI-inspired density, controls, and tonal layering. The macOS app uses SwiftUI and AppKit conventions: semantic system colors, SF typography and symbols, standard menu-bar hierarchy, grouped forms, sidebar navigation, and platform materials.

Both apps reject a dense control wall with weak hierarchy. Related controls belong in clearly named sections with concise helper text, while the current resize action remains visually primary. Cross-platform consistency comes from task order, language, and behavior, not from forcing identical pixels or controls.

**Key Characteristics:**
- Platform-native controls and interaction states
- WinUI-inspired settings on Windows; SwiftUI and AppKit conventions on macOS
- Restrained accent use reserved for actions, selection, focus, and status
- Compact but deliberate spacing that separates tasks without wasting room
- Structural layering through surfaces and strokes rather than heavy shadows

## 2. Colors

The product identity uses a restrained blue family, while application chrome and controls defer to each platform's adaptive semantic colors.

### Shared identity
- **Icon Blue Light** (`#60A5FA`): Top of the macOS app-icon gradient.
- **Icon Blue** (`#3B6EEB`): Center of the macOS app-icon gradient and shared identity reference.
- **Icon Blue Deep** (`#2240BE`): Bottom of the macOS app-icon gradient.

### Windows primary
- **Windows Action Blue** (`#005FB8`): Primary actions, selected controls, and focus indicators.
- **Windows Action Blue Hover** (`#005FB8E6`): Pointer-hover feedback for accent controls.

### Windows neutral
- **Primary Ink** (`#000000E4`): Headings, labels, and primary content.
- **Secondary Ink** (`#0000009B`): Helper text and supporting metadata.
- **Tertiary Ink** (`#00000072`): Low-emphasis hints that are not required to complete a task.
- **Windows Canvas** (`#F3F3F3`): App background and title bar.
- **Layer Wash** (`#FFFFFF80`): Scrollable content layer.
- **Translucent Card** (`#FFFFFFB3`): Grouped settings and preset surfaces.
- **Quiet Divider** (`#0000000F`): Row separation and card outlines.
- **Control Stroke** (`#00000029`): Stronger control boundaries.
- **Favorite Gold** (`#D48B00`): The active favorite star, paired with its filled icon shape.
- **System Danger** (`#C42B1C`): Destructive actions and validation errors.

### macOS semantic colors
- **Accent** (`Color.accentColor`): Prominent buttons, active controls, selected symbols, and numbered onboarding steps. Respect the user's system accent rather than hardcoding blue.
- **Primary / Secondary / Tertiary** (`.primary`, `.secondary`, `.tertiary`): Adaptive text hierarchy in both appearances.
- **Window / Bar / Grouped Surface** (`.background`, `.bar`, `.quaternary`): Native window content, save bars, and quiet grouped cards.
- **Separator** (`.separator`): Adaptive one-pixel structure around grouped content.
- **Success / Attention / Error** (`.green`, `.orange` or `.yellow`, `.red`): Permission, approval, and load-error status, always paired with text and an SF Symbol.

**The Adaptive macOS Rule.** Never replace SwiftUI semantic colors with fixed light-mode values. macOS surfaces must adapt to appearance, increased contrast, and the user's accent choice.

**The One Accent Rule.** Use accent color only for primary actions, current selection, active toggles, focus, and meaningful status. Inactive content stays neutral.

## 3. Typography

### Windows

**Display Font:** Segoe UI Variable Text (with Segoe UI and system sans-serif fallbacks)  
**Body Font:** Segoe UI Variable Text (with Segoe UI and system sans-serif fallbacks)  
**Icon Fonts:** Segoe Fluent Icons and Segoe MDL2 Assets

**Character:** A single familiar Windows family keeps labels compact and highly legible. Hierarchy comes from weight, size, and spacing rather than competing typefaces.

### Hierarchy
- **Headline** (600, 20px, 28px): Screen title and current preset emphasis.
- **Title** (600, 14px, 20px): Section names and important row labels.
- **Body** (400, 14px, 20px): Controls and task-oriented copy; prose stays below 70ch.
- **Label** (400, 12px, 16px): Helper text, metadata, and compact hints.

### macOS

**Display Font:** SF Pro Rounded for the 26pt onboarding welcome and 22pt About title only.  
**Body Font:** SF Pro Text through SwiftUI semantic styles (`.body`, `.callout`, `.subheadline`, `.footnote`, `.caption`).  
**Icons:** SF Symbols in controls and status labels; the menu-bar asset remains a monochrome template image.

**Character:** System typography should be allowed to respond to macOS accessibility and rendering behavior. Rounded display type is a small moment of personality, not a general-purpose label font.

### macOS hierarchy
- **Welcome** (bold rounded, 26pt): Onboarding title only.
- **About title** (bold rounded, 22pt): Product name in the About screen.
- **Headline** (`.headline`): Settings sections and onboarding step titles.
- **Body / Callout / Subheadline** (system semantic): Controls, descriptions, and explanatory copy.
- **Footnote / Caption** (system semantic): Units, status, version, and tertiary guidance.

**The Sentence Case Rule.** Use natural sentence case for section names, labels, and actions. Avoid repeated uppercase tracked labels.

## 4. Elevation

The system is layered, not lifted. The app canvas, content wash, and translucent cards establish depth through tonal contrast and fine strokes. Shadows are reserved for content that actually floats above the page.

### Shadow Vocabulary
- **Flyout** (`box-shadow: 0 8px 16px rgba(0, 0, 0, 0.14)`): Dialogs and transient overlays only.

On macOS, defer window, popover, menu, and sheet elevation to AppKit. Grouped settings use `.formStyle(.grouped)`, onboarding cards use a quiet `.quaternary` fill with a `.separator` stroke, and bottom action regions use `.bar`. Do not add custom drop shadows to standard SwiftUI controls or settings cards.

**The Structural Elevation Rule.** Resting settings surfaces use borders and tonal layers; only overlays receive a shadow.

## 5. Components

Controls should feel tactile, compact, and recognizably native to the platform that renders them.

### Windows settings

### Buttons
- **Shape:** Compact 4px corners and a 32px control height.
- **Primary:** Windows Action Blue with white text and 12px horizontal padding.
- **Hover / Focus:** Slightly soften the blue on hover; use a visible blue focus outline without changing layout.
- **Secondary / Ghost:** Use translucent neutral fills for standard buttons and transparent blue text for low-emphasis inline actions.

### Cards / Containers
- **Corner Style:** 8px surface corners.
- **Background:** Translucent Card over the Layer Wash.
- **Shadow Strategy:** Flat by default; see Elevation.
- **Border:** Quiet Divider with a stronger bottom edge when needed.
- **Internal Padding:** 12px vertical and 16px horizontal for settings rows.

### Inputs / Fields
- **Style:** Translucent fill, 4px corners, quiet border, and a stronger bottom stroke.
- **Focus:** White active fill with a 2px Windows Action Blue bottom stroke and visible outline.
- **Error / Disabled:** System Danger for error copy; disabled controls retain a clear boundary and non-interactive cursor.

### Navigation
- **Style:** Prefer in-page section navigation or grouped document flow for this compact utility. Current destinations use the accent and all destinations remain keyboard reachable.

### Toggle Switches
- **Style:** 40px by 20px neutral track with a circular thumb; active state uses Windows Action Blue.
- **State:** Labels remain explicit so the switch color is never the only indication of meaning.

### Preset Rows
- **Style:** Compact single-line choices with a native radio indicator, name, dimensions, and adjacent actions.
- **State:** Selection uses both the checked indicator and a subtle blue surface tint. Edit, favorite, and delete actions must remain keyboard discoverable.

### macOS menu bar
- **Structure:** Frontmost app context, permission action when required, Resize Now, status, preset picker, Settings/Updates/About, then Quit. Use standard `Button`, `Picker(.inline)`, and `Divider` controls so the menu inherits native spacing and keyboard behavior.
- **Labels:** Use ellipses for actions that open another window or system surface. Keep standard shortcuts for Settings (`Command-,`) and Quit (`Command-Q`).
- **Icon:** The 18pt menu-bar asset is a black-and-alpha template image so macOS supplies the correct appearance and pressed state.

### macOS settings
- **Navigation:** Use a balanced `NavigationSplitView` with a 150-200pt `.sidebar` list and SF Symbols for General, Presets, Shortcuts, Updates, and About.
- **Forms:** Use grouped `Form` sections for preferences and explanations. Keep the persistent Revert/Save region in a `.bar` surface, with Save as `.borderedProminent` and the default action.
- **Presets:** Use an inset, alternating-row `List` with standard rounded text fields, SF Symbol radio and favorite controls, and borderless add/remove actions.
- **Fields and controls:** Prefer native SwiftUI `Toggle`, `TextField`, `KeyboardShortcuts.Recorder`, `Link`, and system button styles over custom replicas.

### macOS onboarding
- **Layout:** A centered 96pt app icon and short welcome statement lead into two ordered setup cards, followed by one full-width primary action.
- **Cards:** Use 12pt continuous rounded rectangles, `.quaternary.opacity(0.4)` fill, and a one-point `.separator.opacity(0.5)` stroke.
- **Status:** Permission states combine explicit copy, SF Symbols, and semantic green/orange color. Motion is limited to the state change and must honor Reduce Motion.

**The Native Control Rule.** If SwiftUI or AppKit already provides the macOS control, material, menu behavior, dialog, or navigation pattern, use it instead of recreating the Windows treatment.

## 6. Do's and Don'ts

### Do:
- **Do** group controls by the task users are trying to complete and give every section concise helper text.
- **Do** reserve `#005FB8` for Windows primary actions, selection, toggles, and focus.
- **Do** use `Color.accentColor` and adaptive semantic styles in the native macOS app.
- **Do** preserve a logical top-to-bottom keyboard order and visible `:focus-visible` treatment.
- **Do** preserve standard macOS Settings, Quit, default-action, menu, and Accessibility permission conventions.
- **Do** use 4px control corners, 8px surface corners, and the existing 4/8/12/16/24px spacing vocabulary.

### Don't:
- **Don't** create a dense control wall with weak hierarchy or make every option compete equally for attention.
- **Don't** bury frequent tasks among unrelated controls.
- **Don't** import Windows controls, dimensions, Segoe typography, or WinUI card styling into the native SwiftUI app.
- **Don't** import macOS sidebar, menu, or SF Symbol conventions into the Windows Wails app.
- **Don't** introduce decorative patterns that feel foreign to either platform, including glassmorphism, gradient text, or oversized marketing typography.
- **Don't** hide required actions behind hover-only affordances.
- **Don't** add colored side-stripe borders to cards, alerts, or list items.
