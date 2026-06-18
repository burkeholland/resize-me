import Foundation

struct Preset: Codable, Equatable, Identifiable, Sendable {
    var id: String
    var name: String
    var width: Int
    var height: Int
}
