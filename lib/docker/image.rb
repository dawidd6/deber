# frozen_string_literal: true

require "open3"
require "date"

module Deber
  module Docker
    class BuildError < StandardError
    end

    class InspectError < StandardError
    end

    class ListError < StandardError
    end

    class RemoveError < StandardError
    end

    class Image
      attr_reader :name, :tag, :name_tag

      def initialize(name, tag)
        @name = name
        @tag = tag
        @name_tag = "#{name}:#{tag}"
      end

      def built?
        cmd = "docker image ls -q -f reference=#{@name_tag}"
        stdout, stderr, status = Open3.capture3(cmd)
        raise ListError, stderr unless status.success?

        !stdout.empty?
      end

      def build_date
        cmd = "docker image inspect -f '{{ .Created }}' #{@name_tag}"
        stdout, stderr, status = Open3.capture3(cmd)
        raise InspectError, stderr unless status.success?

        DateTime.parse(stdout)
      end

      def build(dockerfile, &block)
        cmd = "docker image build -t #{@name_tag} -"
        status = Open3.popen2e(cmd) do |stdin, stdout_and_stderr, wait_thread|
          stdin.write dockerfile
          stdin.close
          stdout_and_stderr.each(&block)
          wait_thread.value
        end
        raise BuildError unless status.success?

        status
      end

      def remove
        cmd = "docker image rm #{@name_tag}"
        _stdout, stderr, status = Open3.capture3(cmd)
        raise RemoveError, stderr unless status.success?

        status
      end
    end
  end
end
