package firewall

import (
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gopal-lohar/hackathon-2025/internal/firewall/db"
	"github.com/gopal-lohar/hackathon-2025/internal/shared/protocol"
	"github.com/gopal-lohar/hackathon-2025/internal/shared/store"
	"github.com/gopal-lohar/hackathon-2025/internal/shared/utils"
	"github.com/gopal-lohar/hackathon-2025/internal/shared/utils/logger"
	"github.com/gopal-lohar/hackathon-2025/internal/windows"
	"github.com/sirupsen/logrus"
)

type Firewall struct {
	logger        *logrus.Logger
	apiServerConn net.Conn
	windows       *windows.Windows
	ruleStore     *store.RuleStore
}

func NewFirewall() *Firewall {
	db, err := db.NewDB()
	if err != nil {
		logrus.Fatalf("Error creating db connection: %v", err)
	}
	windows := windows.NewWindows()
	ruleStore := store.NewRuleStore(db)
	return &Firewall{
		logger:    logger.NewLogger(),
		windows:   windows,
		ruleStore: ruleStore,
	}
}

func (f *Firewall) Run() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	f.logger.Info("Firewall started")

	// Connect to the API Server
	err := f.connectToAPIServer()
	if err != nil {
		f.logger.Fatalf("Error connecting to api server: %v", err)
		return
	}
	// Listen and handle messages sent by API Server
	go f.listenAPIServerMsgs()

	<-sigChan
	f.logger.Info("Stopping firewall...")
}

func (f *Firewall) connectToAPIServer() error {
	conn, err := net.Dial("tcp", "localhost:8081")
	if err != nil {
		f.logger.Fatalf("Error connecting to api server: %v", err)
	}
	f.logger.Info("Connected to server")
	f.apiServerConn = conn
	return nil
}
func isIPAddress(host string) bool {
	ip := net.ParseIP(host)
	return ip != nil
}
func (f *Firewall) listenAPIServerMsgs() {
	defer f.apiServerConn.Close()
	for {
		netMsg, err := utils.ReceiveNetMsg(f.apiServerConn)
		if err != nil {
			if err == io.EOF {
				f.logger.Info("Connection with api server was closed")
			} else {
				f.logger.Warnf("Error receiving message from api server: %v", err)
			}
			return
		}

		f.logger.Infof("Received message from api server: %v", netMsg)
		//  Handle message
		switch m := netMsg.GetMessageType().(type) { // Changed to a type switch with variable m to extract the actual type.
		case *protocol.NetworkMessage_Policy:
			f.logger.Infof("Received a policy message: %+v", m)
			// add a new rule to db
			// fix bad code (impossible)
			program := netMsg.GetPolicy().GetAppPath()
			protocol := netMsg.GetPolicy().GetProtocol()
			remoteIp := netMsg.GetPolicy().GetRemoteIp()
			action := netMsg.GetPolicy().GetAction()
			ips := []string{remoteIp} // Default to the original value

			// Use net.LookupIP to resolve domain name if needed
			if !isIPAddress(remoteIp) {
				resolvedIPs, err := net.LookupIP(remoteIp)
				if err != nil {
					f.logger.Warnf("Error resolving domain name %s: %v", remoteIp, err)
				} else {
					// Replace the list with the resolved IPs
					ips = make([]string, 0, len(resolvedIPs))
					for _, ip := range resolvedIPs {
						// Only include IPv4 addresses if that's what netsh requires
						if ip.To4() != nil {
							ips = append(ips, ip.String())
						}
					}
					if len(ips) == 0 {
						f.logger.Warnf("No IPv4 addresses found for domain: %s", remoteIp)
						ips = []string{remoteIp} // Fallback to original
					}
				}
			}

			// Store the original domain name in the database
			temp := store.Temp{
				EndpointID: m.Policy.GetEndpointId(),
				Enabled:    false,
			}

			// Create a rule for each resolved IP
			for _, ip := range ips {
				id, err := f.ruleStore.AddRule(program, protocol, ip, action, true, temp)
				if err != nil {
					f.logger.Warnf("Error adding rule for IP %s to db: %v", ip, err)
					continue
				}

				if id == -1 {
					f.logger.Warnf("Rule for IP %s already exists in db", ip)
					continue
				}

				name := strconv.Itoa(id)
				err = f.windows.AddNewRule(name, action, program, ip, protocol)
				if err != nil {
					f.logger.Warnf("Error adding rule for IP %s to windows: %v", ip, err)
				} else {
					f.logger.Infof("Created a rule with name: %s for IP: %s (domain: %s)", name, ip, remoteIp)
				}
			}
		}

	}
}
