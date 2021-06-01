#!/usr/bin/env bash
#
# run application
#

# run the server
cd bin; ./imagingsvc -url $DPG_SERVICE_URL -images $DPG_IMAGE_PATH -iiif $IIIF_SERVICE_URL -dbhost $DBHOST -dbport $DBPORT -dbname $DBNAME -dbuser $DBUSER -dbpass $DBPASS

# return the status
exit $?

#
# end of file
#
