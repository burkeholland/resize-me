import AppKit
import KeyboardShortcuts
import SwiftUI

struct SettingsView: View {
    @EnvironmentObject var appState: AppState

    @State private var draft: AppConfig = .default
    @State private var draftShortcut: KeyboardShortcuts.Shortcut?

    private var hasChanges: Bool {
        draft != appState.config
            || draftShortcut != HotkeyMapper.shortcut(fromConfigString: appState.config.hotkey)
    }

    var body: some View {
        VStack(spacing: 0) {
            TabView {
                GeneralTab(draft: $draft, appState: appState)
                    .tabItem { Label("General", systemImage: "gearshape") }

                PresetsTab(draft: $draft)
                    .tabItem { Label("Presets", systemImage: "square.grid.2x2") }

                ShortcutsTab(draftShortcut: $draftShortcut)
                    .tabItem { Label("Shortcuts", systemImage: "keyboard") }

                UpdatesTab(appState: appState)
                    .tabItem { Label("Updates", systemImage: "arrow.triangle.2.circlepath") }

                AboutTab(appState: appState)
                    .tabItem { Label("About", systemImage: "info.circle") }
            }
            .frame(width: 520, height: 400)

            Divider()

            HStack(spacing: 12) {
                if let status = appState.lastStatusMessage {
                    Label(status, systemImage: "checkmark.circle")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                        .lineLimit(1)
                }

                Spacer()

                Button("Revert") {
                    revertDraft()
                }
                .disabled(!hasChanges)

                Button("Save") {
                    var next = draft
                    // Cleared shortcut falls back to the default hotkey via normalization.
                    next.hotkey = HotkeyMapper.configString(from: draftShortcut)
                    if appState.saveConfig(next) {
                        revertDraft()
                    }
                }
                .buttonStyle(.borderedProminent)
                .keyboardShortcut(.defaultAction)
                .disabled(!hasChanges)
            }
            .padding(.horizontal, 16)
            .padding(.vertical, 12)
            .background(.bar)
        }
        .onAppear {
            revertDraft()
        }
    }

    private func revertDraft() {
        draft = appState.config
        draftShortcut = HotkeyMapper.shortcut(fromConfigString: appState.config.hotkey)
    }
}

// MARK: - General

private struct GeneralTab: View {
    @Binding var draft: AppConfig
    let appState: AppState

    var body: some View {
        Form {
            Section("Startup") {
                Toggle("Launch at login", isOn: $draft.autoStart)

                if appState.launchAtLoginService.requiresApproval {
                    HStack(alignment: .firstTextBaseline) {
                        Label {
                            Text("Approval is required to enable launch at login.")
                                .font(.subheadline)
                        } icon: {
                            Image(systemName: "exclamationmark.triangle.fill")
                                .foregroundStyle(.yellow)
                        }
                        Spacer()
                        Button("Open Login Items…") {
                            appState.launchAtLoginService.openLoginItemsSettings()
                        }
                        .controlSize(.small)
                    }
                }
            }

            Section("Behavior") {
                Toggle("Center window after resize", isOn: $draft.centerAfterResize)
            }

            if let loadError = appState.loadError {
                Section {
                    Label(loadError, systemImage: "xmark.octagon.fill")
                        .font(.subheadline)
                        .foregroundStyle(.red)
                }
            }
        }
        .formStyle(.grouped)
    }
}

// MARK: - Presets

private struct PresetsTab: View {
    @Binding var draft: AppConfig
    @State private var selection: String?

    var body: some View {
        VStack(spacing: 0) {
            List(selection: $selection) {
                ForEach($draft.presets) { $preset in
                    PresetRow(
                        preset: $preset,
                        isActive: preset.id == draft.activePresetId,
                        makeActive: { draft.activePresetId = preset.id }
                    )
                    .tag(preset.id)
                }
                .onDelete { offsets in
                    deletePresets(at: offsets)
                }
            }
            .listStyle(.inset)
            .alternatingRowBackgrounds()

            Divider()

            HStack(spacing: 8) {
                Button {
                    let preset = Preset(id: "", name: "New Preset", width: 1280, height: 720)
                    draft.presets.append(preset)
                } label: {
                    Image(systemName: "plus")
                }
                .help("Add a preset")

                Button {
                    if let selection,
                       let index = draft.presets.firstIndex(where: { $0.id == selection }) {
                        deletePresets(at: IndexSet(integer: index))
                    }
                } label: {
                    Image(systemName: "minus")
                }
                .disabled(selection == nil)
                .help("Remove the selected preset")

                Spacer()

                Text("Sizes are in points (logical units), not physical pixels.")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
            }
            .buttonStyle(.borderless)
            .padding(.horizontal, 12)
            .padding(.vertical, 8)
            .background(.bar)
        }
    }

    private func deletePresets(at offsets: IndexSet) {
        let removedIds = offsets.map { draft.presets[$0].id }
        draft.presets.remove(atOffsets: offsets)
        if removedIds.contains(selection ?? "") {
            selection = nil
        }
        if removedIds.contains(draft.activePresetId) {
            draft.activePresetId = draft.presets.first?.id ?? ""
        }
    }
}

private struct PresetRow: View {
    @Binding var preset: Preset
    let isActive: Bool
    let makeActive: () -> Void

    var body: some View {
        HStack(spacing: 10) {
            Button(action: makeActive) {
                Image(systemName: isActive ? "largecircle.fill.circle" : "circle")
                    .foregroundStyle(isActive ? Color.accentColor : Color.secondary)
            }
            .buttonStyle(.plain)
            .help(isActive ? "Active preset" : "Make this the active preset")

            TextField("Name", text: $preset.name)
                .textFieldStyle(.roundedBorder)

            TextField("Width", value: $preset.width, format: .number.grouping(.never))
                .textFieldStyle(.roundedBorder)
                .frame(width: 64)
                .multilineTextAlignment(.trailing)

            Text("×")
                .foregroundStyle(.secondary)

            TextField("Height", value: $preset.height, format: .number.grouping(.never))
                .textFieldStyle(.roundedBorder)
                .frame(width: 64)
                .multilineTextAlignment(.trailing)
        }
        .padding(.vertical, 2)
    }
}

// MARK: - Shortcuts

private struct ShortcutsTab: View {
    @Binding var draftShortcut: KeyboardShortcuts.Shortcut?

    var body: some View {
        Form {
            Section("Global Shortcut") {
                KeyboardShortcuts.Recorder("Resize frontmost window:", shortcut: $draftShortcut)

                HStack {
                    Spacer()
                    Button("Restore Default (⌃⌥R)") {
                        draftShortcut = HotkeyMapper.shortcut(fromConfigString: AppConfig.defaultHotkey)
                    }
                    .controlSize(.small)
                }
            }

            Section {
                Label {
                    Text("Global shortcuts may be unavailable while another app has Secure Keyboard Entry active (e.g. password fields, some terminals).")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                } icon: {
                    Image(systemName: "lock.shield")
                        .foregroundStyle(.secondary)
                }
            }
        }
        .formStyle(.grouped)
    }
}

// MARK: - Updates

private struct UpdatesTab: View {
    let appState: AppState

    var body: some View {
        Form {
            Section("Sparkle updates") {
                Label("Automatic update checks are enabled and point at the appcast feed configured in the app bundle.", systemImage: "sparkles")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)

                Button("Check for Updates…") {
                    appState.updateService.checkForUpdates()
                }
                .buttonStyle(.borderedProminent)
            }

            Section("Release notes") {
                Text("The current feed URL is https://burkeholland.github.io/resize-me/appcast.xml. Replace this with your signed release appcast when you publish the first notarized build.")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
            }
        }
        .formStyle(.grouped)
    }
}

// MARK: - About

private struct AboutTab: View {
    let appState: AppState

    private var version: String {
        Bundle.main.object(forInfoDictionaryKey: "CFBundleShortVersionString") as? String ?? "dev"
    }

    var body: some View {
        VStack(spacing: 14) {
            Spacer()

            Image(nsImage: NSApp.applicationIconImage)
                .resizable()
                .interpolation(.high)
                .frame(width: 84, height: 84)

            VStack(spacing: 4) {
                Text("ResizeMe")
                    .font(.system(size: 22, weight: .bold, design: .rounded))
                Text("Version \(version)")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }

            Text("Native macOS port of ResizeMe.")
                .font(.subheadline)
                .foregroundStyle(.secondary)

            Link(destination: URL(string: "https://github.com/burkeholland/resize-me")!) {
                Label("View on GitHub", systemImage: "link")
            }

            Button("Check for Updates…") {
                appState.updateService.checkForUpdates()
            }
            .buttonStyle(.bordered)

            Spacer()

            Text(Bundle.main.object(forInfoDictionaryKey: "NSHumanReadableCopyright") as? String ?? "")
                .font(.caption)
                .foregroundStyle(.tertiary)
                .padding(.bottom, 12)
        }
        .frame(maxWidth: .infinity)
        .padding()
    }
}
