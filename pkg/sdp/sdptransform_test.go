package sdp

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	txt := `v=0
o=- 4852366920923876131 1737621625 IN IP4 0.0.0.0
s=-
t=0 0
a=msid-semantic:WMS*
a=fingerprint:sha-256 47:20:F0:7A:A1:9A:97:6C:F7:8A:1E:EC:C4:3F:72:4B:24:3F:9E:66:3F:09:0C:FB:FF:F5:F5:08:3B:78:C3:90
a=extmap-allow-mixed
a=group:BUNDLE 0 1
m=video 9 UDP/TLS/RTP/SAVPF 96 97 102 103 104 105 106 107 108 109 127 125 39 40 45 46 98 99 100 101 112 113
c=IN IP4 0.0.0.0
a=setup:actpass
a=mid:0
a=ice-ufrag:oWRmZjUxbxRgiLKd
a=ice-pwd:ATxASlnMCZCjbBUBmkvmWkwYOUnQXIky
a=rtcp-mux
a=rtcp-rsize
a=rtpmap:96 VP8/90000
a=rtcp-fb:96 goog-remb 
a=rtcp-fb:96 ccm fir
a=rtcp-fb:96 nack 
a=rtcp-fb:96 nack pli
a=rtcp-fb:96 nack 
a=rtcp-fb:96 nack pli
a=rtcp-fb:96 transport-cc 
a=rtpmap:97 rtx/90000
a=fmtp:97 apt=96
a=rtcp-fb:97 nack 
a=rtcp-fb:97 nack pli
a=rtcp-fb:97 transport-cc 
a=rtpmap:102 H264/90000
a=fmtp:102 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42001f
a=rtcp-fb:102 goog-remb 
a=rtcp-fb:102 ccm fir
a=rtcp-fb:102 nack 
a=rtcp-fb:102 nack pli
a=rtcp-fb:102 nack 
a=rtcp-fb:102 nack pli
a=rtcp-fb:102 transport-cc 
a=rtpmap:103 rtx/90000
a=fmtp:103 apt=102
a=rtcp-fb:103 nack 
a=rtcp-fb:103 nack pli
a=rtcp-fb:103 transport-cc 
a=rtpmap:104 H264/90000
a=fmtp:104 level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=42001f
a=rtcp-fb:104 goog-remb 
a=rtcp-fb:104 ccm fir
a=rtcp-fb:104 nack 
a=rtcp-fb:104 nack pli
a=rtcp-fb:104 nack 
a=rtcp-fb:104 nack pli
a=rtcp-fb:104 transport-cc 
a=rtpmap:105 rtx/90000
a=fmtp:105 apt=104
a=rtcp-fb:105 nack 
a=rtcp-fb:105 nack pli
a=rtcp-fb:105 transport-cc 
a=rtpmap:106 H264/90000
a=fmtp:106 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f
a=rtcp-fb:106 goog-remb 
a=rtcp-fb:106 ccm fir
a=rtcp-fb:106 nack 
a=rtcp-fb:106 nack pli
a=rtcp-fb:106 nack 
a=rtcp-fb:106 nack pli
a=rtcp-fb:106 transport-cc 
a=rtpmap:107 rtx/90000
a=fmtp:107 apt=106
a=rtcp-fb:107 nack 
a=rtcp-fb:107 nack pli
a=rtcp-fb:107 transport-cc 
a=rtpmap:108 H264/90000
a=fmtp:108 level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=42e01f
a=rtcp-fb:108 goog-remb 
a=rtcp-fb:108 ccm fir
a=rtcp-fb:108 nack 
a=rtcp-fb:108 nack pli
a=rtcp-fb:108 nack 
a=rtcp-fb:108 nack pli
a=rtcp-fb:108 transport-cc 
a=rtpmap:109 rtx/90000
a=fmtp:109 apt=108
a=rtcp-fb:109 nack 
a=rtcp-fb:109 nack pli
a=rtcp-fb:109 transport-cc 
a=rtpmap:127 H264/90000
a=fmtp:127 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=4d001f
a=rtcp-fb:127 goog-remb 
a=rtcp-fb:127 ccm fir
a=rtcp-fb:127 nack 
a=rtcp-fb:127 nack pli
a=rtcp-fb:127 nack 
a=rtcp-fb:127 nack pli
a=rtcp-fb:127 transport-cc 
a=rtpmap:125 rtx/90000
a=fmtp:125 apt=127
a=rtcp-fb:125 nack 
a=rtcp-fb:125 nack pli
a=rtcp-fb:125 transport-cc 
a=rtpmap:39 H264/90000
a=fmtp:39 level-asymmetry-allowed=1;packetization-mode=0;profile-level-id=4d001f
a=rtcp-fb:39 goog-remb 
a=rtcp-fb:39 ccm fir
a=rtcp-fb:39 nack 
a=rtcp-fb:39 nack pli
a=rtcp-fb:39 nack 
a=rtcp-fb:39 nack pli
a=rtcp-fb:39 transport-cc 
a=rtpmap:40 rtx/90000
a=fmtp:40 apt=39
a=rtcp-fb:40 nack 
a=rtcp-fb:40 nack pli
a=rtcp-fb:40 transport-cc 
a=rtpmap:45 AV1/90000
a=rtcp-fb:45 goog-remb 
a=rtcp-fb:45 ccm fir
a=rtcp-fb:45 nack 
a=rtcp-fb:45 nack pli
a=rtcp-fb:45 nack 
a=rtcp-fb:45 nack pli
a=rtcp-fb:45 transport-cc 
a=rtpmap:46 rtx/90000
a=fmtp:46 apt=45
a=rtcp-fb:46 nack 
a=rtcp-fb:46 nack pli
a=rtcp-fb:46 transport-cc 
a=rtpmap:98 VP9/90000
a=fmtp:98 profile-id=0
a=rtcp-fb:98 goog-remb 
a=rtcp-fb:98 ccm fir
a=rtcp-fb:98 nack 
a=rtcp-fb:98 nack pli
a=rtcp-fb:98 nack 
a=rtcp-fb:98 nack pli
a=rtcp-fb:98 transport-cc 
a=rtpmap:99 rtx/90000
a=fmtp:99 apt=98
a=rtcp-fb:99 nack 
a=rtcp-fb:99 nack pli
a=rtcp-fb:99 transport-cc 
a=rtpmap:100 VP9/90000
a=fmtp:100 profile-id=2
a=rtcp-fb:100 goog-remb 
a=rtcp-fb:100 ccm fir
a=rtcp-fb:100 nack 
a=rtcp-fb:100 nack pli
a=rtcp-fb:100 nack 
a=rtcp-fb:100 nack pli
a=rtcp-fb:100 transport-cc 
a=rtpmap:101 rtx/90000
a=fmtp:101 apt=100
a=rtcp-fb:101 nack 
a=rtcp-fb:101 nack pli
a=rtcp-fb:101 transport-cc 
a=rtpmap:112 H264/90000
a=fmtp:112 level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=64001f
a=rtcp-fb:112 goog-remb 
a=rtcp-fb:112 ccm fir
a=rtcp-fb:112 nack 
a=rtcp-fb:112 nack pli
a=rtcp-fb:112 nack 
a=rtcp-fb:112 nack pli
a=rtcp-fb:112 transport-cc 
a=rtpmap:113 rtx/90000
a=fmtp:113 apt=112
a=rtcp-fb:113 nack 
a=rtcp-fb:113 nack pli
a=rtcp-fb:113 transport-cc 
a=extmap:2 urn:ietf:params:rtp-hdrext:sdes:rtp-stream-id
a=extmap:3 urn:ietf:params:rtp-hdrext:sdes:repaired-rtp-stream-id
a=extmap:4 http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01
a=extmap:1 urn:ietf:params:rtp-hdrext:sdes:mid
a=ssrc-group:FID 914374404 3146946336
a=ssrc:914374404 cname:oJjwEVGKvtByoySf
a=ssrc:914374404 msid:oJjwEVGKvtByoySf SPNUlLGRveMFAjbi
a=ssrc:914374404 mslabel:oJjwEVGKvtByoySf
a=ssrc:914374404 label:SPNUlLGRveMFAjbi
a=ssrc:3146946336 cname:oJjwEVGKvtByoySf
a=ssrc:3146946336 msid:oJjwEVGKvtByoySf SPNUlLGRveMFAjbi
a=ssrc:3146946336 mslabel:oJjwEVGKvtByoySf
a=ssrc:3146946336 label:SPNUlLGRveMFAjbi
a=msid:oJjwEVGKvtByoySf SPNUlLGRveMFAjbi
a=sendrecv
m=audio 9 UDP/TLS/RTP/SAVPF 111 9 0 8
c=IN IP4 0.0.0.0
a=setup:actpass
a=mid:1
a=ice-ufrag:oWRmZjUxbxRgiLKd
a=ice-pwd:ATxASlnMCZCjbBUBmkvmWkwYOUnQXIky
a=rtcp-mux
a=rtcp-rsize
a=rtpmap:111 opus/48000/2
a=fmtp:111 minptime=10;useinbandfec=1
a=rtcp-fb:111 transport-cc 
a=rtpmap:9 G722/8000
a=rtcp-fb:9 transport-cc 
a=rtpmap:0 PCMU/8000
a=rtcp-fb:0 transport-cc 
a=rtpmap:8 PCMA/8000
a=rtcp-fb:8 transport-cc 
a=extmap:4 http://www.ietf.org/id/draft-holmer-rmcat-transport-wide-cc-extensions-01
a=ssrc:38231512 cname:pOJNyirNWqotuPBB
a=ssrc:38231512 msid:pOJNyirNWqotuPBB wNIKoYdtACOIlzxT
a=ssrc:38231512 mslabel:pOJNyirNWqotuPBB
a=ssrc:38231512 label:wNIKoYdtACOIlzxT
a=msid:pOJNyirNWqotuPBB wNIKoYdtACOIlzxT
a=sendrecv`

	jsonData := Parse(txt)

	fmt.Printf("%+v\n", jsonData)

}

func TestHackySdp(t *testing.T) {
	txt := `
	v=0
o=- 3710604898417546434 2 IN IP4 127.0.0.1
s=-
t=0 0
a=msid-semantic: WMS Jvlam5X3SX1OP6pn20zWogvaKJz5Hjf9OnlV
a=group:BUNDLE audio video
m=audio 1 RTP/SAVPF 111 103 104 0 8 107 106 105 13 126
c=IN IP4 0.0.0.0
a=rtpmap:111 opus/48000/2
a=rtpmap:103 ISAC/16000
a=rtpmap:104 ISAC/32000
a=rtpmap:0 PCMU/8000
a=rtpmap:8 PCMA/8000
a=rtpmap:107 CN/48000
a=rtpmap:106 CN/32000
a=rtpmap:105 CN/16000
a=rtpmap:13 CN/8000
a=rtpmap:126 telephone-event/8000
a=fmtp:111 minptime=10
a=rtcp:1 IN IP4 0.0.0.0
a=extmap:1 urn:ietf:params:rtp-hdrext:ssrc-audio-level
a=crypto:1 AES_CM_128_HMAC_SHA1_80 inline:8QVQSHJ2AM8gIumHpYRRdWHyZ5NkLhaTD1AENOWx
a=mid:audio
a=maxptime:60
a=sendrecv
a=ice-ufrag:lat6xwB1/flm+VwG
a=ice-pwd:L5+HonleGeFHa8jPZLc/kr0E
a=candidate:1127303604 1 udp 2122260223 0.0.0.0 60672 typ host generation 0
a=candidate:229815620 1 tcp 1518280447 0.0.0.0 0 typ host tcptype active generation 0
a=candidate:1 1 TCP 2128609279 10.0.1.1 9 typ host tcptype active
a=candidate:2 1 TCP 2124414975 10.0.1.1 8998 typ host tcptype passive
a=candidate:3 1 TCP 2120220671 10.0.1.1 8999 typ host tcptype so
a=candidate:4 1 TCP 1688207359 192.0.2.3 9 typ srflx raddr 10.0.1.1 rport 9 tcptype active
a=candidate:5 1 TCP 1684013055 192.0.2.3 45664 typ srflx raddr 10.0.1.1 rport 8998 tcptype passive generation 5
a=candidate:6 1 TCP 1692401663 192.0.2.3 45687 typ srflx raddr 10.0.1.1 rport 8999 tcptype so
a=ice-options:google-ice
a=ssrc:2754920552 cname:t9YU8M1UxTF8Y1A1
a=ssrc:2754920552 msid:Jvlam5X3SX1OP6pn20zWogvaKJz5Hjf9OnlV Jvlam5X3SX1OP6pn20zWogvaKJz5Hjf9OnlVa0
a=ssrc:2754920552 mslabel:Jvlam5X3SX1OP6pn20zWogvaKJz5Hjf9OnlV
a=ssrc:2754920552 label:Jvlam5X3SX1OP6pn20zWogvaKJz5Hjf9OnlVa0
a=rtcp-mux
m=video 1 RTP/SAVPF 100 116 117
c=IN IP4 0.0.0.0
a=rtpmap:100 VP8/90000
a=rtpmap:116 red/90000
a=rtpmap:117 ulpfec/90000
a=rtcp:12312
a=rtcp-fb:100 ccm fir
a=rtcp-fb:100 nack
a=rtcp-fb:100 goog-remb
a=extmap:2 urn:ietf:params:rtp-hdrext:toffset
a=crypto:1 AES_CM_128_HMAC_SHA1_80 inline:8QVQSHJ2AM8gIumHpYRRdWHyZ5NkLhaTD1AENOWx
a=mid:video
a=sendrecv
a=ice-ufrag:lat6xwB1/flm+VwG
a=ice-pwd:L5+HonleGeFHa8jPZLc/kr0E
a=ice-options:google-ice
a=ssrc:2566107569 cname:t9YU8M1UxTF8Y1A1
a=ssrc:2566107569 msid:Jvlam5X3SX1OP6pn20zWogvaKJz5Hjf9OnlV Jvlam5X3SX1OP6pn20zWogvaKJz5Hjf9OnlVv0
a=ssrc:2566107569 mslabel:Jvlam5X3SX1OP6pn20zWogvaKJz5Hjf9OnlV
a=ssrc:2566107569 label:Jvlam5X3SX1OP6pn20zWogvaKJz5Hjf9OnlVv0
a=rtcp-mux
a=framerate:1234.0
m=application 9 DTLS/SCTP 5000
c=IN IP4 0.0.0.0
b=AS:30
a=setup:active
a=mid:33db2c4da91d73fd
a=ice-ufrag:pDUB98Lc+2dc5+JF
a=ice-pwd:G/CIMBOa9RQINDL4Y8NjpotH
a=fingerprint:sha-256 F0:37:78:FE:3D:13:E9:10:B5:0C:4C:9E:48:37:E7:A0:F8:16:DC:1A:2C:69:67:B0:DF:E6:CB:73:F8:EF:BA:02
a=sctpmap:5000 webrtc-datachannel 1024
a=framerate:29.97
`

	sdp := Parse(txt)

	fmt.Println(sdp)
}
