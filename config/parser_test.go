package config

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParse(t *testing.T) {
	Convey("TestParseConfig", t, func() {
		data := `{
			"inbounds": [
				{
					"listen": "127.0.0.1:8053"
				}
			],
			"groups": [
				{
					"outbounds": [
						{
							"setting": {
								"addr": "223.5.5.5"
							}
						}
					]
				}
			]
		}`
		want := `{
			"inbounds": [
				{
					"protocol": "dns",
					"listen": "127.0.0.1:8053",
					"net": "udp",
					"tls_cert": "",
					"tls_key": ""
				}
			],
			"groups": [
				{
					"tag": "default",
					"outbounds": [
						{
							"protocol": "dns",
							"setting": {
								"addr": "223.5.5.5"
							}
						}
					],
					"cache": {}
				}
			],
			"routing": null,
			"log": {
				"Level": "",
				"Filename": ""
			}
		}`
		cfg, err := Parse([]byte(data))
		So(err, ShouldBeNil)
		b, _ := json.Marshal(cfg)
		So(string(b), ShouldEqualJSON, want)
	})
}
