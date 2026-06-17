import Foundation
import Sparkle

@MainActor
final class SparkleUpdateService {
    private static let keyPlaceholder = "REPLACE_WITH_SPARKLE_PUBLIC_KEY"

    let canCheckForUpdates: Bool
    private let updater: SPUStandardUpdaterController?

    init() {
#if DEBUG
        canCheckForUpdates = false
        updater = nil
#else
        let key = Bundle.main.object(forInfoDictionaryKey: "SUPublicEDKey") as? String
        let hasRealPublicKey = key?
            .trimmingCharacters(in: .whitespacesAndNewlines)
            .isEmpty == false
            && key != Self.keyPlaceholder
        canCheckForUpdates = hasRealPublicKey
        updater = hasRealPublicKey
            ? SPUStandardUpdaterController(startingUpdater: true, updaterDelegate: nil, userDriverDelegate: nil)
            : nil
#endif
    }

    func checkForUpdates() {
        guard canCheckForUpdates, let updater else { return }
        updater.checkForUpdates(nil)
    }
}
