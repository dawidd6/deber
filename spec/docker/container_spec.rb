# frozen_string_literal: true

require "docker/image"
require "docker/container"

describe Deber::Docker::Container do
  subject do
    image = Deber::Docker::Image.new "alpine", "latest"
    Deber::Docker::Container.new "deber-test", image
  end

  it "checks if container is not created" do
    expect(subject.created?).to eq(false)
  end

  it "creates a container" do
    expect(subject.create.success?).to eq(true)
  end

  it "checks if container is created" do
    expect(subject.created?).to eq(true)
  end

  it "removes a container" do
    expect(subject.remove.success?).to eq(true)
  end

  after :all do
    `docker rm -f deber-test 2>/dev/null`
  end
end
