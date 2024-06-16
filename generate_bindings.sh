#!/usr/bin/env bash
set -eu

# Generate Go bindings for contract

#
# CONTRACTS_PATH="./src"

if [ "$#" -ne 2 ]; then
	echo "This script takes 2 arguments - CONTRACT_NAME PACKAGE"
	exit 1
fi

need_cmd() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "need '$1' (command not found)"
		exit 1
	fi
}

need_cmd forge
need_cmd abigen

NAME=$1

# Fully qualified: path-to-contract-file:contract-name
TYPE=$(echo "$NAME" | cut -d ':' -f2)
PACKAGE=$2

# Convert to lower case to respect golang package naming conventions
TYPE_LOWER=$(echo ${TYPE} | tr '[:upper:]' '[:lower:]')
# FILENAME="${TYPE_LOWER}_deployed.go"

mkdir -p dist/
mkdir -p sdk/go/
OUTPUT_DIR=sdk/go/

CWD=$(pwd)
# Build contracts
# cd ${CONTRACTS_PATH}
forge inspect "${NAME}" abi >${OUTPUT_DIR}/${TYPE}.abi
forge inspect "${NAME}" bytecode >${OUTPUT_DIR}/${TYPE}.bin
forge inspect "${NAME}" deployedBytecode >${CWD}/dist/${TYPE_LOWER}_deployed.hex

# Run ABIGEN
#cd ${CWD}
./abigen \
	--abi ${OUTPUT_DIR}/${TYPE}.abi \
	--bin ${OUTPUT_DIR}/${TYPE}.bin \
	--pkg ${PACKAGE} \
	--type ${TYPE} \
	--out ./${PACKAGE}/${TYPE_LOWER}.go
