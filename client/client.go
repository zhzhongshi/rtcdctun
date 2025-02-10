package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/pion/webrtc/v4"
)

type dataChannelWriter struct {
	*webrtc.DataChannel
}

func (s *dataChannelWriter) Write(b []byte) (int, error) {
	err := s.DataChannel.Send(b)
	return len(b), err
}
func main() {
	var ListenAddr string
	flag.StringVar(&ListenAddr, "listen", "127.0.0.1:8445", "Listening address")
	flag.Parse()
	// Configure and create a new PeerConnection.
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}
	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		handleError(err)
	}

	// Create DataChannel.
	ch1, err := pc.CreateDataChannel("ch1", nil)
	if err != nil {
		handleError(err)
	}
	ch1.OnClose(func() {
		fmt.Println("sendChannel has closed")
	})
	ch1.OnOpen(func() {
		fmt.Println("sendChannel has opened")
		//ch1.SendText("ch1 hello")
	})
	ch1.OnMessage(func(msg webrtc.DataChannelMessage) {
		log(fmt.Sprintf("Message from DataChannel %s payload %s", ch1.Label(), string(msg.Data)))

		//todo
	})
	// // Create DataChannel.
	// ch2, err := pc.CreateDataChannel("ch2", nil)
	// if err != nil {
	// 	handleError(err)
	// }
	// ch2.OnClose(func() {
	// 	fmt.Println("sendChannel has closed")
	// })
	// ch2.OnOpen(func() {
	// 	fmt.Println("sendChannel has opened")
	// 	ch2.SendText("ch2 hello")
	// })
	// ch2.OnMessage(func(msg webrtc.DataChannelMessage) {
	// 	log(fmt.Sprintf("Message from DataChannel %s payload %s", ch2.Label(), string(msg.Data)))

	// 	//todo
	// })

	// Create offer
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		handleError(err)
	}
	if err := pc.SetLocalDescription(offer); err != nil {
		handleError(err)
	}
	// Add handlers for setting up the connection.
	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log(fmt.Sprint(state))
	})
	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {

		}
	})

	// if err := sendChannel.SendText(message); err != nil {
	// 	handleError(err)
	// }
	encodedDescr := encode(pc.LocalDescription())
	fmt.Println(encodedDescr)

	answer := webrtc.SessionDescription{}
	decode(readUntilNewline(), &answer)
	if err := pc.SetRemoteDescription(answer); err != nil {
		handleError(err)
	}
	l, err := net.Listen("tcp", ListenAddr)
	if err != nil {
		handleError(err)
	}
	for {
		sock, err := l.Accept()
		if err != nil {
			handleError(err)
		}

		go func(conn net.Conn, pc *webrtc.PeerConnection) {
			dc, err := pc.CreateDataChannel(string(conn.RemoteAddr().String()), nil)
			fmt.Println("New data channel created:" + dc.Label())
			if err != nil {
				return
			}
			//dc.Lock()
			dc.OnOpen(func() {
				io.Copy(&dataChannelWriter{dc}, sock)
			})
			dc.OnMessage(func(payload webrtc.DataChannelMessage) {
				if !payload.IsString {
					_, err := sock.Write(payload.Data)
					if err != nil {
						handleError(err)
					}
				}
			})
			dc.OnClose(func() {
				fmt.Println("Data channel closed")
				conn.Close()
			})
		}(sock, pc)
	}

	// Stay alive
}

func log(msg string) {
	fmt.Println(msg)
}

func handleError(err error) {
	log("Unexpected error. Check console.")
	panic(err)
}

// Read from stdin until we get a newline
func readUntilNewline() (in string) {
	var err error

	r := bufio.NewReader(os.Stdin)
	for {
		in, err = r.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			panic(err)
		}

		if in = strings.TrimSpace(in); len(in) > 0 {
			break
		}
	}

	fmt.Println("")
	return
}

// JSON encode + base64 a SessionDescription
func encode(obj *webrtc.SessionDescription) string {
	b, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(b)
}

// Decode a base64 and unmarshal JSON into a SessionDescription
func decode(in string, obj *webrtc.SessionDescription) {
	b, err := base64.StdEncoding.DecodeString(in)
	if err != nil {
		panic(err)
	}

	if err = json.Unmarshal(b, obj); err != nil {
		panic(err)
	}
}
