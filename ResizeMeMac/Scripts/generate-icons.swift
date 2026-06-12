#!/usr/bin/swift
// Renders the ResizeMe app icon + menu bar template icon into the asset catalog.
import AppKit

let assetRoot = URL(fileURLWithPath: CommandLine.arguments.count > 1
    ? CommandLine.arguments[1]
    : "ResizeMe/Assets.xcassets")
let appIconDir = assetRoot.appendingPathComponent("AppIcon.appiconset")
let menuIconDir = assetRoot.appendingPathComponent("MenuBarIcon.imageset")

func render(_ pixels: Int, to url: URL, draw: (CGContext, CGFloat) -> Void) {
    let rep = NSBitmapImageRep(
        bitmapDataPlanes: nil, pixelsWide: pixels, pixelsHigh: pixels,
        bitsPerSample: 8, samplesPerPixel: 4, hasAlpha: true, isPlanar: false,
        colorSpaceName: .deviceRGB, bytesPerRow: 0, bitsPerPixel: 0
    )!
    let gctx = NSGraphicsContext(bitmapImageRep: rep)!
    NSGraphicsContext.saveGraphicsState()
    NSGraphicsContext.current = gctx
    let ctx = gctx.cgContext
    ctx.setAllowsAntialiasing(true)
    ctx.interpolationQuality = .high
    draw(ctx, CGFloat(pixels))
    NSGraphicsContext.restoreGraphicsState()
    try! rep.representation(using: .png, properties: [:])!.write(to: url)
    print("wrote \(url.lastPathComponent) (\(pixels)px)")
}

func rgba(_ r: CGFloat, _ g: CGFloat, _ b: CGFloat, _ a: CGFloat = 1) -> CGColor {
    CGColor(srgbRed: r / 255, green: g / 255, blue: b / 255, alpha: a)
}

// Double-headed diagonal arrow along p1 -> p2.
func arrowPath(from p1: CGPoint, to p2: CGPoint, shaft: CGFloat, head: CGFloat) -> CGPath {
    let dx = p2.x - p1.x, dy = p2.y - p1.y
    let len = sqrt(dx * dx + dy * dy)
    let ux = dx / len, uy = dy / len          // unit along
    let nx = -uy, ny = ux                     // unit normal
    let hw = head * 0.62                      // arrowhead half width
    let sw = shaft / 2
    let a1 = CGPoint(x: p1.x + ux * head, y: p1.y + uy * head) // base of head 1
    let a2 = CGPoint(x: p2.x - ux * head, y: p2.y - uy * head) // base of head 2
    let path = CGMutablePath()
    // tip 1
    path.move(to: p1)
    path.addLine(to: CGPoint(x: a1.x + nx * hw, y: a1.y + ny * hw))
    path.addLine(to: CGPoint(x: a1.x + nx * sw, y: a1.y + ny * sw))
    // shaft to head 2
    path.addLine(to: CGPoint(x: a2.x + nx * sw, y: a2.y + ny * sw))
    path.addLine(to: CGPoint(x: a2.x + nx * hw, y: a2.y + ny * hw))
    path.addLine(to: p2)
    path.addLine(to: CGPoint(x: a2.x - nx * hw, y: a2.y - ny * hw))
    path.addLine(to: CGPoint(x: a2.x - nx * sw, y: a2.y - ny * sw))
    path.addLine(to: CGPoint(x: a1.x - nx * sw, y: a1.y - ny * sw))
    path.addLine(to: CGPoint(x: a1.x - nx * hw, y: a1.y - ny * hw))
    path.closeSubpath()
    return path
}

// MARK: - App icon (designed on a 1024pt canvas)

func drawAppIcon(_ ctx: CGContext, size: CGFloat) {
    let s = size / 1024

    // Big Sur style: content squircle inset within transparent canvas.
    let content = CGRect(x: 100 * s, y: 100 * s, width: 824 * s, height: 824 * s)
    let squircle = CGPath(roundedRect: content, cornerWidth: 186 * s, cornerHeight: 186 * s, transform: nil)

    // Soft drop shadow.
    ctx.saveGState()
    ctx.setShadow(offset: CGSize(width: 0, height: -12 * s), blur: 36 * s,
                  color: rgba(10, 20, 60, 0.45))
    ctx.addPath(squircle)
    ctx.setFillColor(rgba(30, 80, 200))
    ctx.fillPath()
    ctx.restoreGState()

    // Gradient background.
    ctx.saveGState()
    ctx.addPath(squircle)
    ctx.clip()
    let bg = CGGradient(colorsSpace: CGColorSpace(name: CGColorSpace.sRGB)!,
                        colors: [rgba(96, 165, 250), rgba(59, 110, 235), rgba(34, 64, 190)] as CFArray,
                        locations: [0, 0.55, 1])!
    ctx.drawLinearGradient(bg,
                           start: CGPoint(x: content.midX, y: content.maxY),
                           end: CGPoint(x: content.midX, y: content.minY),
                           options: [])
    // Subtle top sheen.
    let sheen = CGGradient(colorsSpace: CGColorSpace(name: CGColorSpace.sRGB)!,
                           colors: [rgba(255, 255, 255, 0.28), rgba(255, 255, 255, 0)] as CFArray,
                           locations: [0, 1])!
    ctx.drawLinearGradient(sheen,
                           start: CGPoint(x: content.midX, y: content.maxY),
                           end: CGPoint(x: content.midX, y: content.maxY - 330 * s),
                           options: [])
    ctx.restoreGState()

    // Back "ghost" window — implies the smaller, pre-resize state.
    let backWin = CGRect(x: 244 * s, y: 380 * s, width: 360 * s, height: 280 * s)
    ctx.saveGState()
    ctx.addPath(CGPath(roundedRect: backWin, cornerWidth: 40 * s, cornerHeight: 40 * s, transform: nil))
    ctx.setStrokeColor(rgba(255, 255, 255, 0.42))
    ctx.setLineWidth(22 * s)
    ctx.strokePath()
    ctx.restoreGState()

    // Front window.
    let win = CGRect(x: 244 * s, y: 268 * s, width: 536 * s, height: 432 * s)
    let winPath = CGPath(roundedRect: win, cornerWidth: 52 * s, cornerHeight: 52 * s, transform: nil)
    ctx.saveGState()
    ctx.setShadow(offset: CGSize(width: 0, height: -8 * s), blur: 24 * s, color: rgba(10, 25, 80, 0.35))
    ctx.addPath(winPath)
    ctx.setFillColor(rgba(255, 255, 255, 0.94))
    ctx.fillPath()
    ctx.restoreGState()

    // Title bar.
    let titleH = 88 * s
    ctx.saveGState()
    ctx.addPath(winPath)
    ctx.clip()
    ctx.setFillColor(rgba(226, 234, 248))
    ctx.fill(CGRect(x: win.minX, y: win.maxY - titleH, width: win.width, height: titleH))
    ctx.setFillColor(rgba(180, 196, 224, 0.8))
    ctx.fill(CGRect(x: win.minX, y: win.maxY - titleH - 3 * s, width: win.width, height: 3 * s))
    ctx.restoreGState()

    // Traffic lights.
    let dotR = 17 * s
    let dotY = win.maxY - titleH / 2
    let dotColors = [rgba(255, 95, 87), rgba(255, 189, 46), rgba(40, 200, 64)]
    for (i, c) in dotColors.enumerated() {
        let x = win.minX + 52 * s + CGFloat(i) * 56 * s
        ctx.setFillColor(c)
        ctx.fillEllipse(in: CGRect(x: x - dotR, y: dotY - dotR, width: dotR * 2, height: dotR * 2))
    }

    // Diagonal resize arrow in window body.
    let arrow = arrowPath(
        from: CGPoint(x: 336 * s, y: 348 * s),
        to: CGPoint(x: 692 * s, y: 558 * s),
        shaft: 30 * s, head: 96 * s
    )
    ctx.saveGState()
    ctx.addPath(arrow)
    let arrowGrad = CGGradient(colorsSpace: CGColorSpace(name: CGColorSpace.sRGB)!,
                               colors: [rgba(59, 110, 235), rgba(34, 64, 190)] as CFArray,
                               locations: [0, 1])!
    ctx.clip()
    ctx.drawLinearGradient(arrowGrad,
                           start: CGPoint(x: 336 * s, y: 558 * s),
                           end: CGPoint(x: 692 * s, y: 348 * s),
                           options: [.drawsBeforeStartLocation, .drawsAfterEndLocation])
    ctx.restoreGState()
}

// MARK: - Menu bar template icon (18pt canvas, black + alpha only)

func drawMenuIcon(_ ctx: CGContext, size: CGFloat) {
    let s = size / 18
    let black = CGColor(srgbRed: 0, green: 0, blue: 0, alpha: 1)

    // Window outline.
    let win = CGRect(x: 1.5 * s, y: 2.5 * s, width: 15 * s, height: 12 * s)
    ctx.addPath(CGPath(roundedRect: win, cornerWidth: 2.6 * s, cornerHeight: 2.6 * s, transform: nil))
    ctx.setStrokeColor(black)
    ctx.setLineWidth(1.5 * s)
    ctx.strokePath()

    // Diagonal resize arrow.
    let arrow = arrowPath(
        from: CGPoint(x: 5.1 * s, y: 5.6 * s),
        to: CGPoint(x: 12.9 * s, y: 11.4 * s),
        shaft: 1.4 * s, head: 3.4 * s
    )
    ctx.addPath(arrow)
    ctx.setFillColor(black)
    ctx.fillPath()
}

// MARK: - Emit files

let appSizes: [(String, Int)] = [
    ("icon_16x16@1x.png", 16), ("icon_16x16@2x.png", 32),
    ("icon_32x32@1x.png", 32), ("icon_32x32@2x.png", 64),
    ("icon_128x128@1x.png", 128), ("icon_128x128@2x.png", 256),
    ("icon_256x256@1x.png", 256), ("icon_256x256@2x.png", 512),
    ("icon_512x512@1x.png", 512), ("icon_512x512@2x.png", 1024),
]
for (name, px) in appSizes {
    render(px, to: appIconDir.appendingPathComponent(name), draw: drawAppIcon)
}

render(18, to: menuIconDir.appendingPathComponent("menu_18x18@1x.png"), draw: drawMenuIcon)
render(36, to: menuIconDir.appendingPathComponent("menu_18x18@2x.png"), draw: drawMenuIcon)
