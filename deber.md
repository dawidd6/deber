deber(1) -- Debian packaging with Docker
=============================================

## SYNOPSIS

`deber` [flags]

## DESCRIPTION

Utility made with simplicity in mind to provide
an easy way for building Debian packages in
Docker containers.

## OPTIONS

 * `-d`, `--dist` *string* : 
  specify which Distribution to use (default "unstable")
 
 * `--dpkg-buildpackage-flags` *string* :
  specify flags passed to dpkg-buildpackage (default "-tc")
 
 * `-h`, `--help` :
  help for deber
 
 * `--lintian-flags` *string* :
  specify flags passed to lintian (default "-i")
 
 * `-n`, `--network` :
  enable network in container during packaging step
 
 * `-o`, `--os` *string* :
  specify which OS to use (default "debian")
 
 * `-r`, `--repo` *string* :
  specify a local repository to be mounted in container
 
 * `--show-steps` :
  show available steps in order
 
 * `-v`, `--verbose` :
  show more output
 
 * `--version` :
  version for deber
 
 * `-i`, `--with-steps` *string* :
  specify which of the steps should execute
 
 * `-e`, `--without-steps` *string* :
  specify which of the steps should not execute

## EXAMPLES

Using deber with gbp:
    
    $ gbp buildpackage --git-builder deber
    
Using deber (with flags) with gbp:
        
    $ gbp buildpackage --git-builder deber --os debian --dist buster
  
Specifying different OS and Distribution:
  
    $ deber --os ubuntu --dist bionic
  
Execute only selected steps:
    
    $ deber --with-steps build,create,start
  
Execute all steps except a few:
    
    $ deber --without-steps stop,remove
  
Build package with custom dpkg-buildpackage flags:
    
    $ deber --dpkg-buildpackage-flags "-nc -S"
  
Test package with custom lintian flags:
    
    $ deber --lintian-flags "-i -I"
  
Mount local repo and use it in container:
    
    $ deber --repo $HOME/repo/unstable

## SEE ALSO

gbp(1), gbp-buildpackage(1), lintian(1)