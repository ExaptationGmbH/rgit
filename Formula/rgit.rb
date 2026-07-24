# Homebrew formula for rgit.
#
# Usage once published to a tap (e.g. github.com/ExaptationGmbH/homebrew-tap):
#   brew install ExaptationGmbH/tap/rgit
#
# To install straight from a local checkout while developing:
#   brew install --build-from-source ./Formula/rgit.rb
#
# When you cut a release, update `url` to the release tarball and set `sha256`
# to `shasum -a 256 <tarball>`. The `head` block lets users install the tip
# with `brew install --HEAD rgit` without a release.
class Rgit < Formula
  desc "Run a git command across every repo beneath the current directory"
  homepage "https://github.com/ExaptationGmbH/rgit"
  url "https://github.com/ExaptationGmbH/rgit/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "REPLACE_WITH_TARBALL_SHA256"
  license "MIT"
  head "https://github.com/ExaptationGmbH/rgit.git", branch: "main"

  depends_on "go" => :build

  def install
    ldflags = "-s -w -X main.version=#{version}"
    system "go", "build", *std_go_args(ldflags: ldflags)
  end

  test do
    # In an empty dir there are no repos; rgit should say so and exit non-zero.
    output = shell_output("#{bin}/rgit status 2>&1", 1)
    assert_match "no git repositories found", output

    assert_match version.to_s, shell_output("#{bin}/rgit --version")
  end
end
