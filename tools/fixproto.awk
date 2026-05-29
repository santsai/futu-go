# vim: noet ts=4
{
	# Remove trailing carriage return
	# sub(/\r$/, "")

	# Remove trailing whitespace
	# sub(/[[:space:]]*$/, "")

	# Spelling error
	sub(/_Unkonw /, "_Unknown ")
	sub(/_Unknow /, "_Unknown ")

	# Replace package line & capture pkg_name
	if ($0 ~ /^package .*;$/) {

		# eg: package: Qot_GetBasicQot
		# pkg_name: QotGetBasicQot
		pkg_name = $0
		sub(/^package /, "", pkg_name)
		sub(/;$/, "", pkg_name)
		sub(/_/, "", pkg_name)

		$0 = "package futupb;\n"
		$0 = $0 "option go_package = \"github.com/santsai/futu-go/pb\";"

		if (pkg_name == "Common") {
			pkg_name = "Types"
		}

		# fix imports
		if (pkg_name != "Types") {
			$0 = $0 "\nimport \"Types.proto\";"
		}
	}

	# Remove existing imports
	if ($0 ~ /^import \".*Common.proto\";/) {
		$0 = ""
	}

	# Replace go_package option line
	if ($0 ~ /^option .*_package .*;$/) {
		$0 = ""
	}

	# duplicate names fix
	if (pkg_name ~ /^Qot.*Screen$/) {

		PFX = pkg_name
		sub("Qot", "", PFX)

		split("Boundary Interval Sort", kws, " ")
		nkws = length(kws)
		for (i = 1; i <= nkws; i++) {
			if ($0 ~ (" " kws[i] " ")) {
				sub(kws[i], PFX kws[i], $0)
			}
		}
	}

	# duplicate name fix
	if (pkg_name ~ /^Notify$/) {
		split("GtwEvent ProgramStatus ConnectStatus QotRight APILevel APIQuota UsedQuota", kws, " ")
		nkws = length(kws)
		for (i = 1; i <= nkws; i++) {
			if ($0 ~ (" " kws[i] " ")) {
				sub(kws[i], "Notify" kws[i], $0)
			}
		}
	}

	# rename C2S -> xxxRequest
	if ($0 ~/^message C2S {/) {
		request_name = pkg_name "Request"
		sub(/C2S/, request_name, $0)
	}

	# rename S2C -> xxxResponse
	if ($0 ~/^message S2C {/) {
		response_name = pkg_name "Response"
		sub(/S2C/, response_name, $0)
	}

	if ($0 ~ /^message Request {$/) {
		$0 = "message " pkg_name "Request_Internal {"
	}

	if ($0 ~ /^message Response {$/) {
		$0 = "message " pkg_name "Response_Internal {"
	}

	if ($0 ~ /required C2S c2s = 1;/) {
		sub(/C2S c2s/, request_name " payload", $0)
	}

	if ($0 ~ /optional S2C s2c = 4;/) {
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

	# remove package names
	# eg: Qot_Common.Security -> Security
	sub(/ Common\./, " ", $0)
	sub(/ Qot_Common\./, " ", $0)
	sub(/ Trd_Common\./, " ", $0)

	print
}
