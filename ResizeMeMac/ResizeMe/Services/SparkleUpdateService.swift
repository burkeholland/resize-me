import Foundation
import Sparkle

@MainActor
final class SparkleUpdateService {
    private let updater: SPUStandardUpdaterController

    init() {
        self.updater = SPUStandardUpdaterController(startingUpdater: true, updaterDelegate: nil, userDriverDelegate: nil)
    }

    func checkForUpdates() {
        updater.checkForUpdates(nil)
    }
}
