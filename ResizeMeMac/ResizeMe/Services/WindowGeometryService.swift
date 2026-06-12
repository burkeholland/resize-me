import CoreGraphics
import Foundation

/// A display description in AppKit (bottom-left-origin) global coordinates.
struct ScreenInfo: Equatable {
    /// Full frame in AppKit space.
    var frame: CGRect
    /// Frame excluding menu bar and Dock, in AppKit space.
    var visibleFrame: CGRect
}

enum WindowGeometryService {
    /// Returns the frame.maxY of the zero-origin screen.
    /// If none has origin == .zero, falls back to the first screen's frame.maxY.
    static func primaryReferenceMaxY(screens: [ScreenInfo]) -> CGFloat {
        if let zeroOrigin = screens.first(where: { $0.frame.origin == .zero }) {
            return zeroOrigin.frame.maxY
        }
        return screens.first?.frame.maxY ?? 0
    }

    /// Converts an AX coordinate rect (top-left origin) to AppKit space.
    static func axToAppKit(_ rect: CGRect, primaryMaxY: CGFloat) -> CGRect {
        let y = primaryMaxY - rect.origin.y - rect.height
        return CGRect(x: rect.origin.x, y: y, width: rect.width, height: rect.height)
    }

    /// Converts an AppKit rect to AX coordinate space.
    static func appKitToAX(_ rect: CGRect, primaryMaxY: CGFloat) -> CGRect {
        let y = primaryMaxY - rect.origin.y - rect.height
        return CGRect(x: rect.origin.x, y: y, width: rect.width, height: rect.height)
    }

    /// Returns the screen whose frame has the greatest intersection area with rect.
    /// Falls back to nearest center when there is no overlap.
    static func screenContaining(_ rect: CGRect, screens: [ScreenInfo]) -> ScreenInfo? {
        guard !screens.isEmpty else { return nil }

        let best = screens.max { left, right in
            let leftArea = left.frame.intersection(rect).area
            let rightArea = right.frame.intersection(rect).area
            if leftArea == rightArea {
                return left.frame.center.distance(to: rect.center) < right.frame.center.distance(to: rect.center)
            }
            return leftArea < rightArea
        }

        if let best, best.frame.intersection(rect).area > 0 {
            return best
        }

        return screens.min { left, right in
            left.frame.center.distance(to: rect.center) < right.frame.center.distance(to: rect.center)
        }
    }

    /// Computes the target rect in AppKit space.
    static func targetRect(currentAppKitRect: CGRect,
                           presetWidth: Int,
                           presetHeight: Int,
                           center: Bool,
                           screens: [ScreenInfo]) -> CGRect {
        let size = CGSize(width: CGFloat(presetWidth), height: CGFloat(presetHeight))

        guard let screen = screenContaining(currentAppKitRect, screens: screens) else {
            return CGRect(origin: currentAppKitRect.origin, size: size)
        }

        let vf = screen.visibleFrame
        var origin = currentAppKitRect.origin

        if center {
            origin.x = vf.midX - size.width / 2
            origin.y = vf.midY - size.height / 2
        } else {
            origin.x = currentAppKitRect.origin.x
            origin.y = currentAppKitRect.maxY - size.height
        }

        origin.x = min(max(origin.x, vf.minX), max(vf.maxX - size.width, vf.minX))
        origin.y = min(origin.y, vf.maxY - size.height)
        if origin.y + size.height < vf.minY + 40 {
            origin.y = vf.minY + 40 - size.height
        }

        return CGRect(origin: origin, size: size)
    }
}

private extension CGRect {
    var area: CGFloat {
        width * height
    }

    var center: CGPoint {
        CGPoint(x: midX, y: midY)
    }
}

private extension CGPoint {
    func distance(to other: CGPoint) -> CGFloat {
        hypot(x - other.x, y - other.y)
    }
}

#if canImport(AppKit)
import AppKit

extension WindowGeometryService {
    @MainActor
    static func currentScreens() -> [ScreenInfo] {
        NSScreen.screens.map { ScreenInfo(frame: $0.frame, visibleFrame: $0.visibleFrame) }
    }
}
#endif
