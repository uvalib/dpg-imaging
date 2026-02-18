# DPG Imaging

DPG Imaging is a tool that allows review, sequencing and metadata assignment for digitized items.
Requires an IIIF server capable of serving .tif images. Target server is:

https://cantaloupe-project.github.io/

Image storage for the IIIF server and this service is shared.

### System Requirements
* GO version 1.19.0+
* Node 22.13+
* exiftool
* Imagemagick

### Database Notes

The DPG Inaging backend uses a MySQL DB to track everything. The schema is managed by
https://github.com/golang-migrate/migrate and the scripts are in ./backend/db/migrations.

Install the migrate binary on your host system. For OSX, the easiest method is brew. Execute:

`brew install golang-migrate`.

Define your MySQL connection params in an environment variable, like this:

`export IMAGINGDB='mysql://user:password@tcp(host:port)/dbname'`

Run migrations like this:

`migrate -database ${IMAGINGDB} -path backend/db/migrations up`

Example migrate commads to create a migration and run one:

* `migrate create -ext sql -dir backend/db/migrations -seq update_projects`
* `migrate -database ${IMAGINGDB} -path backend/db/migrations/ up`

A migration can be reversed by one step using:

* `migrate -database ${IMAGINGDB} -path backend/db/migrations/ down 1`

During devlopment, if a migration fails you will have to force the schema to the prior version.
For example: you are working on migration number 4. The first attempt to run it fails, so you will
need to force the DB to version 3, fix the up migration and try up again. Example force command:

* `migrate -database ${IMAGINGDB} -path backend/db/migrations/ force 3`

