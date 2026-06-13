# ResizeMe macOS release setup guide

This note captures the remaining release work for Burke and the production path that should be wired up once the native Swift app is stable enough for real distribution.

## Current status

- The repo already has an unsigned macOS test workflow at `.github/workflows/macos-unsigned-release.yml`.
- The native app now contains the Sparkle update hooks and placeholder feed metadata in `ResizeMeMac/project.yml` and `ResizeMeMac/ResizeMe/Resources/Info.plist`.
- The real signing, notarization, appcast, and Homebrew/cask path is still the final release-stage work.

## What Burke must do to finish the real setup

### 1. Confirm the production bundle identity and signing identity

Before any signed release build is attempted:

1. Confirm the final macOS bundle ID for the native app (`com.resizeme.mac` is currently the placeholder in the plan).
2. Confirm the Developer ID Application certificate that will be used for release signing.
3. Make sure the same signing identity is used consistently for all release builds; changing it later can break Accessibility grants and update behavior.

> This is important because Accessibility permission is tied to the signed bundle identity, not just the app source code.

### 2. Generate the Sparkle EdDSA key pair

Sparkle uses an EdDSA public/private key pair for signed appcasts.

1. In Xcode, open the Swift package dependency for Sparkle.
2. Run the Sparkle key generator tool from the package's `artifacts/sparkle/Sparkle/bin/` folder.
3. Save the private key securely and copy the public key into the app's Info.plist as `SUPublicEDKey`.
4. Add the exported private key text to GitHub Actions secrets as `SPARKLE_PRIVATE_KEY`.

The public key placeholder is currently in `ResizeMeMac/project.yml` and should be replaced with the real value before the first real Sparkle test.

### 3. Add the required GitHub Actions secrets

Add these repository secrets under GitHub → Settings → Secrets and variables → Actions:

- `SPARKLE_PRIVATE_KEY` — the exported private key from the Sparkle key generator
- `DEVELOPER_ID_CERTIFICATE_BASE64` — base64-encoded `.p12` Developer ID certificate
- `DEVELOPER_ID_CERTIFICATE_PASSWORD` — password for that `.p12`
- `KEYCHAIN_PASSWORD` — temporary keychain password used in CI
- `APPLE_ID` — Apple ID email for notarization
- `APP_PASSWORD` — app-specific password for notarization
- `APPLE_TEAM_ID` — Apple Developer Team ID

These are the same secret categories used by the Tiny Clips release setup and are required before a signed release workflow can run end-to-end.

### 4. Enable GitHub Pages for the appcast

Sparkle expects an appcast feed that is reachable over HTTPS.

1. Enable GitHub Pages for the repo.
2. Publish the appcast from the `docs/` folder (or another Pages path that the workflow will write to).
3. Confirm the final feed URL matches `SUFeedURL` in the app plist.

Expected shape:

- `https://<owner>.github.io/<repo>/appcast.xml`

### 5. Add the signed release workflow

The current unsigned workflow is suitable for testing only. The production path should:

1. build the native macOS app in Release mode,
2. sign it with the Developer ID certificate,
3. notarize the build with Apple,
4. create a ZIP or DMG artifact,
5. generate the Sparkle appcast and sign the update item,
6. publish the release to GitHub Releases,
7. deploy the updated `appcast.xml` to GitHub Pages.

This is the same high-level flow used by Tiny Clips for direct-download release distribution.

### 6. Test the actual update path before public release

Before shipping to users:

1. create a test release tag,
2. verify the appcast is reachable,
3. install the test artifact on a clean machine,
4. confirm Sparkle finds the update and shows release notes,
5. confirm the update downloads and installs correctly.

### 7. Add Homebrew cask packaging after the release URL is stable

Homebrew should be treated as the final delivery layer, not the first one.

The Tiny Clips repo uses a cask pattern like this:

- `Casks/tiny-clips.rb`
- version pinned to the GitHub release tag
- SHA256 pinned to the downloaded ZIP
- `url` pointing at the GitHub release asset
- `postflight` removing quarantine attributes with `xattr -dr com.apple.quarantine`

For ResizeMe, the same pattern should be adopted once:

- the bundle ID is finalized,
- the signing identity is stable,
- the release URL format is permanent,
- the maintainer is happy with the distribution ownership model.

The cask should live in a `Casks/resize-me.rb` file in the tap repository and be updated automatically from the release workflow when a new tagged release is published.

## Recommended Burke checklist

- [ ] Finalize the macOS bundle ID and signing identity
- [ ] Generate the Sparkle key pair and store the private key securely
- [ ] Add all Sparkle + notarization GitHub secrets
- [ ] Enable GitHub Pages and verify the appcast URL
- [ ] Add the signed/notarized release workflow
- [ ] Run one real Sparkle update test on a clean machine
- [ ] Add the Homebrew cask once the release artifact and versioning are stable

## Reference model

Tiny Clips is the best reference implementation for the release/distribution side of this work:

- Sparkle setup and update-check flow
- direct-download release workflow
- appcast generation and release note handling
- Homebrew cask packaging and version/SHA automation

Use that repo as a pattern, but copy the pieces deliberately into this repo and document them here so the release path remains maintainable.
