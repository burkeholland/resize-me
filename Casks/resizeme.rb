cask "resizeme" do
  version "0.0.1"
  sha256 "65514eaa1c7e265acae48ec29297599e169ef6939ee142d2363a41e8ceef1d7d"

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
