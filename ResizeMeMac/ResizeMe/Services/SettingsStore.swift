import Foundation
import os.log

final class SettingsStore {
    let fileURL: URL
    private let logger = Logger(subsystem: "com.resizeme.mac", category: "settings")

    struct LoadResult {
        let config: AppConfig
        let loadError: String?
    }

    init(fileURL: URL? = nil) {
        if let fileURL {
            self.fileURL = fileURL
        } else {
            let supportDir = FileManager.default.urls(for: .applicationSupportDirectory, in: .userDomainMask).first ?? FileManager.default.temporaryDirectory
            self.fileURL = supportDir.appendingPathComponent("ResizeMe", isDirectory: true).appendingPathComponent("settings.json")
        }
    }

    func load() -> LoadResult {
        let tempURL = fileURL.appendingPathExtension("tmp")
        try? FileManager.default.removeItem(at: tempURL)

        guard FileManager.default.fileExists(atPath: fileURL.path) else {
            return LoadResult(config: .default, loadError: nil)
        }

        do {
            let data = try Data(contentsOf: fileURL)
            let decoder = JSONDecoder()
            let decoded = try decoder.decode(AppConfig.self, from: data)
            do {
                let normalized = try ConfigNormalizer.normalize(decoded, fallback: .default)
                return LoadResult(config: normalized, loadError: nil)
            } catch {
                logger.error("settings normalization failed: \(error.localizedDescription, privacy: .public)")
                return LoadResult(config: .default, loadError: error.localizedDescription)
            }
        } catch let error as DecodingError {
            logger.error("settings decode failed: \(error.localizedDescription, privacy: .public)")
            return LoadResult(config: .default, loadError: "parse settings: \(error.localizedDescription)")
        } catch {
            logger.error("settings read failed: \(error.localizedDescription, privacy: .public)")
            return LoadResult(config: .default, loadError: error.localizedDescription)
        }
    }

    func save(_ config: AppConfig) throws {
        let parent = fileURL.deletingLastPathComponent()
        try FileManager.default.createDirectory(at: parent, withIntermediateDirectories: true)

        let encoder = JSONEncoder()
        encoder.outputFormatting = [.prettyPrinted, .sortedKeys]
        let data = try encoder.encode(config)

        let tempURL = fileURL.appendingPathExtension("tmp")
        try data.write(to: tempURL, options: .atomic)

        do {
            _ = try FileManager.default.replaceItemAt(fileURL, withItemAt: tempURL)
        } catch {
            logger.error("settings save failed: \(error.localizedDescription, privacy: .public)")
            try? FileManager.default.removeItem(at: tempURL)
            throw error
        }
    }
}
