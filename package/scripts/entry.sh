#!/usr/bin/env bash
#
# run application
#

# run the server
umask 0002
cd bin; ./imagingsvc -url $DPG_SERVICE_URL \
   -images $DPG_IMAGE_PATH \
   -scan $DPG_SCAN_PATH \
   -finalize $DPG_FINALIZE_PATH \
   -iiif $IIIF_SERVICE_URL \
   -finalizeurl $DPG_FINALIZE_URL \
   -tsurl $DPG_TRACKSYS_URL \
   -jwtkey $DPG_JWT_KEY \
   -dbhost $DBHOST \
   -dbport $DBPORT \
   -dbname $DBNAME \
   -dbuser $DBUSER \
   -dbpass $DBPASS

# return the status
exit $?

#
# end of file
#
