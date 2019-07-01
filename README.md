
```

# make sure you enabled CORS addon on Minishift
# https://github.com/aerogear/mobile-developer-console#enable-cors-in-the-openshift-cluster

if minishift addons list | grep cors ; then
    minishift addons apply cors
else
    MINISHIFT_ADDONS_PATH=/tmp/minishift-addons
    rm -rf $MINISHIFT_ADDONS_PATH && git clone https://github.com/minishift/minishift-addons.git $MINISHIFT_ADDONS_PATH
    # Not needed after https://github.com/minishift/minishift-addons/pull/187 is merged
    cd $MINISHIFT_ADDONS_PATH
    git fetch origin pull/187/head:cors-fix && git checkout cors-fix
    minishift addons install /tmp/minishift-addons/add-ons/cors
    minishift addons apply cors
fi
minishift addon apply cors

oc login -u system:admin
make cluster/clean
make cluster/prepare

OPENSHIFT_HOST=$(minishift ip):8443 make code/run


kubectl apply  -n mobile-developer-console -f deploy/crds/mdc_v1alpha1_mobiledeveloperconsole_cr.yaml

AFTER WORK - things that aren't done by the operator yet.

MDC_ROUTE=$(oc -n mobile-developer-console get route example-mdc-mdc-proxy --template "{{.spec.host}}")

cat <<EOF | oc apply -f -
apiVersion: v1
grantMethod: auto
kind: OAuthClient
metadata:
  name: mobile-developer-console
secret: SECRETPLACEHOLDER
redirectURIs: ["https://${MDC_ROUTE}"]
EOF


# needed to make the MDC server side SA to list/watch/etc some resources in all namespaces
oc create clusterrole mobileclient-admin --verb=create,delete,get,list,patch,update,watch --resource=mobileclients,secrets,configmaps
oc adm policy add-cluster-role-to-user mobileclient-admin system:serviceaccount:mobile-developer-console:example-mdc

# not sure if we need the same things for keycloakrealms, mobilesecurityserviceapps


oc rollout latest example-mdc

```