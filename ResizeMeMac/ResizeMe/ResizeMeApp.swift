import SwiftUI

@MainActor
final class AppDelegate: NSObject, NSApplicationDelegate {
    func applicationDidFinishLaunching(_ notification: Notification) {
        NSApp.setActivationPolicy(.accessory)
        if AppState.shared.config.firstRun {
            showOnboardingWindow()
        }
    }

    private var onboardingWindow: NSWindow?

    func showOnboardingWindow() {
        let window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 520, height: 560),
            styleMask: [.titled, .closable],
            backing: .buffered, defer: false
        )
        window.title = "Welcome to ResizeMe"
        window.minSize = NSSize(width: 480, height: 520)
        window.center()
        window.isReleasedWhenClosed = false
        window.contentViewController = NSHostingController(rootView: OnboardingView().environmentObject(AppState.shared))
        window.makeKeyAndOrderFront(nil)
        NSApp.activate(ignoringOtherApps: true)
        onboardingWindow = window
    }
}

@main
struct ResizeMeApp: App {
    @NSApplicationDelegateAdaptor(AppDelegate.self) private var appDelegate
    @StateObject private var appState = AppState.shared

    var body: some Scene {
        MenuBarExtra("ResizeMe", image: "MenuBarIcon") {
            MenuContentView()
                .environmentObject(appState)
        }
        Window("Settings", id: "settings") {
            SettingsView()
                .environmentObject(appState)
        }
        .windowResizability(.contentMinSize)
    }
}
