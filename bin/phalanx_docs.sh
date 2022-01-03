#!/bin/bash

# Show help
usage() {
  cat << EOS >&2
phalanx_docs.sh
  Generate documents for Phalanx indexing from flat JSONL(NDJSON).

USAGE:
  phalanx_docs.sh [options] <input_file>

OPTIONS:
  -i, --id-field        Specify the unique ID field name in the input data.
  -h, --help            Show this help message.

ARGS:
  <input_file>          Input file path.
EOS
}

# Show error message
invalid() {
  usage 1>&2
  echo "$@" 1>&2
  exit 1
}

# Parse command line options
ARGS=()
while (( $# > 0 ))
do
  case $1 in
    -h | --help)
      usage
      exit 0
      ;;
    -i | --id-field | --id-field=*)
      if [[ -n "${ID_FIELD}" ]]; then
        invalid "Duplicated 'id-field'."
        exit 1
      elif [[ "$1" =~ ^--id-field= ]]; then
        ID_FIELD=$(echo $1 | sed -e 's/^--id-field=//')
      elif [[ -z "$2" ]] || [[ "$2" =~ ^-+ ]]; then
        invalid "'id-field' requires an argument."
        exit 1
      else
        ID_FIELD="$2"
        shift
      fi
      ;;
    -*)
      invalid "Illegal option -- '$(echo $1 | sed 's/^-*//')'."
      exit 1
      ;;
    *)
      ARGS=("${ARGS[@]}" "$1")
      ;;
  esac
  shift
done

# Set input file name
FILENAME=${ARGS[0]}

if  [ -p /dev/stdin ]; then
    cat -
else
    cat ${FILENAME}
fi | jq -c -r '.'${ID_FIELD}' as $id | . |= .+ {"_id": $id | tostring}'
