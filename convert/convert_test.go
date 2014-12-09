package convert

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCIDR(t *testing.T) {
	checkHeader(
		t,
		cidrHeader,
		[]string{"network"},
	)

	v4net := "1.1.1.0/24"
	checkLine(
		t,
		cidrLine,
		v4net,
		[]string{v4net},
	)

	v6net := "2001:db8:85a3:42::/64"
	checkLine(
		t,
		cidrLine,
		v6net,
		[]string{v6net},
	)
}

func TestRange(t *testing.T) {
	checkHeader(
		t,
		rangeHeader,
		[]string{"network_start_ip", "network_last_ip"},
	)

	checkLine(
		t,
		rangeLine,
		"1.1.1.0/24",
		[]string{"1.1.1.0", "1.1.1.255"},
	)

	checkLine(
		t,
		rangeLine,
		"2001:0db8:85a3:0042::/64",
		[]string{"2001:db8:85a3:42::", "2001:db8:85a3:42:ffff:ffff:ffff:ffff"},
	)
}

func TestIntRange(t *testing.T) {
	checkHeader(
		t,
		intRangeHeader,
		[]string{"network_start_integer", "network_last_integer"},
	)

	checkLine(
		t,
		intRangeLine,
		"1.1.1.0/24",
		[]string{"16843008", "16843263"},
	)

	checkLine(
		t,
		intRangeLine,
		"2001:0db8:85a3:0042::/64",
		[]string{"42540766452641155289225172512357220352",
			"42540766452641155307671916586066771967"},
	)
}

func checkHeader(
	t *testing.T,
	makeHeader headerFunc,
	expected []string) {

	suffix := []string{"city", "country"}
	assert.Equal(
		t,
		append(expected, suffix...),
		makeHeader(suffix),
	)
}

func checkLine(
	t *testing.T,
	makeLine lineFunc,
	network string,
	expected []string) {
	p, err := makePrefix(network)
	if err != nil {
		t.Fatal(err)
	}

	suffix := []string{"1", "2"}
	assert.Equal(
		t,
		append(expected, suffix...),
		makeLine(p, suffix),
	)
}

func TestCIDROutput(t *testing.T) {
	checkOutput(
		t,
		"CIDR only",
		true,
		false,
		false,
		[]interface{}{
			"network",
			"1.0.0.0/24",
			"4.69.140.16/29",
			"5.61.192.0/21",
			"2001:4220::/32",
			"2402:d000::/32",
			"2406:4000::/32",
		},
	)
}

func TestRangeOutput(t *testing.T) {
	checkOutput(
		t,
		"range only",
		false,
		true,
		false,
		[]interface{}{
			"network_start_ip,network_last_ip",
			"1.0.0.0,1.0.0.255",
			"4.69.140.16,4.69.140.23",
			"5.61.192.0,5.61.199.255",
			"2001:4220::,2001:4220:ffff:ffff:ffff:ffff:ffff:ffff",
			"2402:d000::,2402:d000:ffff:ffff:ffff:ffff:ffff:ffff",
			"2406:4000::,2406:4000:ffff:ffff:ffff:ffff:ffff:ffff",
		},
	)
}

func TestIntRangeOutput(t *testing.T) {
	checkOutput(
		t,
		"integer range only",
		false,
		false,
		true,
		[]interface{}{
			"network_start_integer,network_last_integer",
			"16777216,16777471",
			"71666704,71666711",
			"87932928,87934975",
			"42541829336310884227257139937291534336,42541829415539046741521477530835484671",
			"47866811183171600627242296191018336256,47866811262399763141506633784562286591",
			"47884659703622814097215369772150030336,47884659782850976611479707365693980671",
		},
	)
}

func TestAllOutput(t *testing.T) {
	checkOutput(
		t,
		"all output options",
		true,
		true,
		true,
		[]interface{}{
			"network,network_start_ip,network_last_ip,network_start_integer,network_last_integer",
			"1.0.0.0/24,1.0.0.0,1.0.0.255,16777216,16777471",
			"4.69.140.16/29,4.69.140.16,4.69.140.23,71666704,71666711",
			"5.61.192.0/21,5.61.192.0,5.61.199.255,87932928,87934975",
			"2001:4220::/32,2001:4220::,2001:4220:ffff:ffff:ffff:ffff:ffff:ffff,42541829336310884227257139937291534336,42541829415539046741521477530835484671",
			"2402:d000::/32,2402:d000::,2402:d000:ffff:ffff:ffff:ffff:ffff:ffff,47866811183171600627242296191018336256,47866811262399763141506633784562286591",
			"2406:4000::/32,2406:4000::,2406:4000:ffff:ffff:ffff:ffff:ffff:ffff,47884659703622814097215369772150030336,47884659782850976611479707365693980671",
		},
	)
}

func checkOutput(
	t *testing.T,
	name string,
	cidr bool,
	ipRange bool,
	intRange bool,
	expected []interface{},
) {
	input := `network,geoname_id,registered_country_geoname_id,represented_country_geoname_id,is_anonymous_proxy,is_satellite_provider
1.0.0.0/24,2077456,2077456,,0,0
4.69.140.16/29,6252001,6252001,,0,0
5.61.192.0/21,2635167,2635167,,0,0
2001:4220::/32,357994,357994,,0,0
2402:d000::/32,1227603,1227603,,0,0
2406:4000::/32,1835841,1835841,,0,0
`
	var outbuf bytes.Buffer

	err := Convert(strings.NewReader(input), &outbuf, cidr, ipRange, intRange)
	if err != nil {
		t.Fatal(err)
	}

	// This is a regexp as Go 1.4 does not quote empty fields while earlier
	// versions do
	outTMPL := `%s,geoname_id,registered_country_geoname_id,represented_country_geoname_id,is_anonymous_proxy,is_satellite_provider
%s,2077456,2077456,(?:"")?,0,0
%s,6252001,6252001,(?:"")?,0,0
%s,2635167,2635167,(?:"")?,0,0
%s,357994,357994,(?:"")?,0,0
%s,1227603,1227603,(?:"")?,0,0
%s,1835841,1835841,(?:"")?,0,0
`

	assert.Regexp(
		t,
		fmt.Sprintf(outTMPL, expected...),
		outbuf.String(),
	)
}

func TestFileWriting(t *testing.T) {
	input := `network,something
1.0.0.0/24,"some more"
`

	expected := `network,network_start_ip,network_last_ip,network_start_integer,network_last_integer,something
1.0.0.0/24,1.0.0.0,1.0.0.255,16777216,16777471,some more
`

	inFile, err := ioutil.TempFile("", "input")
	if err != nil {
		t.Fatal(err)
	}
	defer inFile.Close()

	outFile, err := ioutil.TempFile("", "output")
	if err != nil {
		t.Fatal(err)
	}
	defer outFile.Close()

	inFile.WriteString(input)

	err = ConvertFile(inFile.Name(), outFile.Name(), true, true, true)
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer(nil)
	io.Copy(buf, outFile)

	assert.Equal(t, expected, buf.String())
}
