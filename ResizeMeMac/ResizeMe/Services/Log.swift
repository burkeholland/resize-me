import Foundation
import os.log

enum Log {
    static let subsystem = "com.resizeme.mac"
    static let permissions = Logger(subsystem: subsystem, category: "permissions")
    static let hotkeys = Logger(subsystem: subsystem, category: "hotkeys")
    static let resize = Logger(subsystem: subsystem, category: "resize")
    static let settings = Logger(subsystem: subsystem, category: "settings")
    static let loginItem = Logger(subsystem: subsystem, category: "loginItem")
}
