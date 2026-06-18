import XCTest
@testable import ResizeMe

final class WindowGeometryTests: XCTestCase {
    let main = ScreenInfo(frame: CGRect(x: 0, y: 0, width: 1512, height: 982), visibleFrame: CGRect(x: 0, y: 76, width: 1512, height: 868))
    let right = ScreenInfo(frame: CGRect(x: 1512, y: 0, width: 1920, height: 1080), visibleFrame: CGRect(x: 1512, y: 0, width: 1920, height: 1042))
    let left = ScreenInfo(frame: CGRect(x: -1920, y: 0, width: 1920, height: 1080), visibleFrame: CGRect(x: -1920, y: 0, width: 1920, height: 1042))

    func testPrimaryReferenceIsZeroOriginScreen() {
        XCTAssertEqual(WindowGeometryService.primaryReferenceMaxY(screens: [main, right]), 982)
        XCTAssertEqual(WindowGeometryService.primaryReferenceMaxY(screens: [right, main]), 982)
    }

    func testAxAppKitRoundTrip() {
        let primaryMaxY: CGFloat = 982
        let rect = CGRect(x: 100, y: 50, width: 640, height: 360)
        let appKit = WindowGeometryService.axToAppKit(rect, primaryMaxY: primaryMaxY)
        XCTAssertEqual(appKit.origin.y, 572)
        XCTAssertEqual(WindowGeometryService.appKitToAX(appKit, primaryMaxY: primaryMaxY), rect)

        // A window entirely above the primary display's top edge in AX space
        // (AX y from -500 to -400) must land above primaryMaxY in AppKit space:
        // appKitY = 982 - (-500) - 100 = 1382.
        let above = CGRect(x: 120, y: -500, width: 200, height: 100)
        let aboveAppKit = WindowGeometryService.axToAppKit(above, primaryMaxY: primaryMaxY)
        XCTAssertEqual(aboveAppKit.origin.y, 1382)
        XCTAssertGreaterThan(aboveAppKit.origin.y, primaryMaxY)
        XCTAssertEqual(WindowGeometryService.appKitToAX(aboveAppKit, primaryMaxY: primaryMaxY), above)
    }

    func testScreenContainingGreatestOverlap() {
        let rectMostlyRight = CGRect(x: 1512, y: 0, width: 1600, height: 900)
        XCTAssertEqual(WindowGeometryService.screenContaining(rectMostlyRight, screens: [main, right])?.frame, right.frame)

        let rectMostlyMain = CGRect(x: 100, y: 100, width: 1400, height: 700)
        XCTAssertEqual(WindowGeometryService.screenContaining(rectMostlyMain, screens: [main, right])?.frame, main.frame)
    }

    func testScreenContainingNoOverlapPicksNearest() {
        let rect = CGRect(x: 99999, y: 100, width: 100, height: 100)
        XCTAssertNotNil(WindowGeometryService.screenContaining(rect, screens: [main, right]))
        XCTAssertEqual(WindowGeometryService.screenContaining(rect, screens: [main, right])?.frame, right.frame)
    }

    func testCenterWithinVisibleFrame() {
        let current = CGRect(x: 10, y: 100, width: 400, height: 300)
        let target = WindowGeometryService.targetRect(currentAppKitRect: current, presetWidth: 640, presetHeight: 360, center: true, screens: [main])
        XCTAssertEqual(target.origin.x, 436, accuracy: 0.001)
        XCTAssertEqual(target.origin.y, 330, accuracy: 0.001)
        XCTAssertEqual(target.size, CGSize(width: 640, height: 360))
    }

    func testNonCenterKeepsTopLeft() {
        let current = CGRect(x: 200, y: 300, width: 500, height: 400)
        let target = WindowGeometryService.targetRect(currentAppKitRect: current, presetWidth: 640, presetHeight: 360, center: false, screens: [main])
        XCTAssertEqual(target.origin.x, 200, accuracy: 0.001)
        XCTAssertEqual(target.origin.y, 340, accuracy: 0.001)
    }

    func testOversizedPresetClampsTopVisible() {
        let target = WindowGeometryService.targetRect(currentAppKitRect: CGRect(x: 10, y: 100, width: 400, height: 300), presetWidth: 3840, presetHeight: 2160, center: true, screens: [main])
        XCTAssertEqual(target.size, CGSize(width: 3840, height: 2160))
        XCTAssertEqual(target.origin.x, 0, accuracy: 0.001)
        XCTAssertEqual(target.origin.y, -1216, accuracy: 0.001)
        XCTAssertEqual(target.origin.y + target.size.height, 944, accuracy: 0.001)
    }

    func testNegativeOriginDisplay() {
        let current = CGRect(x: -1800, y: 100, width: 800, height: 600)
        let target = WindowGeometryService.targetRect(currentAppKitRect: current, presetWidth: 1280, presetHeight: 720, center: true, screens: [left, main])
        XCTAssertEqual(target.origin.x, -1600, accuracy: 0.001)
        XCTAssertEqual(target.origin.y, 161, accuracy: 0.001)
    }

    func testEmptyScreensReturnsRequestedSizeAtCurrentOrigin() {
        let current = CGRect(x: 12, y: 34, width: 400, height: 300)
        let target = WindowGeometryService.targetRect(currentAppKitRect: current, presetWidth: 640, presetHeight: 360, center: false, screens: [])
        XCTAssertEqual(target.origin, current.origin)
        XCTAssertEqual(target.size, CGSize(width: 640, height: 360))
    }
}
