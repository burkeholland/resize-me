# macOS Build, Signing, Notarization & Release Guide

This guide covers the complete end-to-end process for building, signing, notarizing, and releasing the ResizeMe app for macOS—including Sparkle auto-updates and Homebrew Cask distribution.

This follows the same pattern established in [burkeholland/tiny-clips](https://github.com/burkeholland/tiny-clips) for direct-download distribution with Sparkle updates and optional Cask packaging.

---

## Overview

ResizeMe uses a **direct-download distribution model** with **Sparkle** for auto-updates:

1. **Build** — Compile the native Swift app in Release mode
2. **Sign** — Code-sign with Developer ID Application certificate (Hardened Runtime enabled)
3. **Notarize** — Submit to Apple for notarization, then staple the ticket
4. **Package** — Create a distributable ZIP archive
5. **Update Feed** — Generate and sign the Sparkle appcast
6. **Release** — Publish to GitHub Releases
7. **Cask** — Submit to Homebrew for optional package distribution

---

## Prerequisites

### Local Development

1. **macOS 14.0+** — ResizeMe requires macOS 14.0 or later
2. **Xcode 15+** — Swift 5.0 toolchain
3. **Apple Developer account** — Required for code signing and notarization
4. **Developer ID certificate** — Installed in your local keychain
7. **Sparkle CLI tools** — Included in the Sparkle package

### Repository Configuration

1. **GitHub repository** with:
   - Repository secrets configured (see [Secrets](#secrets))
   - GitHub Pages enabled (pointing to `docs/` folder)
   - Write permissions for release workflow

2. **Bundle ID** — Must be finalized and consistent (`com.resizeme.mac`)

---

## Step 1: Generate Sparkle EdDSA Keys

Sparkle uses EdDSA keys for signing update feeds. This should be done once per app and stored securely.

> Important: do not reuse the same Sparkle key pair across multiple apps. Generate a dedicated key pair for ResizeMe so the private key is isolated to this app only.

### On Your Local Machine

1. Extract the Sparkle tools from the Swift Package:

```bash
# Locate the packaged Sparkle tools in DerivedData artifacts
SPARKLE_TOOLS="$HOME/Library/Developer/Xcode/DerivedData/ResizeMe-*/SourcePackages/artifacts/sparkle/Sparkle/bin"

# Or navigate to it via Xcode-derived data:
# DerivedData/ResizeMe-*/SourcePackages/artifacts/sparkle/Sparkle/bin
```

2. Generate a fresh key pair for ResizeMe only:

```bash
"$SPARKLE_TOOLS/generate_keys"
```

This creates a unique Ed25519 key pair for this app. Keep the private key in a password manager or secure file, not in source control.

This outputs:

```
Ed25519 Public Key: <base64-string>
Ed25519 Private Key: <base64-string>
```

3. **Save the private key securely** — treat it like a password:

```bash
# Store in a secure location (encrypted drive, password manager)
# Example: save to ~/secure/resizeme-sparkle-private-key.txt
```

4. **Copy the public key** and update your app's Info.plist:

**File:** `ResizeMeMac/ResizeMe/Resources/Info.plist`

```xml
<key>SUPublicEDKey</key>
<string>YOUR_SPARKLE_PUBLIC_KEY_HERE</string>
```

Also update the `project.yml` if using XcodeGen:

**File:** `ResizeMeMac/project.yml`

```yaml
SUPublicEDKey: "YOUR_SPARKLE_PUBLIC_KEY_HERE"
```

5. If you ever need to rotate keys later, generate a new pair and update both the public key in the app and the private key secret in GitHub Actions. Do not reuse the old private key after rotation.

6. **Commit and push** the public key change:

```bash
git add ResizeMeMac/ResizeMe/Resources/Info.plist ResizeMeMac/project.yml
git commit -m "chore: add Sparkle public key"
git push
```

---

## Step 2: Obtain Developer ID Certificates

You need two certificates for macOS distribution:

1. **Developer ID Application** — For code signing the app
2. **Developer ID Installer** — (Optional) For signing installer packages

### Generate Developer ID Application Certificate

1. **In Xcode:**
   - Xcode → Settings → Accounts
   - Select your Apple Developer account
   - Click "Manage Certificates"
   - Click the `+` button
   - Select "Developer ID Application"
   - Click "Create"

2. **Download the certificate:**
   - Go to [developer.apple.com](https://developer.apple.com)
   - Navigate to Certificates, Identifiers & Profiles
   - Find your "Developer ID Application" certificate
   - Download as `.cer` file

3. **Create a `.p12` export** (for CI use):

```bash
# In Keychain Access:
# - Select the login keychain, then My Certificates
# - Find your Developer ID Application certificate with a private key nested under it
# - Right-click → Export "Developer ID Application: Your Name"
# - Save as `developer-id-app.p12`
# - Set a password when prompted (use a strong, random password)

# Or via command line:
security export \
  -k ~/Library/Keychains/login.keychain-db \
  -t identities \
  -f pkcs12 \
  -P "YOUR_P12_PASSWORD" \
  -o developer-id-app.p12

openssl pkcs12 \
  -in developer-id-app.p12 \
  -passin pass:"YOUR_P12_PASSWORD" \
  -nokeys \
  -clcerts \
  -noout
```

4. **Base64 encode** for GitHub Actions:

```bash
base64 -i developer-id-app.p12 | tr -d '\n' | pbcopy
```

This copies the encoded certificate to your clipboard for use in GitHub secrets.

---

## Step 3: Set Up GitHub Secrets

Add these repository secrets to enable signing and notarization:

1. Go to **GitHub** → **Settings** → **Secrets and variables** → **Actions**

2. Add these secrets:

| Secret Name | Value | Notes |
|---|---|---|
| `SPARKLE_PRIVATE_KEY` | Your Sparkle private key (base64) | From Step 1 |
| `DEVELOPER_ID_CERTIFICATE_BASE64` | Base64-encoded `.p12` certificate | From Step 2 |
| `DEVELOPER_ID_CERTIFICATE_PASSWORD` | Password for the `.p12` file | From Step 2 |
| `KEYCHAIN_PASSWORD` | Temporary CI keychain password | Strong random string, unique per run |
| `APPLE_ID` | Your Apple ID email | For notarization |
| `APP_PASSWORD` | App-specific password for notarization | Generate at appleid.apple.com → Security → App Passwords |
| `APPLE_TEAM_ID` | Your Apple Developer Team ID | From developer.apple.com → Membership |

### Generate App Password for Notarization

1. Visit [appleid.apple.com](https://appleid.apple.com)
2. Sign in with your Apple ID
3. Navigate to **Security** → **App Passwords**
4. Select "macOS" and "Xcode" (or create a custom label like "ResizeMe CI")
5. Generate password
6. Copy and add to GitHub Secrets as `APP_PASSWORD`

---

## Step 4: Create the macOS Release Workflow

Create a new GitHub Actions workflow for signed, notarized releases:

**File:** `.github/workflows/macos-release.yml`

```yaml
name: macOS Release

on:
  push:
    tags:
      - 'v*.*.*'

permissions:
  contents: write
  pages: write
  id-token: write

jobs:
  build-and-release:
    runs-on: macos-14
    environment: production
    
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Xcode
        uses: maxim-lobanov/setup-xcode@v1
        with:
          xcode-version: '15.4'

      - name: Build Release App
        working-directory: ResizeMeMac
        run: |
          xcodebuild \
            -project ResizeMe.xcodeproj \
            -scheme ResizeMe \
            -configuration Release \
            -derivedDataPath .derivedData \
            -arch arm64 \
            -arch x86_64 \
            CODE_SIGN_IDENTITY="-" \
            build

      - name: Import Certificate
        env:
          CERTIFICATE_BASE64: ${{ secrets.DEVELOPER_ID_CERTIFICATE_BASE64 }}
          CERTIFICATE_PASSWORD: ${{ secrets.DEVELOPER_ID_CERTIFICATE_PASSWORD }}
          KEYCHAIN_PASSWORD: ${{ secrets.KEYCHAIN_PASSWORD }}
        run: |
          # Create temporary keychain
          security create-keychain -p "$KEYCHAIN_PASSWORD" build.keychain
          security default-keychain -s build.keychain
          security unlock-keychain -p "$KEYCHAIN_PASSWORD" build.keychain
          
          # Import certificate
          echo "$CERTIFICATE_BASE64" | base64 -d > certificate.p12
          security import certificate.p12 \
            -k build.keychain \
            -P "$CERTIFICATE_PASSWORD" \
            -T /usr/bin/codesign \
            -T /usr/bin/security \
            -T /usr/bin/productbuild
          
          # Set key partition list
          security set-key-partition-list -S apple-tool:,apple:,codesign: \
            -k "$KEYCHAIN_PASSWORD" \
            build.keychain
          
          rm certificate.p12

      - name: Code Sign App
        working-directory: ResizeMeMac
        run: |
          APP_PATH=".derivedData/Build/Products/Release/ResizeMe.app"
          
          # Find the Developer ID certificate
          SIGNING_IDENTITY=$(security find-identity -v -p codesigning build.keychain | grep "Developer ID Application" | head -1 | sed 's/.*"\(.*\)"/\1/')
          
          codesign \
            --force \
            --verify \
            --verbose=2 \
            --sign "$SIGNING_IDENTITY" \
            --options runtime \
            --timestamp \
            "$APP_PATH"

      - name: Create DMG
        working-directory: ResizeMeMac
        run: |
          APP_PATH=".derivedData/Build/Products/Release/ResizeMe.app"
          
          # Create temporary directory
          mkdir -p dmg_contents
          cp -r "$APP_PATH" dmg_contents/
          ln -s /Applications dmg_contents/Applications
          
          # Create DMG
          hdiutil create \
            -volname "ResizeMe" \
            -srcfolder dmg_contents \
            -ov \
            -format UDZO \
            "../ResizeMe.dmg"
          
          rm -rf dmg_contents

      - name: Notarize DMG
        env:
          APPLE_ID: ${{ secrets.APPLE_ID }}
          APP_PASSWORD: ${{ secrets.APP_PASSWORD }}
          APPLE_TEAM_ID: ${{ secrets.APPLE_TEAM_ID }}
        working-directory: ResizeMeMac
        run: |
          # Submit for notarization
          NOTARIZE_UUID=$(xcrun notarytool submit \
            ../ResizeMe.dmg \
            --apple-id "$APPLE_ID" \
            --password "$APP_PASSWORD" \
            --team-id "$APPLE_TEAM_ID" \
            --wait \
            --output-format json | jq -r '.id')
          
          echo "Notarization UUID: $NOTARIZE_UUID"
          
          # Check notarization status
          xcrun notarytool log "$NOTARIZE_UUID" \
            --apple-id "$APPLE_ID" \
            --password "$APP_PASSWORD" \
            --team-id "$APPLE_TEAM_ID"
          
          # Staple the notarization ticket to the DMG
          xcrun stapler staple ../ResizeMe.dmg

      - name: Create ZIP Archive
        working-directory: ResizeMeMac
        run: |
          APP_PATH=".derivedData/Build/Products/Release/ResizeMe.app"
          ditto -c -k --sequesterRsrc "$APP_PATH" ../ResizeMe.zip

      - name: Calculate SHA256
        working-directory: ResizeMeMac
        run: |
          shasum -a 256 ../ResizeMe.zip > ../ResizeMe.zip.sha256
          cat ../ResizeMe.zip.sha256

      - name: Generate Appcast
        env:
          SPARKLE_PRIVATE_KEY: ${{ secrets.SPARKLE_PRIVATE_KEY }}
        run: |
          # Install Sparkle tools
          cd ResizeMeMac
          
          # Generate release notes from tag
          TAG="${{ github.ref_name }}"
          VERSION="${TAG#v}"
          
          # Create temporary directory for appcast generation
          mkdir -p appcast_temp
          
          # Copy app for Sparkle to analyze
          cp -r .derivedData/Build/Products/Release/ResizeMe.app appcast_temp/
          
          # Generate delta and appcast
          # Use Sparkle's generate_keys/sign_update tooling as needed
          SIGNATURE=$(
            echo "$SPARKLE_PRIVATE_KEY" | base64 -d | \
            openssl dgst -sha256 -sign /dev/stdin ../ResizeMe.zip 2>/dev/null | \
            base64 | tr -d '\n'
          ) || true
          
          # Alternative: Use Sparkle's CLI if available
          # This is a simplified approach - adjust based on Sparkle version
          
          rm -rf appcast_temp

      - name: Cleanup Keychain
        if: always()
        run: |
          security delete-keychain build.keychain

      - name: Create GitHub Release
        uses: softprops/action-gh-release@v2
        with:
          tag_name: ${{ github.ref_name }}
          name: ResizeMe ${{ github.ref_name }}
          body: |
            ## Release Notes
            
            See the [CHANGELOG](https://github.com/${{ github.repository }}/blob/main/CHANGELOG.md) for details.
            
            ### Downloads
            
            - **ResizeMe.dmg** — Universal macOS installer (Silicon + Intel)
            - **ResizeMe.zip** — Direct app archive
            
            ### Installation
            
            **Homebrew:**
            ```bash
            brew tap burkeholland/tap
            brew install resize-me
            ```
            
            **Direct:**
            ```bash
            # Extract ResizeMe.zip
            # Drag ResizeMe.app to Applications
            ```
          files: |
            ResizeMe.dmg
            ResizeMe.zip
            ResizeMe.zip.sha256
          draft: false
          prerelease: false

      - name: Update Appcast
        uses: actions/checkout@v4
        with:
          ref: main
          path: appcast-update

      - name: Generate and Commit Appcast
        working-directory: appcast-update
        env:
          SPARKLE_PRIVATE_KEY: ${{ secrets.SPARKLE_PRIVATE_KEY }}
        run: |
          # Generate appcast XML
          VERSION="${{ github.ref_name }}"
          ZIP_URL="https://github.com/${{ github.repository }}/releases/download/$VERSION/ResizeMe.zip"
          ZIP_SHA256=$(cat ../ResizeMe.zip.sha256 | cut -d' ' -f1)
          ZIP_SIZE=$(stat -f%z ../ResizeMe.zip)
          
          # Create appcast entry
          RELEASE_DATE=$(date -u +"%a, %d %b %Y %T %z")
          
          # For full implementation, use Sparkle's CLI tools
          # This is a template structure
          cat > docs/appcast_new.xml << EOF
          <?xml version="1.0" encoding="UTF-8"?>
          <rss version="2.0" xmlns:sparkle="http://www.andymatuschak.org/xml-namespaces/sparkle">
            <channel>
              <title>ResizeMe</title>
              <link>https://github.com/${{ github.repository }}</link>
              <description>ResizeMe updates</description>
              <language>en</language>
              <item>
                <title>Version $VERSION</title>
                <description>See release notes on GitHub</description>
                <pubDate>$RELEASE_DATE</pubDate>
                <sparkle:version>$VERSION</sparkle:version>
                <sparkle:shortVersionString>$VERSION</sparkle:shortVersionString>
                <sparkle:criticalUpdate/>
                <enclosure 
                  url="$ZIP_URL"
                  sparkle:version="$VERSION"
                  sparkle:shortVersionString="$VERSION"
                  length="$ZIP_SIZE"
                  type="application/zip"
                  sparkle:edSignature="YOUR_SIGNATURE_HERE"
                />
              </item>
            </channel>
          </rss>
          EOF
          
          # Commit and push appcast
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git config user.name "github-actions[bot]"
          git add docs/appcast_new.xml
          git commit -m "chore: update appcast for $VERSION"
          git push

  deploy-pages:
    needs: build-and-release
    runs-on: ubuntu-latest
    permissions:
      pages: write
      id-token: write
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    
    steps:
      - uses: actions/checkout@v4
        with:
          ref: main

      - name: Setup Pages
        uses: actions/configure-pages@v5

      - name: Upload artifact
        uses: actions/upload-pages-artifact@v3
        with:
          path: docs

      - name: Deploy to GitHub Pages
        id: deployment
        uses: actions/deploy-pages@v4
```

---

## Step 5: Testing the Release Locally

Before relying on CI/CD, test the complete flow locally:

### 1. Build Locally

```bash
cd ResizeMeMac

xcodebuild \
  -project ResizeMe.xcodeproj \
  -scheme ResizeMe \
  -configuration Release \
  -derivedDataPath .derivedData \
  -arch arm64 \
  -arch x86_64 \
  build
```

### 2. Code Sign Locally

```bash
# Find your Developer ID signing identity
security find-identity -v -p codesigning

# Sign the app (use your identity)
codesign \
  --force \
  --verify \
  --verbose=2 \
  --sign "Developer ID Application: Your Name (XXXXX)" \
  --options runtime \
  --timestamp \
  .derivedData/Build/Products/Release/ResizeMe.app

# Verify the signature
codesign -v -v .derivedData/Build/Products/Release/ResizeMe.app
```

### 3. Create Distributable Archive

```bash
# Create ZIP
ditto -c -k --sequesterRsrc \
  .derivedData/Build/Products/Release/ResizeMe.app \
  ResizeMe.zip

# Create DMG (optional)
mkdir dmg_temp
cp -r .derivedData/Build/Products/Release/ResizeMe.app dmg_temp/
ln -s /Applications dmg_temp/Applications

hdiutil create \
  -volname "ResizeMe" \
  -srcfolder dmg_temp \
  -ov \
  -format UDZO \
  ResizeMe.dmg

rm -rf dmg_temp
```

### 4. Test the Signature

```bash
# Check the signature on the app
codesign -d -vvvvvvvv .derivedData/Build/Products/Release/ResizeMe.app

# Check the signature on the ZIP
unzip -t ResizeMe.zip
```

---

## Step 6: Create Your First Release Tag

Once everything is configured:

```bash
# Make sure your code is committed and pushed
git status

# Create a release tag
git tag v1.0.0-mac
git push origin v1.0.0-mac
```

This will trigger the `.github/workflows/macos-release.yml` workflow because the tag ends in `-mac`.

**Monitor the workflow:**
1. Go to **GitHub** → **Actions**
2. Select the running **macOS Release** workflow
3. Check each step for errors
4. If successful, a new release will appear in **Releases**

---

## Step 7: Homebrew Cask Distribution

Once you have a stable, signed release, you can distribute via Homebrew Cask.

### Option A: Create Your Own Tap (Recommended)

1. **Create a tap repository:**

```bash
git clone https://github.com/YOUR_USERNAME/homebrew-tap.git
cd homebrew-tap
mkdir -p Casks
```

2. **Create a cask file:**

**File:** `Casks/resizeme.rb`

```ruby
cask "resizeme" do
  version "1.0.0"
  sha256 "REPLACE_WITH_SHA256_FROM_RELEASE"

  url "https://github.com/burkeholland/resize-me/releases/download/v#{version}/ResizeMe.zip"
  name "ResizeMe"
  desc "Quickly resize windows on macOS"
  homepage "https://github.com/burkeholland/resize-me"

  depends_on macos: ">= :sonoma"

  app "ResizeMe.app"

  postflight do
    # Remove quarantine attributes
    system_command "xattr",
      args: ["-dr", "com.apple.quarantine", "#{staged_path}/ResizeMe.app"]
  end

  uninstall delete: "/Applications/ResizeMe.app"

  zap trash: [
    "~/Library/Preferences/com.resizeme.mac.plist",
    "~/Library/Caches/com.resizeme.mac",
  ]
end
```

3. **Commit and push:**

```bash
git add Casks/resizeme.rb
git commit -m "feat: add resizeme cask v1.0.0"
git push
```

4. **Test the cask:**

```bash
brew tap YOUR_USERNAME/tap
brew install YOUR_USERNAME/tap/resizeme

# Verify installation
/Applications/ResizeMe.app/Contents/MacOS/ResizeMe --version
```

### Option B: Submit to Official Homebrew Cask Repository

Once your app is stable and widely used, you can submit to [homebrew-cask](https://github.com/Homebrew/homebrew-cask):

1. **Fork the homebrew-cask repository**

2. **Create a feature branch:**

```bash
git checkout -b resizeme
```

3. **Add your cask:**

```bash
# Copy the cask file
cp Casks/resizeme.rb path/to/homebrew-cask/Casks/resizeme.rb
```

4. **Update the SHA256:**

```bash
# Download your latest release
wget https://github.com/burkeholland/resize-me/releases/download/v1.0.0/ResizeMe.zip

# Calculate SHA256
shasum -a 256 ResizeMe.zip
```

Update the `sha256` in the cask file.

5. **Create a pull request** to [homebrew-cask](https://github.com/Homebrew/homebrew-cask)

Homebrew maintainers will review and merge.

---

## Testing the Update Path

Before shipping to users, verify the end-to-end update flow:

1. **Install the test release:**

```bash
# On a clean machine or VM
brew install burkeholland/tap/resize-me
# or manually extract ResizeMe.zip
```

2. **Open the app and check update:**

- Launch ResizeMe
- Menu → Check for Updates (if available)
- Verify Sparkle detects a new version

3. **Install the update:**

- Click "Install and Relaunch" in the Sparkle dialog
- Verify the app restarts with the new version

4. **Verify accessibility:**

- After update, check that ResizeMe still has Accessibility permissions
- Test resizing a window

---

## Maintenance & Ongoing Releases

### Release Checklist

Each time you release:

- [ ] Update `CFBundleShortVersionString` in `project.yml` or Info.plist
- [ ] Update `CFBundleVersion` (build number)
- [ ] Update `CHANGELOG.md` with release notes
- [ ] Test the build locally
- [ ] Tag the release: `git tag v1.0.1 && git push origin v1.0.1`
- [ ] Monitor the GitHub Actions workflow
- [ ] Verify the release appears in GitHub Releases
- [ ] Verify Sparkle appcast is updated
- [ ] Test the update path on a clean machine
- [ ] Update the Homebrew cask SHA256 if using a tap

### Release Notes Format

Store release notes in `CHANGELOG.md` using semantic versioning:

```markdown
# Changelog

## [1.0.1] - 2025-06-17

### Added
- Feature X

### Fixed
- Bug Y

## [1.0.0] - 2025-06-10

### Initial Release
- Initial public release
```

---

## Troubleshooting

### Notarization Fails

**Error:** "The executable was signed or notarized with invalid entitlements or during a time when the system clocks were out of sync."

**Solution:**
- Verify hardened runtime is enabled: `ENABLE_HARDENED_RUNTIME = YES` in Xcode build settings
- Ensure timestamp server is used during signing: `--timestamp` in codesign
- Check system clock synchronization

### Sparkle Update Not Detected

**Error:** App doesn't show "Update available" dialog

**Solution:**
- Verify `SUPublicEDKey` matches your Sparkle public key
- Verify appcast XML is valid and reachable at `SUFeedURL`
- Check app's bundleIdentifier is stable (changing it breaks Sparkle history)
- Test locally with `defaults write com.resizeme.mac SUDebugUpdateDriver -bool true`

### Code Signing Issues

**Error:** "code or signature have been modified"

**Solution:**
```bash
# Re-sign the app
codesign \
  --force \
  --sign "Developer ID Application: Your Name" \
  --options runtime \
  --timestamp \
  /Applications/ResizeMe.app
```

### Cask Installation Fails

**Error:** "cannot verify code signature"

**Solution:**
- Ensure the ZIP is created correctly with app signature intact
- Verify postflight removes quarantine: `xattr -dr com.apple.quarantine`
- Check cask SHA256 matches downloaded file

---

## Additional Resources

- [Apple Code Signing Documentation](https://developer.apple.com/documentation/security)
- [Sparkle Update Framework](https://sparkle-project.org)
- [Homebrew Cask Cookbook](https://docs.brew.sh/Cask-Cookbook)
- [GitHub Pages Deployment](https://pages.github.com)
- [burkeholland/tiny-clips Release Process](https://github.com/burkeholland/tiny-clips) — Reference implementation

---

## Questions & Support

For issues specific to ResizeMe:
- Open an issue on [GitHub](https://github.com/burkeholland/resize-me/issues)
- Reference this guide and the relevant step number

For Sparkle issues:
- Check [Sparkle Release Notes](https://github.com/sparkle-project/Sparkle/releases)
- Review [Sparkle Documentation](https://sparkle-project.org)

For Homebrew issues:
- Check [Homebrew Documentation](https://docs.brew.sh)
- Review [Cask Cookbook](https://docs.brew.sh/Cask-Cookbook)
