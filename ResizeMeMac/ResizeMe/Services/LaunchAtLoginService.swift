import Foundation
import ServiceManagement

@MainActor
final class LaunchAtLoginService {
    var isEnabled: Bool {
        SMAppService.mainApp.status == .enabled
    }

    var requiresApproval: Bool {
        SMAppService.mainApp.status == .requiresApproval
    }

    func setEnabled(_ enabled: Bool) throws {
        if enabled {
            do {
                try SMAppService.mainApp.register()
                Log.loginItem.notice("Launch at login enabled")
            } catch {
                Log.loginItem.error("Failed to enable launch at login: \(error.localizedDescription, privacy: .public)")
                throw error
            }
        } else {
            do {
                try SMAppService.mainApp.unregister()
                Log.loginItem.notice("Launch at login disabled")
            } catch {
                Log.loginItem.error("Failed to disable launch at login: \(error.localizedDescription, privacy: .public)")
                throw error
            }
        }
    }

    func openLoginItemsSettings() {
        SMAppService.openSystemSettingsLoginItems()
    }

    func reconcile(autoStart: Bool) {
        let status = SMAppService.mainApp.status
        if autoStart && status == .notRegistered {
            do {
                try SMAppService.mainApp.register()
                Log.loginItem.notice("Reconciled launch at login: registered")
            } catch {
                Log.loginItem.error("Reconcile register failed: \(error.localizedDescription, privacy: .public)")
            }
        } else if !autoStart && status == .enabled {
            do {
                try SMAppService.mainApp.unregister()
                Log.loginItem.notice("Reconciled launch at login: unregistered")
            } catch {
                Log.loginItem.error("Reconcile unregister failed: \(error.localizedDescription, privacy: .public)")
            }
        }
    }
}
