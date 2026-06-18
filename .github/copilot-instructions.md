# ResizeMe repository instructions

This repository contains two related apps:
- `ResizeMe/` — the original Go/Wails app
- `ResizeMeMac/` — the native Swift menu-bar app for macOS

## Generic repository rules
- If a request does not clearly say which app is meant, ask which project to work on before making changes.
- If the user says “the app”, “the windows app”, or otherwise leaves the target ambiguous, prompt for clarification instead of guessing.
- Prefer the existing project structure and current naming patterns over introducing new abstractions.
- Keep changes surgical and consistent with the surrounding code.
- Update docs when behavior or setup changes.
- Use the app-specific guidance files in `.github/instructions/` when the task is focused on one implementation path.

## App-specific guidance
- Use `.github/instructions/macos-swift.instructions.md` for native macOS Swift work in `ResizeMeMac/`.
- Use `.github/instructions/go-wails.instructions.md` for the original Go/Wails implementation in `ResizeMe/`.
