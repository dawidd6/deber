# frozen_string_literal: true

require "docker/image"

describe Deber::Docker::Image do
  subject do
    Deber::Docker::Image.new "deber", "test"
  end

  it "checks if image is not built" do
    expect(subject.built?).to eq(false)
  end

  it "builds an image" do
    dockerfile = <<~EODF
      FROM alpine

      RUN echo file > /test
    EODF
    status = subject.build(dockerfile) do |line|
      expect(line).not_to eq(nil)
      expect(line.empty?).to eq(false)
    end
    expect(status.success?).to eq(true)
  end

  it "checks if image is built" do
    expect(subject.built?).to eq(true)
  end

  it "gets build date of image" do
    expect(subject.build_date).to be_instance_of(DateTime)
  end

  it "removes an image" do
    expect(subject.remove.success?).to eq(true)
  end

  after :all do
    `docker rmi -f name:tag 2>/dev/null`
  end
end
