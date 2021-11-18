#!/bin/bash

ROOT=$(cd $(dirname $0)/../../; pwd)

set -o errexit
set -o nounset
set -o pipefail

while [[ $# -gt 0 ]]; do
    case ${1} in
        --tmpdir)
            tmpdir="$2"
            shift
            ;;
        *)
            usage
            ;;
    esac
    shift
done
cp -f validatorconfig.yaml kubernetes 
export CA_BUNDLE=$(openssl base64 -A < ${tmpdir}/caCert.pem)

echo ${CA_BUNDLE}

sed -i "s/\${CA_BUNDLE}/${CA_BUNDLE}/g" `grep -rl "{CA_BUNDLE}" ./kubernetes`
