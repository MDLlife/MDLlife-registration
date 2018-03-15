package utils

func CombineDatetime(y string, m string, d string) string {
	str := y + "-"

	if len(m) == 1 {
		str += "0"
	}

	str += m + "-"

	if len(d) == 1 {
		str += "0"
	}

	str += d

	return str
}
