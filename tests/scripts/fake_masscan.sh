#!/bin/sh

set -eu

fixture_path=""
if [ "$#" -gt 0 ]; then
	fixture_path="$1"
	shift
fi

if [ -n "${MASSCAN_STDERR:-}" ]; then
	printf '%s\n' "$MASSCAN_STDERR" >&2
fi

output_path="-"
while [ "$#" -gt 0 ]; do
	arg="$1"
	shift

	case "$arg" in
	-oJ|-oX|-oL|-oG|-oB)
		if [ "$#" -gt 0 ]; then
			output_path="$1"
			shift
		fi
		;;
	esac
done

content="${MASSCAN_STDOUT:-}"
if [ -n "$fixture_path" ] && [ -f "$fixture_path" ]; then
	content=$(cat "$fixture_path")
fi

if [ "${MASSCAN_OUTPUT_TO_FILE:-0}" = "1" ] && [ "$output_path" != "-" ]; then
	printf '%s' "$content" > "$output_path"
else
	printf '%s' "$content"
fi

if [ -n "${MASSCAN_EXIT_CODE:-}" ]; then
	exit "$MASSCAN_EXIT_CODE"
fi

exit 0
