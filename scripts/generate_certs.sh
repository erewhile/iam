#!/bin/bash
set -e

BASE_DIR=$(cd "$(dirname "$0")/.." && pwd)
CERT_DIR="$BASE_DIR/certs"

mkdir -p "$CERT_DIR"

if [ ! -f "$CERT_DIR/iam_private.pem" ]; then
    openssl genpkey -algorithm RSA -out "$CERT_DIR/iam_private.pem" -pkeyopt rsa_keygen_bits:2048
    chmod 600 "$CERT_DIR/iam_private.pem"
fi

if [ ! -f "$CERT_DIR/iam_public.pem" ]; then
    openssl rsa -in "$CERT_DIR/iam_private.pem" -pubout -out "$CERT_DIR/iam_public.pem"
fi

echo "certs check/generation done in $CERT_DIR"
ls -l "$CERT_DIR"