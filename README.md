
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


kubectl apply  -n mdc -f deploy/crds/mdc_v1alpha1_mobiledeveloperconsole_cr.yaml
```