import AppKit
import SwiftUI

struct OnboardingView: View {
    @EnvironmentObject var appState: AppState

    var body: some View {
        VStack(spacing: 0) {
            hero

            VStack(spacing: 14) {
                permissionCard
                launchAtLoginCard
            }
            .padding(.horizontal, 28)
            .padding(.top, 22)

            Spacer(minLength: 16)

            footer
        }
        .frame(minWidth: 480, minHeight: 520)
        .background(.background)
        .onAppear {
            appState.permissionService.refresh()
        }
    }

    // MARK: - Hero

    private var hero: some View {
        VStack(spacing: 12) {
            Image(nsImage: NSApp.applicationIconImage)
                .resizable()
                .interpolation(.high)
                .frame(width: 96, height: 96)
                .padding(.top, 28)

            Text("Welcome to ResizeMe")
                .font(.system(size: 26, weight: .bold, design: .rounded))

            Text("Resize any window to exact dimensions\nwith a keyboard shortcut.")
                .font(.callout)
                .foregroundStyle(.secondary)
                .multilineTextAlignment(.center)
        }
        .frame(maxWidth: .infinity)
        .padding(.bottom, 8)
    }

    // MARK: - Steps

    private var permissionCard: some View {
        StepCard(number: 1, title: "Grant Accessibility permission") {
            VStack(alignment: .leading, spacing: 12) {
                Text("ResizeMe uses macOS Accessibility APIs to move and resize windows of other apps. Nothing is recorded or transmitted.")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .fixedSize(horizontal: false, vertical: true)

                HStack(spacing: 10) {
                    if appState.permissionService.isTrusted {
                        Label("Permission granted", systemImage: "checkmark.circle.fill")
                            .font(.subheadline.weight(.medium))
                            .foregroundStyle(.green)
                            .transition(.opacity.combined(with: .scale(scale: 0.9)))
                    } else {
                        Button("Grant Permission…") {
                            appState.permissionService.requestPermission()
                        }
                        .buttonStyle(.borderedProminent)

                        Button("Open System Settings") {
                            appState.permissionService.openSystemSettings()
                        }

                        Spacer()

                        Label("Not granted", systemImage: "exclamationmark.circle.fill")
                            .font(.subheadline)
                            .foregroundStyle(.orange)
                    }
                }
                .animation(.spring(duration: 0.3), value: appState.permissionService.isTrusted)
            }
        }
    }

    private var launchAtLoginCard: some View {
        StepCard(number: 2, title: "Launch at login", subtitle: "Optional") {
            Toggle(isOn: Binding(
                get: { appState.config.autoStart },
                set: { value in
                    var next = appState.config
                    next.autoStart = value
                    _ = appState.saveConfig(next)
                }
            )) {
                Text("Start ResizeMe automatically when you log in.")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }
            .toggleStyle(.switch)
            .controlSize(.small)
        }
    }

    // MARK: - Footer

    private var footer: some View {
        VStack(spacing: 10) {
            Button {
                appState.completeFirstRun()
                NSApp.keyWindow?.close()
            } label: {
                Text("Get Started")
                    .font(.body.weight(.semibold))
                    .frame(maxWidth: .infinity)
            }
            .buttonStyle(.borderedProminent)
            .controlSize(.large)
            .keyboardShortcut(.defaultAction)

            Text("You can change everything later in Settings.")
                .font(.caption)
                .foregroundStyle(.tertiary)
        }
        .padding(.horizontal, 28)
        .padding(.bottom, 24)
    }
}

// MARK: - Step card container

private struct StepCard<Content: View>: View {
    let number: Int
    let title: String
    var subtitle: String? = nil
    @ViewBuilder var content: Content

    var body: some View {
        VStack(alignment: .leading, spacing: 12) {
            HStack(spacing: 10) {
                Text("\(number)")
                    .font(.subheadline.weight(.bold))
                    .foregroundStyle(.white)
                    .frame(width: 24, height: 24)
                    .background(Circle().fill(Color.accentColor))

                Text(title)
                    .font(.headline)

                if let subtitle {
                    Text(subtitle)
                        .font(.caption.weight(.medium))
                        .foregroundStyle(.secondary)
                        .padding(.horizontal, 7)
                        .padding(.vertical, 2)
                        .background(Capsule().fill(.quaternary.opacity(0.6)))
                }

                Spacer(minLength: 0)
            }

            content
        }
        .padding(16)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(
            RoundedRectangle(cornerRadius: 12, style: .continuous)
                .fill(.quaternary.opacity(0.4))
        )
        .overlay(
            RoundedRectangle(cornerRadius: 12, style: .continuous)
                .strokeBorder(.separator.opacity(0.5), lineWidth: 1)
        )
    }
}
