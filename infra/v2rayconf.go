package infra

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
)

func parseHost(s string) string {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = "\"" + parts[i] + "\""
	}
	s = strings.Join(parts, ",")
	return s
}

func prettyPrint(b []byte) ([]byte, error) {
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "\t")
	return out.Bytes(), err
}

func DefaultConf() V2genConfig {
	Settings := make(V2genConfig)

	//default settings
	Settings["loglevel"] = "warning"
	Settings["socksPort"] = "1080"
	Settings["udp"] = "true"
	Settings["httpPort"] = "1081"
	Settings["security"] = "aes-256-gcm"
	Settings["mux"] = "true"
	Settings["concurrency"] = "8"
	Settings["dns1"] = "https://1.1.1.1/dns-query"
	Settings["dns2"] = "https://dns.quad9.net/dns-query"
	Settings["china"] = "true"
	Settings["tls"] = "null"
	Settings["kcp"] = "null"
	Settings["ws"] = "null"
	Settings["quic"] = "null"
	Settings["http"] = "null"
	Settings["allowInsecure"] = "false"
	Settings["mtu"] = "1350"
	Settings["tti"] = "20"
	Settings["up"] = "5"
	Settings["down"] = "20"
	Settings["congestion"] = "false"
	Settings["readBufferSize"] = "1"
	Settings["writeBufferSize"] = "1"

	return Settings
}

func GenV2RayConf(conf V2genConfig, template ...[]byte) ([]byte, error) {
	v2rayConf := ConfigTpl
	if len(template) > 0 {
		if len(template) != 1 {
			return nil, errors.New("too many templates")
		}
		v2rayConf = string(template[0])
	}

	if conf["china"] == "true" {
		v2rayConf = strings.ReplaceAll(v2rayConf, "{{china_ip}}", "\n"+`"geoip:cn",`)
		v2rayConf = strings.ReplaceAll(v2rayConf, "{{china_sites}}", ChinaSites)
	} else {
		v2rayConf = strings.ReplaceAll(v2rayConf, "{{china_ip}}", "")
		v2rayConf = strings.ReplaceAll(v2rayConf, "{{china_sites}}", "")
	}

	// set stream
	if conf["tls"] == "tls" {
		v2rayConf = strings.ReplaceAll(v2rayConf, "{{tls}}", TLSObject)
	} else {
		v2rayConf = strings.ReplaceAll(v2rayConf, "{{tls}}", "null")
	}

	switch conf["network"] {
	case "kcp":
		v2rayConf = strings.ReplaceAll(v2rayConf, "{{kcp}}", KcpObject)
	case "ws":
		v2rayConf = strings.ReplaceAll(v2rayConf, "{{ws}}", WsObject)
	case "http":
		v2rayConf = strings.ReplaceAll(v2rayConf, "{{http}}", HttpObject)
		conf["host"] = parseHost(conf["host"])
	case "quic":
		v2rayConf = strings.ReplaceAll(v2rayConf, "{{quic}}", QuicObject)
	}

	// set other settings
	for k, v := range conf {
		v2rayConf = strings.ReplaceAll(v2rayConf, "{{"+k+"}}", v)
	}

	return prettyPrint([]byte(v2rayConf))
}

const ConfigTpl = `{
  "log": {
    "loglevel": "{{loglevel}}"
  },
  "inbounds": [
    {
      "port": {{socksPort}},
      "protocol": "socks",
      "settings": {
		"udp": {{udp}}
      }
    },
    {
      "port": {{httpPort}},
      "protocol": "http",
      "settings": {
		"udp": {{udp}}
      }
    }
  ],
  "outbounds": [ 
	{
    "protocol": "vmess",
    "settings": {
      "vnext": [
        {
          "address": "{{address}}",
          "port": {{serverPort}},
          "users": [
            {
              "id": "{{uuid}}",
              "alterId": {{aid}},
              "security": "{{security}}"
            }
          ]
        }
      ]
    },
    "streamSettings": {
      "network": "{{network}}",
      "security": "{{streamSecurity}}",
      "tlsSettings": {{tls}},
      "kcpSettings": {{kcp}},
      "wsSettings": {{ws}},
      "httpSettings": {{http}},
      "quicSettings": {{quic}},
	  "mux": {
  		"enabled": {{mux}},
      	"concurrency": {{concurrency}}
      }
    }
  	},
    {
      "protocol": "freedom",
      "settings": {},
      "tag": "direct"
    }
],
  "dns": {
    "servers": [
      "{{dns1}}",
      "{{dns2}}",
      "localhost"
    ]
  },
	"routing": {
		"strategy": "rules",
			"settings": {
			"domainStrategy": "IPIfNonMatch",
				"rules": [{{china_sites}}
					{
    			    "type": "field",
    			    "outboundTag": "direct",
     			    "ip": [{{china_ip}}
       				    "geoip:private"
					]
				},
				        {
                    "type": "field",
                    "ip": [
                        "0.0.0.0/8",
                        "10.0.0.0/8",
                        "100.64.0.0/10",
                        "127.0.0.0/8",
                        "169.254.0.0/16",
                        "172.16.0.0/12",
                        "192.0.0.0/24",
                        "192.0.2.0/24",
                        "192.168.0.0/16",
                        "198.18.0.0/15",
                        "198.51.100.0/24",
                        "203.0.113.0/24",
                        "::1/128",
                        "fc00::/7",
                        "fe80::/10"
                    ],
                    "outboundTag": "direct"
                }
			]
		}
	}
}`

const ChinaSites = `
{
	"type": "field",
	"outboundTag": "direct",
    "domain": ["geosite:cn"] 
},`

const (
	TLSObject = `{
 		 "serverName": "{{address}}",
 		 "allowInsecure": {{allowInsecure}},
 		 "alpn": ["http/1.1"]
		}`

	WsObject = `{
 		 "path": "{{path}}",
 		 "headers": {
  		  "Host": "{{host}}"
 		 }
		}`

	KcpObject = `
{
		"mtu": {{mtu}},
		"tti": {{tti}},
		"uplinkCapacity": {{up}},
		"downlinkCapacity": {{down}},
		"congestion": {{congestion}},
		"readBufferSize": {{readBufferSize}},
		"writeBufferSize": {{writeBufferSize}},
		"header": {
		"type": "{{type}}"
		}
		}`

	HttpObject = `{
		"host": [{{host}}],
		"path": "{{path}}"
		}`
	QuicObject = `{
		  "security": "{{host}}",
		  "key": "{{path}}",
		  "header": {
		    "type": "{{type}}"
		  }
		}`
)
