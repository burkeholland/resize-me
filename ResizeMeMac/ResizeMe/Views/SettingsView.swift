import AppKit
import KeyboardShortcuts
import SwiftUI

private enum SettingsTab: String, CaseIterable {
    case general = "General"
    case presets = "Presets"
    case shortcuts = "Shortcuts"
    case updates = "Updates"
    case about = "About"

    var icon: String {
        switch self {
        case .general:
            return "gearshape"
        case .presets:
            return "square.grid.2x2"
        case .shortcuts:
            return "keyboard"
        case .updates:
            return "arrow.triangle.2.circlepath"
        case .about:
            return "info.circle"
        }
    }
}

struct SettingsView: View {
    @EnvironmentObject var appState: AppState

    @State private var draft: AppConfig = .default
    @State private var draftBase: AppConfig = .default
    @State private var draftShortcut: KeyboardShortcuts.Shortcut?
    @State private var shortcutValidationMessage: String?
    @State private var selectedTab: SettingsTab? = .general
    @State private var splitVisibility: NavigationSplitViewVisibility = .all

    private var hasChanges: Bool {
        draft != appState.config
            || draftShortcut != HotkeyMapper.shortcut(fromConfigString: appState.config.hotkey)
    }

    private var tabs: [SettingsTab] {
        if appState.updateService.canCheckForUpdates {
            return SettingsTab.allCases
        }
        return SettingsTab.allCases.filter { $0 != .updates }
    }

    var body: some View {
        NavigationSplitView(columnVisibility: $splitVisibility) {
            List(tabs, id: \.self, selection: $selectedTab) { tab in
                Label(tab.rawValue, systemImage: tab.icon)
                    .tag(tab as SettingsTab?)
            }
            .listStyle(.sidebar)
            .navigationSplitViewColumnWidth(min: 150, ideal: 170, max: 200)
        } detail: {
            VStack(spacing: 0) {
                Group {
                    switch selectedTab ?? .general {
                    case .general:
                        GeneralTab(draft: $draft, appState: appState)
                    case .presets:
                        PresetsTab(draft: $draft)
                    case .shortcuts:
                        ShortcutsTab(
                            draftShortcut: shortcutBinding,
                            validationMessage: shortcutValidationMessage
                        )
                    case .updates:
                        UpdatesTab(appState: appState)
                    case .about:
                        AboutTab(appState: appState)
                    }
                }

                Divider()

                HStack(spacing: 12) {
                    VStack(alignment: .leading, spacing: 2) {
                        Text("Changes are applied only when you save.")
                            .font(.footnote)
                            .foregroundStyle(.secondary)
                        if let status = appState.lastStatusMessage {
                            Label(status, systemImage: "checkmark.circle")
                                .font(.footnote)
                                .foregroundStyle(.secondary)
                                .lineLimit(1)
                        }
                    }

                    Spacer()

                    Button("Revert") {
                        revertDraft()
                    }
                    .disabled(!hasChanges)

                    Button("Save") {
                        if appState.config != draftBase {
                            appState.lastStatusMessage = "Settings changed elsewhere. Review latest values before saving."
                            revertDraft()
                            return
                        }
                        var next = draft
                        // Cleared shortcut falls back to the default hotkey via normalization.
                        next.hotkey = HotkeyMapper.configString(from: draftShortcut)
                        if appState.saveConfig(next) {
                            revertDraft()
                        }
                    }
                    .buttonStyle(.borderedProminent)
                    .keyboardShortcut(.defaultAction)
                    .disabled(!hasChanges || shortcutValidationMessage != nil)
                }
                .padding(.horizontal, 16)
                .padding(.vertical, 12)
                .background(.bar)
            }
        }
        .navigationSplitViewStyle(.balanced)
        .frame(minWidth: 620, minHeight: 460)
        .onAppear {
            revertDraft()
            if selectedTab == .updates && !appState.updateService.canCheckForUpdates {
                selectedTab = .general
            }
        }
    }

    private func revertDraft() {
        draft = appState.config
        draftBase = appState.config
        draftShortcut = HotkeyMapper.shortcut(fromConfigString: appState.config.hotkey)
        shortcutValidationMessage = nil
    }

    private var shortcutBinding: Binding<KeyboardShortcuts.Shortcut?> {
        Binding(
            get: { draftShortcut },
            set: { shortcut in
                guard let shortcut else {
                    draftShortcut = nil
                    shortcutValidationMessage = nil
                    return
                }

                let configString = HotkeyMapper.configString(from: shortcut)
                guard ConfigNormalizer.isValidHotkeyText(configString) else {
                    shortcutValidationMessage = ConfigNormalizer.hotkeyValidationMessage(configString)
                        ?? "Choose A-Z, 0-9, or F1-F20 with at least one modifier."
                    return
                }

                draftShortcut = shortcut
                shortcutValidationMessage = nil
            }
        )
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
    @State private var hiddenPresetsExpanded = false
    @State private var actionMessage: String?
    @FocusState private var focusedPresetAction: String?

    private var visiblePresetIndexes: [Int] {
        draft.presets.indices.filter { !draft.isPresetHidden(id: draft.presets[$0].id) }
    }

    private var hiddenPresetIndexes: [Int] {
        draft.presets.indices.filter { draft.isPresetHidden(id: draft.presets[$0].id) }
    }

    private var selectedPresetIsOnlyVisible: Bool {
        guard let selection, !draft.isPresetHidden(id: selection) else {
            return false
        }
        return draft.visiblePresets.count <= 1
    }

    var body: some View {
        VStack(spacing: 0) {
            List(selection: $selection) {
                Section("Visible presets") {
                    ForEach(visiblePresetIndexes, id: \.self) { index in
                        PresetRow(
                            preset: $draft.presets[index],
                            isActive: draft.presets[index].id == draft.activePresetId,
                            makeActive: { draft.activePresetId = draft.presets[index].id },
                            isFavorite: draft.favoritePresetIds.contains(draft.presets[index].id),
                            toggleFavorite: { toggleFavorite(draft.presets[index].id) },
                            hide: { hidePreset(draft.presets[index].id) },
                            delete: { deletePresetIDs([draft.presets[index].id]) },
                            canHide: draft.visiblePresets.count > 1,
                            focusedPresetAction: $focusedPresetAction
                        )
                        .tag(draft.presets[index].id)
                    }
                    .onDelete { offsets in
                        deletePresetIDs(offsets.map { visiblePresetIndexes[$0] }.map { draft.presets[$0].id })
                    }
                }

                Section {
                    DisclosureGroup("Hidden presets (\(hiddenPresetIndexes.count))", isExpanded: $hiddenPresetsExpanded) {
                        if hiddenPresetIndexes.isEmpty {
                            Text("No presets are hidden.")
                                .foregroundStyle(.secondary)
                        } else {
                            ForEach(hiddenPresetIndexes, id: \.self) { index in
                                let preset = draft.presets[index]
                                HiddenPresetRow(
                                    preset: preset,
                                    isFavorite: draft.favoritePresetIds.contains(preset.id),
                                    show: { showPreset(preset.id) },
                                    delete: { deletePresetIDs([preset.id]) },
                                    focusedPresetAction: $focusedPresetAction
                                )
                                .tag(preset.id)
                            }
                        }
                    }
                } footer: {
                    Text("Hidden presets remain saved but do not appear in resize selectors or quick menus.")
                }
            }
            .listStyle(.inset)
            .alternatingRowBackgrounds()

            if let actionMessage {
                Label(actionMessage, systemImage: "checkmark.circle")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.horizontal, 12)
                    .padding(.vertical, 6)
                    .accessibilityLabel(actionMessage)
            }

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
                       draft.hasPreset(id: selection) {
                        deletePresetIDs([selection])
                    }
                } label: {
                    Image(systemName: "minus")
                }
                .disabled(selection == nil || selectedPresetIsOnlyVisible)
                .help("Delete the selected preset permanently")

                Spacer()

                Text("Preset dimensions are in logical points (pt), not physical pixels.")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
            }
            .buttonStyle(.borderless)
            .padding(.horizontal, 12)
            .padding(.vertical, 8)
            .background(.bar)
        }
    }

    private func hidePreset(_ id: String) {
        guard let preset = draft.findPreset(id: id),
              draft.visiblePresets.contains(where: { $0.id == id }) else {
            actionMessage = "Preset is no longer available."
            return
        }
        guard draft.visiblePresets.count > 1 else {
            actionMessage = "At least one visible preset is required."
            return
        }

        let replacement = draft.visiblePresets.first(where: { $0.id != id })!
        draft.hiddenPresetIds.append(id)
        hiddenPresetsExpanded = true
        if draft.activePresetId == id {
            draft.activePresetId = replacement.id
            actionMessage = "Hidden \(preset.name). \(replacement.name) is now the active preset."
        } else {
            actionMessage = "Hidden \(preset.name)."
        }
        selection = replacement.id
        DispatchQueue.main.async {
            focusedPresetAction = "show-\(id)"
        }
    }

    private func showPreset(_ id: String) {
        guard let preset = draft.findPreset(id: id) else {
            actionMessage = "Preset no longer exists."
            return
        }

        draft.hiddenPresetIds.removeAll(where: { $0 == id })
        actionMessage = "Restored \(preset.name)."
        selection = id
        DispatchQueue.main.async {
            focusedPresetAction = "hide-\(id)"
        }
    }

    private func deletePresetIDs(_ ids: [String]) {
        let removedIDs = Set(ids.filter { draft.hasPreset(id: $0) })
        guard !removedIDs.isEmpty else {
            actionMessage = "Preset no longer exists."
            return
        }

        let remainingVisible = draft.visiblePresets.filter { !removedIDs.contains($0.id) }
        guard !remainingVisible.isEmpty else {
            actionMessage = "At least one visible preset is required."
            return
        }

        let removedActivePreset = removedIDs.contains(draft.activePresetId)
        draft.presets.removeAll(where: { removedIDs.contains($0.id) })
        draft.hiddenPresetIds.removeAll(where: { removedIDs.contains($0) })
        draft.favoritePresetIds.removeAll(where: { removedIDs.contains($0) })
        if removedActivePreset {
            draft.activePresetId = remainingVisible[0].id
        }
        if let selection, removedIDs.contains(selection) {
            self.selection = draft.activePresetId
        }
        actionMessage = removedIDs.count == 1
            ? "Deleted preset permanently."
            : "Deleted \(removedIDs.count) presets permanently."
    }

    private func toggleFavorite(_ id: String) {
        if let index = draft.favoritePresetIds.firstIndex(of: id) {
            draft.favoritePresetIds.remove(at: index)
        } else {
            draft.favoritePresetIds.append(id)
        }
    }
}

private struct PresetRow: View {
    @Binding var preset: Preset
    let isActive: Bool
    let makeActive: () -> Void
    let isFavorite: Bool
    let toggleFavorite: () -> Void
    let hide: () -> Void
    let delete: () -> Void
    let canHide: Bool
    let focusedPresetAction: FocusState<String?>.Binding

    var body: some View {
        HStack(spacing: 10) {
            Button(action: makeActive) {
                Image(systemName: isActive ? "largecircle.fill.circle" : "circle")
                    .foregroundStyle(isActive ? Color.accentColor : Color.secondary)
            }
            .buttonStyle(.plain)
            .help(isActive ? "Active preset" : "Make this the active preset")

            Button(action: toggleFavorite) {
                Image(systemName: isFavorite ? "star.fill" : "star")
                    .foregroundStyle(isFavorite ? Color.yellow : Color.secondary)
            }
            .buttonStyle(.plain)
            .help(isFavorite ? "Remove from favorites" : "Add to favorites")
            .accessibilityLabel(isFavorite ? "Remove \(preset.name) from favorites" : "Add \(preset.name) to favorites")

            TextField("Name", text: $preset.name)
                .textFieldStyle(.roundedBorder)

            TextField("Width (pt)", value: $preset.width, format: .number.grouping(.never))
                .textFieldStyle(.roundedBorder)
                .frame(width: 64)
                .multilineTextAlignment(.trailing)

            Text("×")
                .foregroundStyle(.secondary)

            TextField("Height (pt)", value: $preset.height, format: .number.grouping(.never))
                .textFieldStyle(.roundedBorder)
                .frame(width: 64)
                .multilineTextAlignment(.trailing)

            Button(action: hide) {
                Image(systemName: "eye.slash")
            }
            .buttonStyle(.plain)
            .disabled(!canHide)
            .help("Hide \(preset.name)")
            .accessibilityLabel("Hide \(preset.name)")
            .focused(focusedPresetAction, equals: "hide-\(preset.id)")

            Button(role: .destructive, action: delete) {
                Image(systemName: "trash")
            }
            .buttonStyle(.plain)
            .help("Delete \(preset.name) permanently")
            .accessibilityLabel("Delete \(preset.name) permanently")
        }
        .padding(.vertical, 2)
    }
}

private struct HiddenPresetRow: View {
    let preset: Preset
    let isFavorite: Bool
    let show: () -> Void
    let delete: () -> Void
    let focusedPresetAction: FocusState<String?>.Binding

    var body: some View {
        HStack(spacing: 10) {
            VStack(alignment: .leading, spacing: 2) {
                Text(preset.name)
                Text("\(preset.width) × \(preset.height) pt")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
                Text(isFavorite ? "Hidden · Favorite" : "Hidden")
                    .font(.footnote)
                    .foregroundStyle(.secondary)
            }

            Spacer()

            Button("Show", action: show)
                .help("Show \(preset.name)")
                .accessibilityLabel("Show \(preset.name)")
                .focused(focusedPresetAction, equals: "show-\(preset.id)")

            Button(role: .destructive, action: delete) {
                Image(systemName: "trash")
            }
            .buttonStyle(.plain)
            .help("Delete \(preset.name) permanently")
            .accessibilityLabel("Delete \(preset.name) permanently")
        }
        .padding(.vertical, 2)
    }
}

// MARK: - Shortcuts

private struct ShortcutsTab: View {
    @Binding var draftShortcut: KeyboardShortcuts.Shortcut?
    let validationMessage: String?

    var body: some View {
        Form {
            Section("Global Shortcut") {
                KeyboardShortcuts.Recorder("Resize frontmost window:", shortcut: $draftShortcut)

                if let validationMessage {
                    Label(validationMessage, systemImage: "exclamationmark.triangle.fill")
                        .font(.footnote)
                        .foregroundStyle(.red)
                }

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

            if appState.updateService.canCheckForUpdates {
                Button("Check for Updates…") {
                    appState.updateService.checkForUpdates()
                }
                .buttonStyle(.bordered)
            }

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
