import AppKit
import SwiftUI

struct MenuContentView: View {
    @EnvironmentObject var appState: AppState
    @Environment(\.openWindow) private var openWindow

    private var favoritePresetIDs: Set<String> {
        Set(appState.config.favoritePresetIds.filter { !appState.config.isPresetHidden(id: $0) })
    }

    private var favoritePresets: [Preset] {
        appState.config.favoritePresetIds.compactMap { id in
            guard !appState.config.isPresetHidden(id: id) else {
                return nil
            }
            return appState.config.findPreset(id: id)
        }
    }

    private var otherPresets: [Preset] {
        appState.config.visiblePresets.filter { !favoritePresetIDs.contains($0.id) }
    }

    var body: some View {
        Group {
            if let name = appState.frontmostAppName {
                Text(name)
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            if !appState.permissionService.isTrusted {
                Button("Grant Accessibility Permission…") {
                    appState.permissionService.requestPermission()
                    appState.permissionService.openSystemSettings()
                }
            }

            Button("Resize Now (\(HotkeyMapper.displayString(from: appState.hotkeyService.currentShortcut)))") {
                appState.resizeNow()
            }
            .disabled(!appState.permissionService.isTrusted)

            if let status = appState.lastStatusMessage {
                Text(status)
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            Divider()

            Picker("", selection: Binding(get: {
                appState.config.activePresetId
            }, set: { newValue in
                appState.setActivePreset(newValue)
            })) {
                if !favoritePresets.isEmpty {
                    Section("Favorites") {
                        ForEach(favoritePresets) { preset in
                            Text("\(preset.name) (\(preset.width)×\(preset.height) pt)")
                                .tag(preset.id)
                        }
                    }
                }

                Section("All Presets") {
                    ForEach(otherPresets) { preset in
                        Text("\(preset.name) (\(preset.width)×\(preset.height) pt)")
                            .tag(preset.id)
                    }
                }
            }
            .pickerStyle(.inline)
            .labelsHidden()

            Divider()

            Button("Settings…") {
                NSApp.activate(ignoringOtherApps: true)
                openWindow(id: "settings")
            }
            .keyboardShortcut(",")

            if appState.updateService.canCheckForUpdates {
                Button("Check for Updates…") {
                    appState.updateService.checkForUpdates()
                }
            }

            Button("About ResizeMe") {
                NSApp.activate(ignoringOtherApps: true)
                NSApp.orderFrontStandardAboutPanel(nil)
            }

            Divider()

            Button("Quit ResizeMe") {
                NSApp.terminate(nil)
            }
            .keyboardShortcut("q")
        }
        .onAppear {
            appState.permissionService.refresh()
        }
    }
}
