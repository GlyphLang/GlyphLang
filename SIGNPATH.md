# SignPath Code Signing Setup

This document describes how to set up SignPath for signing Windows executables to avoid SmartScreen warnings.

## Step 1: Apply for SignPath OSS Program

1. Go to https://signpath.io/product/open-source
2. Click "Apply for free"
3. Fill out the application with:
   - Your GitHub repo URL
   - Project description
   - Why you need code signing (SmartScreen warnings)

Approval typically takes 1-3 business days.

## Step 2: After Approval - Configure SignPath

Once approved, you'll get access to the SignPath dashboard. You'll need to:

1. **Create a Project** - link it to your GitHub repo
2. **Create a Signing Policy** - typically "release-signing" for production builds
3. **Install the SignPath GitHub App** on your repo

## Step 3: Update Release Workflow

After signing is set up, modify `.github/workflows/release.yml` to sign the Windows artifacts. Add this job after `windows-installer`:

```yaml
  sign-windows:
    name: Sign Windows Artifacts
    runs-on: ubuntu-latest
    needs: [windows-installer, build]
    steps:
      - name: Download Windows artifacts
        uses: actions/download-artifact@v7
        with:
          path: artifacts

      - name: Submit to SignPath
        uses: signpath/github-action-submit-signing-request@v1
        with:
          api-token: ${{ secrets.SIGNPATH_API_TOKEN }}
          organization-id: ${{ secrets.SIGNPATH_ORG_ID }}
          project-slug: glyphlang
          signing-policy-slug: release-signing
          artifact-configuration-slug: windows
          github-artifact-id: windows-installer
          wait-for-completion: true
          output-artifact-directory: signed-artifacts

      - name: Upload signed installer
        uses: actions/upload-artifact@v6
        with:
          name: windows-installer-signed
          path: signed-artifacts/

      - name: Submit EXE to SignPath
        uses: signpath/github-action-submit-signing-request@v1
        with:
          api-token: ${{ secrets.SIGNPATH_API_TOKEN }}
          organization-id: ${{ secrets.SIGNPATH_ORG_ID }}
          project-slug: glyphlang
          signing-policy-slug: release-signing
          artifact-configuration-slug: windows
          github-artifact-id: glyph-windows-amd64
          wait-for-completion: true
          output-artifact-directory: signed-exe

      - name: Upload signed exe
        uses: actions/upload-artifact@v6
        with:
          name: glyph-windows-amd64-signed
          path: signed-exe/
```

## Step 4: Update upload-assets Job

Modify the `upload-assets` job to depend on signed artifacts:

```yaml
  upload-assets:
    name: Upload Release Assets
    runs-on: ubuntu-latest
    needs: [build, sign-windows]  # Changed from [build, windows-installer]
    steps:
      - name: Download all artifacts
        uses: actions/download-artifact@v7
        with:
          path: artifacts

      - name: List artifacts
        run: find artifacts -type f

      - name: Upload assets to release
        uses: softprops/action-gh-release@v2
        with:
          files: |
            artifacts/glyph-linux-amd64/*.zip
            artifacts/glyph-darwin-amd64/*.zip
            artifacts/glyph-darwin-arm64/*.zip
            artifacts/glyph-windows-amd64-signed/*.zip
            artifacts/windows-installer-signed/*.exe
```

## Step 5: Add Repository Secrets

In your GitHub repository settings, add these secrets:

- `SIGNPATH_API_TOKEN` - Your SignPath API token
- `SIGNPATH_ORG_ID` - Your SignPath organization ID

Both values are available in the SignPath dashboard after approval.

## Verification

After a release, verify signing by:

1. Download the Windows installer or exe
2. Right-click the file and select "Properties"
3. Go to the "Digital Signatures" tab
4. The signature should show as valid with SignPath as the signer

## Resources

- SignPath Documentation: https://docs.signpath.io/
- SignPath GitHub Action: https://github.com/SignPath/github-action-submit-signing-request
- SignPath OSS Program: https://signpath.io/product/open-source
