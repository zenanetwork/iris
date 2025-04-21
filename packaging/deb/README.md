# Debian

For debian packages you will need to add the following layouts during the build

irisd/
DEBIAN/control
DEBIAN/postinst
usr/local/bin/irisd
lib/systemd/system/irisd.service

This will be wrapped during the build package process building

Note this is still a work in progress:

TODO: removal/purge on removal using dpkg
cleanup of control files to list what we want
copyright inclusion

CLI:

iriscli/
DEBIAN/control
DEBIAN/postinst
usr/local/bin/iriscli
