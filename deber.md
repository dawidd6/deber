deber(1) -- Debian packaging with Docker
=============================================

## SYNOPSIS

`deber` [flags]

## DESCRIPTION

Utility made with simplicity in mind to provide
an easy way for building Debian packages in
Docker containers.

## OPTIONS

 * `-i`, `--include` *string* :
  steps which will be executed as the only ones
  
 * `-e`, `--exclude` *string*:
  steps which should be omitted
   
 * `--version` :
  version for deber
  
 * `-h`, `--help` :
  help for deber

## STEPS

The following steps are executed (in that exact order):

`check`

    Check if to-be-build package is already built and in archive.
    
    Note: if package is in archive, then deber will simply exit.

`build`
    
    Build image. This step is skipped if an image is already built.
    
    Also if image is older than 14 days, then deber will try to rebuild it.
    
`create`

    Create container and make needed directories on host system.
     
`start`
      
    Start container.
    
`tarball`

    Move orig upstream tarball from parent directory to build directory.
    
`update`

    Update apt's cache.
    
`deps`

    Install package's build dependencies in container
     
`package`
      
    Run `dpkg-buildpackage` in container.
     
`test`
      
    Run series of commands in Docker container:
       - debc
       - debi
       - lintian
       
`archive`
     
    Move built package artifacts to archive.
         
    Note: this step is skipped if package directory already exists in archive
         
`scan`
         
    Scan packages in archive.
         
`stop`
      
    Stop container.
     
`remove`

    Remove container.

## EXAMPLES

Basic usage of deber with gbp:
    
    $ gbp buildpackage --git-builder deber
    
Excluding some steps:

    $ deber --exclude remove,stop,archive
    
Removing container after unsuccessful build (if needed):
    
    $ deber --include remove,stop
    
Only building image:

    $ deber --include build
    
Only moving tarball and creating container:

Note: this example assumes that you specified `builder = deber` in `gbp.conf`.

    $ gbp buildpackage --include tarball,create
    
## ENVIRONMENT VARIABLES

**DEBER_ARCHIVE**

    Directory where deber will put built packages.
    
**DEBER_DPKG_BUILDPACKAGE_FLAGS**

    Space separated flags to be passed to dpkg-buildpackage in container.

**DEBER_LINTIAN_FLAGS**

    Space separated flags to be passed to lintian in container.

## SEE ALSO

gbp(1), gbp.conf(5), gbp-buildpackage(1), lintian(1)