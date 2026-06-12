import XCTest
@testable import ResizeMe

final class SettingsStoreTests: XCTestCase {
    private func makeStore() -> SettingsStore {
        let url = FileManager.default.temporaryDirectory.appendingPathComponent(UUID().uuidString).appendingPathComponent("settings.json")
        try? FileManager.default.createDirectory(at: url.deletingLastPathComponent(), withIntermediateDirectories: true)
        return SettingsStore(fileURL: url)
    }

    func testLoadMissingFileReturnsDefaults() {
        let store = makeStore()
        let result = store.load()
        XCTAssertEqual(result.config, AppConfig.default)
        XCTAssertNil(result.loadError)
    }

    func testSaveThenLoadRoundTrips() {
        let store = makeStore()
        let config = AppConfig(schemaVersion: 1, presets: [Preset(id: "custom", name: "Custom", width: 800, height: 600)], activePresetId: "custom", centerAfterResize: false, hotkey: "Ctrl+Shift+X", autoStart: false, firstRun: false)
        try? store.save(config)
        let result = store.load()
        XCTAssertEqual(result.config, config)
        XCTAssertNil(result.loadError)
    }

    func testInvalidJSONLoadsDefaultsWithError() {
        let store = makeStore()
        try? "not json {".data(using: .utf8)?.write(to: store.fileURL, options: .atomic)
        let result = store.load()
        XCTAssertEqual(result.config, AppConfig.default)
        XCTAssertNotNil(result.loadError)
    }

    func testGoWrittenFileWithoutSchemaVersionLoads() {
        let store = makeStore()
        let payload = """
        {"presets":[{"id":"a","name":"A","width":800,"height":600}],"activePresetId":"a","centerAfterResize":false,"hotkey":"Ctrl+Alt+R","autoStart":false,"firstRun":false}
        """.data(using: .utf8)!
        try? payload.write(to: store.fileURL, options: .atomic)
        let result = store.load()
        XCTAssertEqual(result.loadError, nil)
        XCTAssertEqual(result.config.presets.count, 1)
        XCTAssertEqual(result.config.presets[0].id, "a")
        XCTAssertEqual(result.config.activePresetId, "a")
        XCTAssertFalse(result.config.centerAfterResize)
        XCTAssertEqual(result.config.schemaVersion, 1)
    }

    func testSaveIsAtomicNoTempLeftBehind() {
        let store = makeStore()
        try? store.save(AppConfig.default)
        XCTAssertFalse(FileManager.default.fileExists(atPath: store.fileURL.appendingPathExtension("tmp").path))
        XCTAssertNoThrow(try Data(contentsOf: store.fileURL))
    }

    func testLeftoverTempFileCleanedOnLoad() {
        let store = makeStore()
        try? "junk".data(using: .utf8)?.write(to: store.fileURL.appendingPathExtension("tmp"), options: .atomic)
        _ = store.load()
        XCTAssertFalse(FileManager.default.fileExists(atPath: store.fileURL.appendingPathExtension("tmp").path))
    }

    func testOutOfRangeDimensionInFileSurfacesError() {
        let store = makeStore()
        let payload = """
        {"schemaVersion":1,"presets":[{"id":"bad","name":"Bad","width":50,"height":600}],"activePresetId":"bad","centerAfterResize":true,"hotkey":"Ctrl+Alt+R","autoStart":false,"firstRun":false}
        """.data(using: .utf8)!
        try? payload.write(to: store.fileURL, options: .atomic)
        let result = store.load()
        XCTAssertEqual(result.config, AppConfig.default)
        XCTAssertNotNil(result.loadError)
        XCTAssertTrue((result.loadError ?? "").contains("width must be between 100 and 10000"))
    }
}
