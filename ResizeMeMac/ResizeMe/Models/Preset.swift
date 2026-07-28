import Foundation

struct Preset: Codable, Equatable, Identifiable, Sendable {
    var id: String
    var name: String
    var width: Int
    var height: Int

    static func newDraft() -> Preset {
        Preset(id: UUID().uuidString, name: "New Preset", width: 1280, height: 720)
    }
}
