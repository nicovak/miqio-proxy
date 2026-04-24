#!/bin/bash
set -e

echo "Bump VERSION.md"
echo "${1}" > VERSION.md

echo "[DOCKER] Login"
echo "$GITHUB_TOKEN" | docker login ghcr.io -u "$GITHUB_ACTOR" --password-stdin

touch .env
echo "GITHUB_PAT=$NICOVAK_PAT" >> .env
echo "GOPRIVATE=github.com/nicovak/miqio-go" >> .env

echo "[DOCKER] Building images"
docker buildx build --secret id=github,src=.env --load --cache-from="type=gha" --cache-to="type=gha,mode=max" -t ghcr.io/nicovak/miqio-proxy:latest -t ghcr.io/nicovak/miqio-proxy:"${1}" .

echo "[DOCKER] Pushing images"
docker push ghcr.io/nicovak/miqio-proxy:latest
docker push ghcr.io/nicovak/miqio-proxy:"${1}"

if [ "$GITHUB_REF" == 'refs/heads/master' ]; then
echo "[KUSTOMIZE] Applying PRD"
pushd k8s/kustomize/overlays/prd
kustomize edit set image ghcr.io/nicovak/miqio-proxy:"${1}"
sed -ri "s/^(\s*app.kubernetes.io\/version\s*:\s*).*/\1${1}/" labels-transformer.yaml
popd

echo "[KUSTOMIZE] Applying QA"
pushd k8s/kustomize/overlays/qa
kustomize edit set image ghcr.io/nicovak/miqio-proxy:"${1}"
sed -ri "s/^(\s*app.kubernetes.io\/version\s*:\s*).*/\1${1}/" labels-transformer.yaml
popd
fi

if [ "$GITHUB_REF" == 'refs/heads/develop' ]; then
echo "[KUSTOMIZE] Applying INT"
pushd k8s/kustomize/overlays/int
kustomize edit set image ghcr.io/nicovak/miqio-proxy:"${1}"
sed -ri "s/^(\s*app.kubernetes.io\/version\s*:\s*).*/\1${1}/" labels-transformer.yaml
popd
fi
