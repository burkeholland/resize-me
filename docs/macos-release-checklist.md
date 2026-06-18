# macOS Release Quick-Start Checklist

Quick reference for releasing ResizeMe on macOS. See [macos-build-sign-release.md](./macos-build-sign-release.md) for full details.

## One-Time Setup

- [ ] **Generate Sparkle keys**
  - Run Sparkle key generator: `sparkle-generate-update-keys`
  - Save private key securely
  - Update `Info.plist` + `project.yml` with public key
  - Commit changes

- [ ] **Obtain Developer ID certificate**
  - Generate in Xcode: Xcode → Settings → Accounts → Manage Certificates
  - Export as `.p12` with strong password
  - Base64 encode for GitHub

- [ ] **Add GitHub Secrets** (Settings → Secrets and variables → Actions)
  - `SPARKLE_PRIVATE_KEY`
  - `DEVELOPER_ID_CERTIFICATE_BASE64`
  - `DEVELOPER_ID_CERTIFICATE_PASSWORD`
  - `KEYCHAIN_PASSWORD`
  - `APPLE_ID`
  - `APP_PASSWORD` (from appleid.apple.com → Security)
  - `APPLE_TEAM_ID`

- [ ] **Enable GitHub Pages**
  - Settings → Pages → Source: Deploy from branch (main)
  - Branch: main, folder: /docs

- [ ] **Create Homebrew tap** (optional)
  - Create `homebrew-tap` repository
  - Create `Casks/resizeme.rb` cask file

## For Each Release

### Before Release

- [ ] Update version in `project.yml`:
  - `CFBundleShortVersionString: "X.Y.Z"`
  - `CFBundleVersion: "N"` (increment)

- [ ] Update `CHANGELOG.md` with release notes

- [ ] Test build locally:
  ```bash
  cd ResizeMeMac
  xcodebuild -project ResizeMe.xcodeproj -scheme ResizeMe \
    -configuration Release -derivedDataPath .derivedData \
    -arch arm64 -arch x86_64 build
  ```

- [ ] Commit and push changes

### Create Release

1. **Create and push tag:**
   ```bash
   git tag v1.0.0-mac
   git push origin v1.0.0
   ```

2. **Monitor GitHub Actions:**
   - Go to Actions → macOS Release workflow
   - Verify all steps pass (build, sign, notarize, appcast)

3. **Verify release:**
   - Check Releases page for new version
   - Download and test ResizeMe.dmg or ResizeMe.zip
   - Verify signature: `codesign -v /Applications/ResizeMe.app`

### After Release

- [ ] **Update Homebrew cask** (if using custom tap):
  - Update SHA256: `shasum -a 256 ResizeMe.zip`
  - Update version in `Casks/resizeme.rb`
  - Push to homebrew-tap repository

- [ ] **Announce release**
  - Twitter/social media
  - Product Hunt (if applicable)
  - Dev community forums

- [ ] **Test update path** (next version):
  - Install previous version: `brew install homebrew-tap/resizeme`
  - Wait for Sparkle to detect new version
  - Verify update installs successfully
  - Check Accessibility permissions still work

## Troubleshooting Quick Fixes

| Issue | Fix |
|-------|-----|
| Notarization fails | Check the notary log; verify hardened runtime, `--timestamp`, and that embedded code is signed before the app |
| Stapler fails on ZIP | Staple `ResizeMe.app`, validate it, then recreate `ResizeMe.zip` from the stapled app |
| Sparkle not detecting update | Verify public key matches; check appcast.xml reachable at `SUFeedURL` |
| Code signature invalid | Re-sign: `codesign --force --sign "Developer ID..." --options runtime --timestamp /Applications/ResizeMe.app` |
| Cask install fails | Verify SHA256 in cask matches; ensure ZIP signature is intact |

## Key Files

| File | Purpose |
|------|---------|
| `ResizeMeMac/project.yml` | Build config (version, bundle ID, Sparkle key) |
| `ResizeMeMac/ResizeMe/Resources/Info.plist` | App plist (Sparkle feed URL, public key) |
| `.github/workflows/macos-release.yml` | CI/CD workflow (sign, notarize, release) |
| `docs/appcast.xml` | Sparkle update feed |
| `Casks/resizeme.rb` | Homebrew cask definition |
| `CHANGELOG.md` | Release notes |

## Useful Commands

```bash
# Build
cd ResizeMeMac && xcodebuild -project ResizeMe.xcodeproj -scheme ResizeMe \
  -configuration Release -derivedDataPath .derivedData build

# Sign
codesign --force --sign "Developer ID Application: Your Name" \
  --options runtime --timestamp ResizeMe.app

# Create ZIP
ditto -c -k --sequesterRsrc ResizeMe.app ResizeMe.zip

# Calculate SHA256
shasum -a 256 ResizeMe.zip

# Verify signature
codesign -v -v ResizeMe.app

# Check for updates (debug mode)
defaults write com.resizeme.mac SUDebugUpdateDriver -bool true
```

## GitHub Workflow File Location

`.github/workflows/macos-release.yml` — triggers on `v*.*.*-mac` tags

---

For full documentation, see [macos-build-sign-release.md](./macos-build-sign-release.md).
