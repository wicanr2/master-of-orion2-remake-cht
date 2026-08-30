#!/bin/sh
set -eu

ledger=${1:-docs/re/remake-traceability.tsv}
tab=$(printf '\t')
expected="immutable_key${tab}semantic${tab}evidence_doc${tab}spec_doc${tab}spec_state${tab}source_paths${tab}test_paths${tab}implementation_state${tab}verification_state${tab}correction_marker"
exec 3<"$ledger"
IFS= read -r header <&3
if [ "$header" != "$expected" ]; then
    echo "traceability: header mismatch: $ledger" >&2
    exit 1
fi

failed=0
line=1
while IFS="$tab" read -r key semantic evidence spec spec_state sources tests implementation verification correction <&3; do
    line=$((line + 1))
    case "$spec_state" in DRAFT|READY|CONFORMED|SUPERSEDED) ;; *) echo "traceability:$line invalid spec state: $spec_state" >&2; failed=1 ;; esac
    case "$implementation" in MISSING|PARTIAL|IMPLEMENTED|CONFORMED) ;; *) echo "traceability:$line invalid implementation state: $implementation" >&2; failed=1 ;; esac
    case "$verification" in NONE|INSUFFICIENT|INTERNAL|ORACLE) ;; *) echo "traceability:$line invalid verification state: $verification" >&2; failed=1 ;; esac

    address=${key##*:}
    old_ifs=$IFS
    IFS=';'
    for path in $evidence; do
        if [ ! -f "$path" ]; then
            echo "traceability:$line missing evidence: $path" >&2
            failed=1
        elif ! grep -Fq "$address" "$path"; then
            echo "traceability:$line evidence lacks immutable address $address: $path" >&2
            failed=1
        fi
    done
    IFS=$old_ifs

    if [ ! -f "$spec" ]; then
        echo "traceability:$line missing spec: $spec" >&2
        failed=1
    else
        if ! grep -Fq "狀態：$spec_state" "$spec"; then
            echo "traceability:$line spec state not declared: $spec ($spec_state)" >&2
            failed=1
        fi
        if ! grep -Fq "RE-TRACE: $key" "$spec"; then
            echo "traceability:$line spec lacks backlink: $spec ($key)" >&2
            failed=1
        fi
        if [ "$correction" != NONE ] && ! grep -Fq "$correction" "$spec"; then
            echo "traceability:$line spec lacks correction marker: $spec ($correction)" >&2
            failed=1
        fi
    fi

    if [ "$implementation" = MISSING ]; then
        if [ "$sources" != - ] || [ "$tests" != - ] || [ "$verification" != NONE ]; then
            echo "traceability:$line MISSING must use source/test '-' and verification NONE" >&2
            failed=1
        fi
    else
        old_ifs=$IFS
        IFS=';'
        for path in $sources $tests; do
            if [ ! -f "$path" ]; then
                echo "traceability:$line missing source/test: $path" >&2
                failed=1
            fi
        done
        IFS=$old_ifs
    fi
done

# 再以獨立 awk 做欄數與空值硬閘門。
awk -F '\t' 'NR > 1 && (NF != 10 || $1 == "" || $3 == "" || $4 == "") { bad=1 } END { exit bad }' "$ledger" || {
    echo "traceability: malformed or empty required field" >&2
    exit 1
}

exit "$failed"
