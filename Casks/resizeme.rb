cask "resizeme" do
  version "1.0.1"
  sha256 "6b494c9db4c95c45edda8940d49515b10613ae088edfe4f4fb13b8033d299721"

  url "https://github.com/burkeholland/resize-me/releases/download/v#{version}-mac/ResizeMe.zip"
  name "ResizeMe"
  desc "Menu bar app for pixel-exact window resizing"
  homepage "https://github.com/burkeholland/resize-me"

  auto_updates true
  depends_on macos: :sonoma

  app "ResizeMe.app"

  zap trash: [
    "~/Library/Application Support/ResizeMe",
    "~/Library/Preferences/com.resizeme.mac.plist",
  ]
end
