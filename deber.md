deber(1) -- Debian packaging with Docker
=============================================

## SYNOPSIS

`deber` [flags]

## DESCRIPTION

Utility made with simplicity in mind to provide
an easy way for building Debian packages in
Docker containers.

## OPTIONS

 * `-h`, `--help` :
  help for deber
   
 * `--dpkg-buildpackage-flags` *string* :
  specify flags passed to dpkg-buildpackage (default "-tc")
 
 * `--lintian-flags` *string* :
  specify flags passed to lintian (default "-i")
 
 * `-u`, `--update-after` *string* :
  perform apt cache update after specified interval (default 30m0s)
 
 * `-f`, `--from` *string* :
  specify which Docker image to use (default "debian:unstable")
 
 * `-r`, `--repo` *string* :
  specify a local repository to be mounted in container
  
 * `-c`, `--clean` :
  stop and remove uncleaned container
   
 * `--version` :
  version for deber

## STEPS

The following steps are executed (in that exact order):

`build`
    
    Build Docker image. This step is skipped if an image is already built.
    
`create`

      Create Docker container.
     
`start`
      
      Start Docker container.
     
`package`
      
      Run series of commands in Docker container:
       - apt-get update
       - mk-build-deps
       - dpkg-buildpackage
     
`test`
      
      Run series of commands in Docker container:
       - debc
       - debi
       - lintian
     
`stop`
      
      Stop Docker container.
     
`remove`

      Remove Docker container.

## EXAMPLES

Using deber with gbp:
    
    $ gbp buildpackage --git-builder deber
    
Using deber (with flags) with gbp:
        
    $ gbp buildpackage --git-builder deber --from debian:buster
  
Specifying different OS and Distribution:
  
    $ deber --from ubuntu:bionic
  
Build package with custom dpkg-buildpackage flags:
    
    $ deber --dpkg-buildpackage-flags "-nc -S"
  
Test package with custom lintian flags:
    
    $ deber --lintian-flags "-i -I"
  
Mount local repo and use it in container:
    
    $ deber --repo $HOME/repo/unstable
    
Update apt cache now:

    $ deber --update-after 0

## SEE ALSO

gbp(1), gbp-buildpackage(1), lintian(1)