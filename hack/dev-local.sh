#!/usr/bin/env bash

set -e

export PGUSER=postgres
export PGPASSWORD=password
export PGDATABASE=tofutf
# here we grab all of the configuration from the running tofutf instance.
export OTF_LOG_HTTP_REQUESTS=true
export OTF_SANDBOX=false
export OTF_V=10
export OTF_SECRET='105cee69ed55b1ed85b1025bea1f2589'
export OTF_DATABASE="postgresql://$PGUSER:$PGPASSWORD@localhost:$port_forward_local_port/$PGDATABASE"
export OTF_PROVIDER_PROXY_URL=https://registry.opentofu.org/v1/providers/
export OTF_PROVIDER_PROXY_IS_ARTIFACTORY=false

air
