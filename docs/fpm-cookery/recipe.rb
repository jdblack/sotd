class  Sotd < FPM::Cookery::Recipe
  homepage    "https://github.com/jdblack/sotd"

  name        "sotd"
  version     "0.2.2"
  description "Song of the day bot"
  maintainer  "James Blackwell <james.blackwell@crowdstrike.com>"

  source      "https://github.com/jdblack/sotd.git", :with => :git, :tag => "v#{version}"

  post_install "post-install.sh"


  def build
    safesystem "go build ."
  end

  def install
    prefix('local/bin/').install 'sotd'
    etc('systemd/system').install workdir('systemd.service')
  end
end


