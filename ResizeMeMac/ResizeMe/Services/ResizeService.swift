import AppKit
import ApplicationServices

@preconcurrency import Foundation

enum ResizeError: Error, Equatable {
    case permissionMissing
    case noFrontmostApp
    case noResizableWindow
    case windowFullscreen
    case windowMinimized
    case resizeRejected
}

struct ResizeOutcome: Equatable {
    let requested: CGSize
    let achieved: CGSize

    var isExact: Bool {
        abs(requested.width - achieved.width) < 1 && abs(requested.height - achieved.height) < 1
    }
}

@MainActor
final class ResizeService {
    private func copyAttribute(_ element: AXUIElement, _ attribute: String) -> CFTypeRef? {
        var value: CFTypeRef?
        let status = AXUIElementCopyAttributeValue(element, attribute as CFString, &value)
        return status == .success ? value : nil
    }

    private func point(from value: AXValue) -> CGPoint? {
        var point = CGPoint.zero
        return AXValueGetValue(value, .cgPoint, &point) ? point : nil
    }

    private func size(from value: AXValue) -> CGSize? {
        var size = CGSize.zero
        return AXValueGetValue(value, .cgSize, &size) ? size : nil
    }

    private func axValue(for point: CGPoint) -> AXValue? {
        var value = point
        return AXValueCreate(.cgPoint, &value)
    }

    private func axValue(for size: CGSize) -> AXValue? {
        var value = size
        return AXValueCreate(.cgSize, &value)
    }

    private func write(_ element: AXUIElement, attribute: String, value: AXValue) -> AXError {
        AXUIElementSetAttributeValue(element, attribute as CFString, value)
    }

    private func apply(position: CGPoint, size: CGSize, to element: AXUIElement, positionFirst: Bool) -> AXError {
        guard let positionValue = axValue(for: position), let sizeValue = axValue(for: size) else {
            return .attributeUnsupported
        }

        if positionFirst {
            let posError = write(element, attribute: kAXPositionAttribute as String, value: positionValue)
            if posError != .success { return posError }
            return write(element, attribute: kAXSizeAttribute as String, value: sizeValue)
        }

        let sizeError = write(element, attribute: kAXSizeAttribute as String, value: sizeValue)
        if sizeError != .success { return sizeError }
        return write(element, attribute: kAXPositionAttribute as String, value: positionValue)
    }

    func resizeFrontmostWindow(to preset: Preset, center: Bool) throws -> ResizeOutcome {
        guard AXIsProcessTrusted() else {
            throw ResizeError.permissionMissing
        }

        guard let app = NSWorkspace.shared.frontmostApplication else {
            throw ResizeError.noFrontmostApp
        }

        let appElement = AXUIElementCreateApplication(app.processIdentifier)

        var windowElement: AXUIElement?
        if let focused = copyAttribute(appElement, kAXFocusedWindowAttribute as String), CFGetTypeID(focused as CFTypeRef) == AXUIElementGetTypeID() {
            windowElement = (focused as! AXUIElement)
        } else if let windows = copyAttribute(appElement, kAXWindowsAttribute as String), CFGetTypeID(windows as CFTypeRef) == CFArrayGetTypeID() {
            let list = windows as! CFArray
            let values = (0..<CFArrayGetCount(list)).compactMap { index -> AXUIElement? in
                let value = CFArrayGetValueAtIndex(list, index)
                guard CFGetTypeID(value as CFTypeRef) == AXUIElementGetTypeID() else { return nil }
                return (value as! AXUIElement)
            }
            windowElement = values.first(where: { element in
                let role = copyAttribute(element, kAXRoleAttribute as String) as? String
                let subrole = copyAttribute(element, kAXSubroleAttribute as String) as? String
                return role == kAXWindowRole as String && (subrole == nil || subrole == "AXStandardWindow")
            })
        }

        guard let windowElement else {
            throw ResizeError.noResizableWindow
        }

        if let minimized = copyAttribute(windowElement, kAXMinimizedAttribute as String) as? Bool, minimized {
            throw ResizeError.windowMinimized
        }

        if let fullscreen = (copyAttribute(windowElement, "AXFullScreen") as? Bool) {
            if fullscreen { throw ResizeError.windowFullscreen }
        }

        guard let currentPosValue = copyAttribute(windowElement, kAXPositionAttribute as String), CFGetTypeID(currentPosValue as CFTypeRef) == AXValueGetTypeID(),
              let currentSizeValue = copyAttribute(windowElement, kAXSizeAttribute as String), CFGetTypeID(currentSizeValue as CFTypeRef) == AXValueGetTypeID(),
              let currentPos = point(from: currentPosValue as! AXValue),
              let currentSize = size(from: currentSizeValue as! AXValue) else {
            throw ResizeError.noResizableWindow
        }

        let screens = WindowGeometryService.currentScreens()
        let primaryMaxY = WindowGeometryService.primaryReferenceMaxY(screens: screens)
        let currentAX = CGRect(origin: currentPos, size: currentSize)
        let currentAppKit = WindowGeometryService.axToAppKit(currentAX, primaryMaxY: primaryMaxY)
        let target = WindowGeometryService.targetRect(
            currentAppKitRect: currentAppKit,
            presetWidth: preset.width,
            presetHeight: preset.height,
            center: center,
            screens: screens
        )
        let targetAX = WindowGeometryService.appKitToAX(target, primaryMaxY: primaryMaxY)

        let requested = CGSize(width: CGFloat(preset.width), height: CGFloat(preset.height))
        let currentArea = currentSize.width * currentSize.height
        let targetArea = targetAX.width * targetAX.height
        let positionFirst = targetArea >= currentArea

        let firstError = apply(position: targetAX.origin, size: targetAX.size, to: windowElement, positionFirst: positionFirst)
        if firstError != AXError.success {
            throw ResizeError.resizeRejected
        }

        let reread = copyAttribute(windowElement, kAXSizeAttribute as String)
        let actualSize = (reread.flatMap { value -> CGSize? in
            guard CFGetTypeID(value as CFTypeRef) == AXValueGetTypeID() else { return nil }
            return size(from: value as! AXValue)
        }) ?? targetAX.size
        let shouldRetry = abs(actualSize.width - requested.width) >= 1 || abs(actualSize.height - requested.height) >= 1

        if shouldRetry {
            let retryError = apply(position: targetAX.origin, size: targetAX.size, to: windowElement, positionFirst: !positionFirst)
            if retryError != AXError.success {
                throw ResizeError.resizeRejected
            }
        }

        let finalValue = copyAttribute(windowElement, kAXSizeAttribute as String)
        let finalSize = (finalValue.flatMap { value -> CGSize? in
            guard CFGetTypeID(value as CFTypeRef) == AXValueGetTypeID() else { return nil }
            return size(from: value as! AXValue)
        }) ?? targetAX.size

        Log.resize.notice("\(app.bundleIdentifier ?? "<unknown>") resize requested=\(requested.width)x\(requested.height) achieved=\(finalSize.width)x\(finalSize.height)")

        return ResizeOutcome(requested: requested, achieved: finalSize)
    }
}
