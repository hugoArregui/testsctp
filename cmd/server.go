package cmd

import (
	"github.com/pion/logging"
	"github.com/pion/sctp"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net"
	"time"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use: "server",
	Run: func(cmd *cobra.Command, args []string) {
		raddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:10001")
		if err != nil {
			panic(err)
		}
		laddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:10002")
		if err != nil {
			panic(err)
		}
		conn, err := net.DialUDP("udp4", raddr, laddr)
		if err != nil {
			panic(err)
		}

		s, err := sctp.Server(sctp.Config{
			NetConn:       conn,
			LoggerFactory: logging.NewDefaultLoggerFactory(),
		})
		if err != nil {
			panic(err)
		}

		go func() {
			since := time.Now()
			for range time.NewTicker(1000 * time.Millisecond).C {
				rbps := float64(s.BytesReceived()*8) / time.Since(since).Seconds()
				log.Printf("Received Mbps: %.03f, totalBytesReceived: %d", rbps/1024/1024,
					s.BytesReceived())
			}
		}()

		for {
			stream, err := s.AcceptStream()
			if err != nil {
				panic(err)
			}

			stream.SetReliabilityParams(false, 2, 10)
			go func() {
				buf := make([]byte, 1024*1024*1024)
				for {
					_, err := stream.Read(buf)
					if err != nil {
						fmt.Println("ERRROR")
						panic(err)
					}
				}
			}()
		}
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
