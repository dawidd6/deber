# frozen_string_literal: true

require_relative "lib/deber/version"

Gem::Specification.new do |spec|
  spec.name          = "deber"
  spec.version       = Deber::VERSION
  spec.authors       = ["Dawid Dziurla"]
  spec.email         = ["dawidd0811@gmail.com"]

  spec.summary       = "Debian packaging with Docker."
  spec.description   = "Debian packaging with Docker."
  spec.homepage      = "https://github.com/dawidd6/deber"
  spec.license       = "MIT"
  spec.required_ruby_version = ">= 2.6"

  # Specify which files should be added to the gem when it is released.
  # The `git ls-files -z` loads the files in the RubyGem that have been added into git.
  spec.files = Dir["lib/**/**"]
  spec.bindir        = "bin"
  spec.executables   = [spec.name]
  spec.require_paths = ["lib"]

  # Uncomment to register a new dependency of your gem
  # spec.add_dependency "example-gem", "~> 1.0"

  # For more information and examples about making a new gem, checkout our
  # guide at: https://bundler.io/guides/creating_gem.html
end
