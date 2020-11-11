#!/bin/bash
# search an input file for a line containing "option go_package" (or an alternate input argument $2)
# output the small-form go_package name, i.e. "pb" instead of "github.com/Infoblox-CTO.../pb;pb"
INPUT_FILE=$1
NAMESPACE_LINE=$2
if [ -z "$NAMESPACE_LINE"]
then
  NAMESPACE_LINE="option go_package"
fi
GO_PACKAGE_CONTENTS=''
while IFS= read -r line
do
    if [[ "$line" == *"$NAMESPACE_LINE"* ]]; then
        GO_PACKAGE_CONTENTS=`echo "$line" | awk '{split($0,a,"\""); print a[2]}'`
    fi
done < "$INPUT_FILE"

IFS='; ' read -r -a GO_PACKAGE_ARRAY <<< "$GO_PACKAGE_CONTENTS"
for element in "${GO_PACKAGE_ARRAY[@]}"
do
    if [[ "$element" != *"/"* ]]; then
        echo "$element"
    fi
done