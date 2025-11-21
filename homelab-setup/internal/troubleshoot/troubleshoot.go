package troubleshoot

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"

	"github.com/zoro11031/homelab-coreos-minipc/homelab-setup/internal/config"
	"github.com/zoro11031/homelab-coreos-minipc/homelab-setup/internal/ui"
)

const (
	FileServerIP = "192.168.7.179"
	GatewayIP    = "192.168.1.1"
	VPSIP        = "64.23.212.68"
	GoogleDNS    = "8.8.8.8"
	GoogleHost   = "google.com"
)

// Run executes the troubleshooting suite
func Run(cfg *config.Config, ui *ui.UI) error {
	ui.Header("Network Troubleshooting Suite")

	// 1. Network Instability
	checkNetworkInstability(ui)

	// 2. DNS Diagnostics
	checkDNS(ui)

	// 3. Port Scanning
	checkPortScanning(ui)

	return nil
}

// pingResult holds the result of a ping test
type pingResult struct {
	PacketLoss float64
	AvgLatency time.Duration
	Unstable   bool
}

// sendPing sends ICMP echo requests to the target
func sendPing(addr string, count int, timeout time.Duration) (*pingResult, error) {
	c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, fmt.Errorf("listen failed: %w (root privileges required)", err)
	}
	defer c.Close()

	var latencies []time.Duration
	received := 0

	// Resolve address
	dst, err := net.ResolveIPAddr("ip4", addr)
	if err != nil {
		return nil, fmt.Errorf("resolve failed: %w", err)
	}

	for i := 0; i < count; i++ {
		wm := icmp.Message{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmp.Echo{
				ID: os.Getpid() & 0xffff, Seq: i,
				Data: []byte("homelab-setup-ping"),
			},
		}
		wb, err := wm.Marshal(nil)
		if err != nil {
			continue
		}

		start := time.Now()
		if _, err := c.WriteTo(wb, dst); err != nil {
			continue
		}

		// Set read deadline
		if err := c.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			continue
		}

		rb := make([]byte, 1500)
		n, _, err := c.ReadFrom(rb)
		if err != nil {
			// Timeout or error
			continue
		}

		duration := time.Since(start)

		rm, err := icmp.ParseMessage(ipv4.ICMPTypeEchoReply.Protocol(), rb[:n])
		if err != nil {
			continue
		}

		switch rm.Type {
		case ipv4.ICMPTypeEchoReply:
			// Verify ID/Seq if strict, but for now just assume it's ours if we got a reply
			if pkt, ok := rm.Body.(*icmp.Echo); ok {
				if pkt.ID == (os.Getpid()&0xffff) && pkt.Seq == i {
					latencies = append(latencies, duration)
					received++
				}
			}
		}

		// Slight delay between pings
		time.Sleep(200 * time.Millisecond)
	}

	loss := float64(count-received) / float64(count) * 100.0
	var totalLat time.Duration
	for _, l := range latencies {
		totalLat += l
	}
	var avgLat time.Duration
	if received > 0 {
		avgLat = totalLat / time.Duration(received)
	}

	unstable := loss > 0 || avgLat > 100*time.Millisecond

	return &pingResult{
		PacketLoss: loss,
		AvgLatency: avgLat,
		Unstable:   unstable,
	}, nil
}

func checkNetworkInstability(ui *ui.UI) {
	ui.Step(fmt.Sprintf("1. Network Instability Check (Target: %s)", FileServerIP))

	ui.Info("Sending ICMP packets...")
	res, err := sendPing(FileServerIP, 5, 1*time.Second)
	if err != nil {
		ui.Errorf("Ping failed: %v", err)
		return
	}

	ui.Infof("  Packet Loss: %.0f%%", res.PacketLoss)
	ui.Infof("  Avg Latency: %v", res.AvgLatency)

	if res.Unstable {
		ui.Warning("  Status: UNSTABLE (Loss > 0% or Latency > 100ms)")
	} else {
		ui.Success("  Status: STABLE")
	}
}

func checkDNS(ui *ui.UI) {
	ui.Step(fmt.Sprintf("2. DNS Diagnostics (Target: %s via %s)", GoogleHost, GatewayIP))

	// 1. Attempt to resolve standard hostname
	start := time.Now()
	ips, err := net.LookupHost(GoogleHost)
	duration := time.Since(start)

	if err == nil && len(ips) > 0 {
		ui.Successf("  ✓ Resolution successful: %s -> %v (%v)", GoogleHost, ips[0], duration)
		return
	}

	ui.Error("  ✗ Resolution failed!")
	ui.Info("  Starting tiered diagnostics...")

	// Tier 1: Check local resolv.conf
	ui.Info("  [Tier 1] Checking /etc/resolv.conf:")
	content, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		ui.Warningf("    Could not read /etc/resolv.conf: %v", err)
	} else {
		ui.Print(string(content))
	}

	// Tier 2: Direct resolution via Public DNS
	ui.Info(fmt.Sprintf("  [Tier 2] Attempting direct resolution via %s...", GoogleDNS))

	// Use net.Resolver to simulate checking external DNS
	// We dial port 53 UDP to see if we can even reach it
	conn, err := net.DialTimeout("udp", GoogleDNS+":53", 2*time.Second)
	if err != nil {
		ui.Errorf("    ✗ Failed to reach %s:53 - Likely a gateway/internet connectivity issue", GoogleDNS)
	} else {
		conn.Close()
		ui.Successf("    ✓ Successfully reached %s:53 - Local DNS configuration might be broken", GoogleDNS)

		// Try an actual lookup using a custom resolver
		r := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Second * 2,
				}
				return d.DialContext(ctx, "udp", GoogleDNS+":53")
			},
		}

		ips, err := r.LookupHost(context.Background(), GoogleHost)
		if err == nil && len(ips) > 0 {
			ui.Successf("    ✓ Direct resolution via %s successful: %v", GoogleDNS, ips[0])
		} else {
			ui.Warningf("    ✗ Direct resolution via %s failed: %v", GoogleDNS, err)
		}
	}
}

func checkPortScanning(ui *ui.UI) {
	ui.Step(fmt.Sprintf("3. Port Scanning (Target: VPS %s)", VPSIP))

	ports := []struct {
		Port    string
		Service string
	}{
		{"80", "HTTP (NPM)"},
		{"443", "HTTPS (NPM)"},
		{"9000", "Portainer"},
		{"9443", "Portainer (SSL)"},
	}

	for _, p := range ports {
		address := net.JoinHostPort(VPSIP, p.Port)
		conn, err := net.DialTimeout("tcp", address, 2*time.Second)

		status := "CLOSED/FILTERED"

		if err == nil {
			status = "OPEN"
			conn.Close()
			ui.Successf("  %-15s : %s (%s)", p.Service, status, p.Port)
		} else {
			ui.Infof("  %-15s : %s (%s) - %v", p.Service, status, p.Port, err)
		}
	}
}
