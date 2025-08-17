{
    # Remove trailing carriage return
    sub(/\r$/, "")

    # Remove trailing whitespace
    sub(/[[:space:]]*$/, "")

	# Spelling error
	sub(/_Unkonw /, "_Unknown ")

    # Replace package line & capture pkg_name
    if ($0 ~ /^package .*;$/) {

		# eg: package: Qot_GetBasicQot
		# pkg_name: QotGetBasicQot
		pkg_name = $0
		sub(/^package /, "", pkg_name)
		sub(/;$/, "", pkg_name)
		sub(/_/, "", pkg_name)

        $0 = "package futupb;"
    }

    # Replace go_package option line
    if ($0 ~ /^option go_package .*;$/) {
        $0 = "option go_package = \"github.com/santsai/futu-go/pb\";"

		# add missing import "Common.proto"
		if (pkg_name == "TestCmd" ||
			pkg_name == "TrdFlowSummary" ||
			pkg_name == "TrdGetMarginRatio" ||
			pkg_name == "UsedQuota" ||
			0) {
			$0 = $0 "\n\nimport \"Common.proto\";" 
		}

		# missing "Qot_Common.proto"
		if (pkg_name == "GetUserInfo" ||
			pkg_name == "Notify" ||
			0) {
			$0 = $0 "\n\nimport \"Qot_Common.proto\";" 
		}

		# missing "Trd_Common.proto"
		if (pkg_name == "TrdUnlockTrade" ||
			0) {
			$0 = $0 "\n\nimport \"Trd_Common.proto\";" 
		}
    }

	# rename C2S -> xxxRequest
	if ($0 ~/^message C2S$/) {
		request_name = pkg_name "Request"
		$0 = "message " request_name
	}

	# rename S2C -> xxxResponse
	if ($0 ~/^message S2C$/) {
		in_response = 0
		in_response_payload = 1

		response_name = pkg_name "Response"
		$0 = "message " response_name
	}

	if ($0 ~ /^message Request$/) {
		$0 = "message " pkg_name "Request_Internal"
	}

	if ($0 ~ /^message Response$/) {
		in_response = 1
		in_response_payload = 0

		$0 = "message " pkg_name "Response_Internal"
	}

	if ($0 ~ /required C2S c2s = 1;$/) {
		sub(/C2S c2s/, request_name " payload", $0)
	}

	if ($0 ~ /optional S2C s2c = 4;$/) {
		sub(/S2C s2c/, response_name " payload", $0)
	}

	# fix enum types
	if ($0 ~ / int32 /) {
		for (rk in enum_replaces) {
			if (index($0, rk) > 0) {
				rv = enum_replaces[rk]
				if (rv != "") {
					sub("int32", rv, $0)
				}
				break
			}
		}

		# fix enum retType -400
		if ($0 ~ /default = -400/) {
			sub("-400", "RetType_Unknown")
		}
	}

	# handle duplicate message in Notify.proto
	# ProgramStatus, QotRight
	if (pkg_name == "Notify") {
		if (!(in_response || in_response_payload) &&
			match($0, /^message (.*)$/)) {
			$0 = $0 "Notice"
		}

		if (in_response_payload &&
			match($0, /optional ([^ ]+)/)) {
			msgname = substr($0, RSTART, RLENGTH)
			sub(/optional ([^ ]+)/, msgname "Notice", $0)
		}
	}

	# remove package names
	# eg: Qot_Common.Security -> Security
	sub(/ Common\./, " ", $0)
	sub(/ Qot_Common\./, " ", $0)
	sub(/ Trd_Common\./, " ", $0)

    print
}
