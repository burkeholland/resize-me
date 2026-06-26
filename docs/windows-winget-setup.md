# ResizeMe Windows winget setup guide

This guide describes how to publish the `ResizeMe/` Go/Wails Windows build to winget by shipping the signed `.exe` release assets directly, not MSIX.

## Target distribution model

Use the existing Wails Windows binaries as the winget installers:

- `ResizeMe-windows-amd64.exe`
- `ResizeMe-windows-arm64.exe`

Do not copy the Tiny Clips MSIX path directly. Tiny Clips uses MSIX because it is a packaged WinUI app and needs package identity, Windows App SDK framework dependencies, WACK validation, `PackageFamilyName`, and `SignatureSha256`. ResizeMe can stay simpler:

- `InstallerType: portable` for direct `.exe` assets.
- `InstallerSha256` only; no MSIX `SignatureSha256`.
- No `PackageFamilyName`.
- No WACK step.
- Add a WebView2 runtime dependency because Wails needs WebView2 on Windows.

If we later want Start menu shortcuts, Add/Remove Programs entries, or a managed uninstall path, switch from direct `.exe` to the Wails NSIS installer and use `InstallerType: nullsoft`. For the requested "ship the exe" path, use `portable`.

## Implemented files

- `.github/workflows/release.yml` — builds x64 and ARM64 Windows `.exe` assets only when a `v*.*.*-windows` tag is pushed, signs them, verifies signatures, generates winget manifests, and creates the GitHub Release.
- `.github/workflows/winget-submit.yml` — manually submits a published Windows release to `microsoft/winget-pkgs` using `wingetcreate`.
- `ResizeMe/packaging/winget/` — source winget manifest templates for `BurkeHolland.ResizeMe`.

## One-time decisions

1. Confirm the winget package identifier. Recommended:

   ```text
   BurkeHolland.ResizeMe
   ```

2. Confirm the displayed publisher string. Recommended:

   ```text
   Burke Holland
   ```

3. Confirm the final release tag format. Recommended:

   ```text
   v0.0.1-windows
   ```

   This mirrors the macOS release tags such as `v0.0.1-mac`, keeps Windows releases easy to filter, and maps to the numeric winget package version `0.0.1`.

4. Keep Windows release asset names stable:

   ```text
   ResizeMe-windows-amd64.exe
   ResizeMe-windows-arm64.exe
   ```

## Signing requirement

The `.exe` assets should be Authenticode signed before they are attached to the GitHub Release. Winget validates hashes, but signing is still important for SmartScreen reputation and user trust.

Recommended signing path:

1. Use Azure Trusted Signing / Azure Artifact Signing, matching the Tiny Clips release pattern.
2. Add repository secrets for the signing account:

   ```text
   AZURE_CLIENT_ID
   AZURE_TENANT_ID
   AZURE_SUBSCRIPTION_ID
   AZURE_ARTIFACT_SIGNING_ENDPOINT
   AZURE_ARTIFACT_SIGNING_ACCOUNT_NAME
   AZURE_ARTIFACT_SIGNING_CERTIFICATE_PROFILE_NAME
   ```

3. Give the Azure identity the certificate profile signer role.
4. Sign both Windows binaries with `signtool sign /fd SHA256 /tr http://timestamp.acs.microsoft.com /td SHA256 ...`.
5. Verify both binaries with `signtool verify /pa /v`.

Alternative: use a traditional code-signing `.pfx` certificate in GitHub Actions secrets and sign with `signtool`. Azure signing is preferable because the private key is not exported into CI.

## GitHub setup required

Add these repository secrets for `.github/workflows/release.yml`:

```text
AZURE_CLIENT_ID
AZURE_TENANT_ID
AZURE_SUBSCRIPTION_ID
AZURE_ARTIFACT_SIGNING_ENDPOINT
AZURE_ARTIFACT_SIGNING_ACCOUNT_NAME
AZURE_ARTIFACT_SIGNING_CERTIFICATE_PROFILE_NAME
```

Because the release workflow runs from tags and does not use a GitHub environment, the Entra app registration needs a federated credential whose subject matches the release tag ref. For the current release:

```text
repo:burkeholland/resize-me:ref:refs/tags/v0.2.0-windows
```

If you keep OIDC authentication, add the matching federated credential when cutting each new Windows release tag, or configure a broader tag-matching federated credential if your tenant supports flexible federated identity credentials.

Add this repository secret for `.github/workflows/winget-submit.yml`:

```text
WINGET_CREATE_GITHUB_TOKEN
```

`WINGET_CREATE_GITHUB_TOKEN` should be a classic GitHub PAT from the maintainer account with `public_repo` scope. `wingetcreate` uses it to fork `microsoft/winget-pkgs` and open the package PR.

## Release workflow

Keep the current PR smoke workflows for validation. The Windows release workflow is tag-gated, matching macOS: pushing `v*.*.*-windows` is what ships the GitHub Release.

1. Trigger on `v*.*.*-windows` tags.
2. Parse the tag into a numeric package version.
3. Build both architectures:

   ```pwsh
   cd ResizeMe
   wails build -platform windows/amd64 -o ResizeMe-windows-amd64.exe
   wails build -platform windows/arm64 -o ResizeMe-windows-arm64.exe
   ```

4. Sign both files in `ResizeMe\build\bin`.
5. Verify signatures.
6. Compute winget hashes:

   ```pwsh
   winget hash ResizeMe\build\bin\ResizeMe-windows-amd64.exe
   winget hash ResizeMe\build\bin\ResizeMe-windows-arm64.exe
   ```

7. Generate a winget manifest artifact with the version, release URLs, and hashes.
8. Create the GitHub Release and attach:

   ```text
   ResizeMe-windows-amd64.exe
   ResizeMe-windows-arm64.exe
   winget manifest YAML files
   ```

Do not create Windows releases from every push to `main`. Winget should use a deliberate semantic Windows release tag so the package version, release notes, and manifest PR are easy to review.

## Winget manifests

Source manifests live in:

```text
ResizeMe/packaging/winget/
```

The release and winget submission workflows copy the version and locale manifests, stamp `PackageVersion`, and regenerate the installer manifest with the actual release URLs and hashes.

### `BurkeHolland.ResizeMe.yaml`

```yaml
# yaml-language-server: $schema=https://aka.ms/winget-manifest.version.1.12.0.schema.json
PackageIdentifier: BurkeHolland.ResizeMe
PackageVersion: 0.0.1
DefaultLocale: en-US
ManifestType: version
ManifestVersion: 1.12.0
```

### `BurkeHolland.ResizeMe.locale.en-US.yaml`

```yaml
# yaml-language-server: $schema=https://aka.ms/winget-manifest.defaultLocale.1.12.0.schema.json
PackageIdentifier: BurkeHolland.ResizeMe
PackageVersion: 0.0.1
PackageLocale: en-US
Publisher: Burke Holland
PublisherUrl: https://github.com/burkeholland
PublisherSupportUrl: https://github.com/burkeholland/resize-me/issues
PackageName: ResizeMe
PackageUrl: https://github.com/burkeholland/resize-me
License: Proprietary
ShortDescription: Resize focused windows to selected presets with a global hotkey.
Description: |-
  ResizeMe is a lightweight window resizing utility for quickly snapping focused
  windows to saved sizes and positions with a global hotkey.
Tags:
  - window-resizer
  - window-manager
  - productivity
  - wails
ManifestType: defaultLocale
ManifestVersion: 1.12.0
```

Add `PrivacyUrl` once the site has a stable privacy page. If ResizeMe gets a committed open-source license file later, update `License` and add `LicenseUrl`.

### `BurkeHolland.ResizeMe.installer.yaml`

```yaml
# yaml-language-server: $schema=https://aka.ms/winget-manifest.installer.1.12.0.schema.json
PackageIdentifier: BurkeHolland.ResizeMe
PackageVersion: 0.0.1
MinimumOSVersion: 10.0.19041.0
InstallerType: portable
Commands:
  - resizeme
Dependencies:
  PackageDependencies:
    - PackageIdentifier: Microsoft.EdgeWebView2Runtime
Installers:
  - Architecture: x64
    InstallerUrl: https://github.com/burkeholland/resize-me/releases/download/v0.0.1-windows/ResizeMe-windows-amd64.exe
    InstallerSha256: <x64 sha256 from winget hash>
  - Architecture: arm64
    InstallerUrl: https://github.com/burkeholland/resize-me/releases/download/v0.0.1-windows/ResizeMe-windows-arm64.exe
    InstallerSha256: <arm64 sha256 from winget hash>
ManifestType: installer
ManifestVersion: 1.12.0
```

Validate the checked-in templates:

```pwsh
winget validate --manifest ResizeMe\packaging\winget
```

After a tagged Windows release exists, test install with the generated manifest artifact from the release:

```pwsh
winget install --manifest <generated-winget-manifest-folder>
```

## Winget submission workflow

`.github/workflows/winget-submit.yml` is manually triggered after the tagged Windows GitHub Release exists:

1. Input: `tag`, for example `v0.0.1-windows`.
2. Download `ResizeMe-windows-amd64.exe` and `ResizeMe-windows-arm64.exe` from that GitHub Release.
3. Copy the source winget manifests into a temporary `manifests` folder.
4. Replace `PackageVersion` in all manifests.
5. Compute `InstallerSha256` for each `.exe`.
6. Generate the installer manifest with `InstallerType: portable` and a `resizeme` command alias.
7. Run `winget validate --manifest manifests`.
8. Submit:

   ```pwsh
   curl.exe -JLO https://aka.ms/wingetcreate/latest
   .\wingetcreate.exe submit --token $env:WINGET_CREATE_GITHUB_TOKEN manifests
   ```

Required repository secret:

```text
WINGET_CREATE_GITHUB_TOKEN
```

Use a classic GitHub PAT from the maintainer account with `public_repo` scope. `wingetcreate` uses it to fork `microsoft/winget-pkgs` and open the package PR.

## Release checklist

- [ ] Confirm `BurkeHolland.ResizeMe` as the winget package identifier.
- [ ] Use Windows release tags like `v0.0.1-windows`, matching the macOS `v0.0.1-mac` pattern.
- [ ] Add Azure signing secrets.
- [ ] Add the `WINGET_CREATE_GITHUB_TOKEN` secret.
- [ ] Create one test release, validate `winget install --manifest`, then submit to `microsoft/winget-pkgs`.

## Reference model

Use Tiny Clips as the workflow model, but remove the MSIX-specific pieces:

- Keep: tag parsing, GitHub Release download, manifest generation, `winget validate`, `wingetcreate submit`, `WINGET_CREATE_GITHUB_TOKEN`.
- Drop: MSIX package generation, `PackageFamilyName`, `SignatureSha256`, MSIX framework dependencies, and WACK.
- Replace: MSIX signing with Authenticode signing for the Wails `.exe` assets.
