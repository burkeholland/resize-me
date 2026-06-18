# ResizeMe for macOS

ResizeMe is a tiny menu-bar helper for macOS that resizes the frontmost window to a preset size with a keyboard shortcut.

If a request does not clearly say whether you mean this native macOS app or the original Go/Wails app in `../ResizeMe/`, ask which one they mean before making changes.

## What it does

- Resizes any app window to a chosen preset dimension
- Runs from the menu bar for quick access
- Uses a global shortcut for fast window resizing
- Supports optional launch-at-login behavior
- Uses macOS Accessibility permissions to move and resize other windows safely

## Highlights

- Simple, lightweight menu-bar experience
- Preset-based resizing with center-after-resize support
- Built-in onboarding and Settings flow
- Sparkle-based update support for release builds

## Requirements

- macOS 14.0 or later
- Xcode 15 or later
- Accessibility permission granted to ResizeMe

## Install

Install the signed, notarized macOS release with Homebrew:

```sh
brew tap burkeholland/resize-me https://github.com/burkeholland/resize-me
brew install --cask resizeme
```

Update it later with:

```sh
brew update
brew upgrade --cask resizeme
```

You can also download `ResizeMe.zip` manually from the [latest macOS release](https://github.com/burkeholland/resize-me/releases?q=-mac).

## Build and run

From the project root:

```sh
cd ResizeMeMac
xcodebuild -project ResizeMe.xcodeproj -scheme ResizeMe -configuration Debug -derivedDataPath .derivedData build CODE_SIGN_IDENTITY='-'
open .derivedData/Build/Products/Debug/ResizeMe.app
```

You can also use the included Xcode task for building and running the app from VS Code.

## Run tests

```sh
cd ResizeMeMac
xcodebuild -project ResizeMe.xcodeproj -scheme ResizeMe -configuration Debug test CODE_SIGN_IDENTITY='-'
```

## Notes

- The app is designed as a menu bar utility, so it stays out of your way until you need it.
- The first run includes an onboarding step to help with Accessibility permission and launch-at-login setup.
- Release builds are configured for Sparkle auto-updates and notarized distribution.

## Related docs

- [Release strategy](../memories/repo/resizeme-release-strategy.md)
- [macOS release checklist](../docs/macos-release-checklist.md)
- [macOS build and signing guide](../docs/macos-build-sign-release.md)
