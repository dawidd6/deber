# frozen_string_literal: true

require "open3"
require "date"

module Deber
  module Docker
    class ExecError < StandardError
    end

    class InspectError < StandardError
    end

    class ListError < StandardError
    end

    class RemoveError < StandardError
    end

    class Container
      attr_reader :name, :image

      def initialize(name, image)
        @name = name
        @image = image
      end

      def created?
        cmd = "docker container ls -a -q -f name=#{@name}"
        stdout, stderr, status = Open3.capture3(cmd)
        raise ListError, stderr unless status.success?

        !stdout.empty?
      end

      def creation_date
        cmd = "docker container inspect -f '{{ .Created }}' #{@name}"
        stdout, stderr, status = Open3.capture3(cmd)
        raise InspectError, stderr unless status.success?

        DateTime.parse(stdout)
      end

      def exec(*cmd, &block)
        cmd = "docker container exec #{@name} -- #{cmd.join(" ")}"
        status = Open3.popen2e(cmd) do |_stdin, stdout_and_stderr, wait_thread|
          stdout_and_stderr.each(&block)
          wait_thread.value
        end
        raise ExecError unless status.success?

        status
      end

      def create
        cmd = "docker container create --name=#{@name} #{@image.name_tag}"
        _stdout, stderr, status = Open3.capture3(cmd)
        raise RemoveError, stderr unless status.success?

        status
      end

      def remove
        cmd = "docker container rm #{@name}"
        _stdout, stderr, status = Open3.capture3(cmd)
        raise RemoveError, stderr unless status.success?

        status
      end
    end
  end
end
