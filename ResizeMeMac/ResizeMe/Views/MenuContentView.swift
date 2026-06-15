import AppKit
import SwiftUI

struct MenuContentView: View {
    @EnvironmentObject var appState: AppState
    @Environment(\.openWindow) private var openWindow

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

            Picker("Preset", selection: Binding(get: {
                appState.config.activePresetId
            }, set: { newValue in
                appState.setActivePreset(newValue)
            })) {
                ForEach(appState.config.presets) { preset in
                    Text("\(preset.name) (\(preset.width)×\(preset.height))")
                        .tag(preset.id)
                }
            }
            .pickerStyle(.inline)

            Divider()

            Button("Settings…") {
                NSApp.activate(ignoringOtherApps: true)
                openWindow(id: "settings")
            }
            .keyboardShortcut(",")

            Button("Check for Updates…") {
                appState.updateService.checkForUpdates()
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
