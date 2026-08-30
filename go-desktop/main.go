// Command sysmon is a small desktop app that shows live CPU, memory,
// swap, disk and network usage.
package main

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
	"github.com/shirou/gopsutil/v4/sensors"
)

const refreshInterval = time.Second

func main() {
	a := app.NewWithID("dev.aurnob.sysmon")
	w := a.NewWindow("System Monitor")

	m := newMonitor()
	w.SetContent(container.NewScroll(m.root))
	w.Resize(fyne.NewSize(540, 800))

	done := make(chan struct{})
	w.SetCloseIntercept(func() {
		close(done)
		w.Close()
	})

	cpu.Percent(0, false) // prime the counter baseline for the first reading

	go func() {
		tick := func() {
			snap := m.sample()
			fyne.Do(func() { m.render(snap) })
		}
		tick() // paint real values immediately

		ticker := time.NewTicker(refreshInterval)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				tick()
			}
		}
	}()

	w.ShowAndRun()
}

// ---------------------------------------------------------------- UI setup

type monitor struct {
	root *fyne.Container

	hostLine  *widget.Label
	cpuBar    *widget.ProgressBar
	cpuDetail *widget.Label
	coreBars  []*widget.ProgressBar
	memBar    *widget.ProgressBar
	memLabel  *widget.Label
	swapBar   *widget.ProgressBar
	swapLabel *widget.Label
	netDown   *widget.Label
	netUp     *widget.Label
	netTotal  *widget.Label
	disks     []diskRow

	prevRx, prevTx uint64
	prevAt         time.Time
}

type diskRow struct {
	mount string
	bar   *widget.ProgressBar
	label *widget.Label
}

func newMonitor() *monitor {
	m := &monitor{}

	info, err := host.Info()
	if err != nil {
		info = &host.InfoStat{Hostname: "unknown"}
	}

	m.hostLine = widget.NewLabel("")
	m.hostLine.Wrapping = fyne.TextWrapWord
	header := widget.NewCard(info.Hostname,
		fmt.Sprintf("%s %s · %s · %s",
			info.Platform, info.PlatformVersion, info.KernelVersion, runtime.GOARCH),
		m.hostLine)

	// CPU card: overall meter + per-core meters.
	m.cpuBar = newMeter()
	m.cpuDetail = widget.NewLabel("")
	body := []fyne.CanvasObject{m.cpuBar, m.cpuDetail}
	if n, _ := cpu.Counts(true); n > 0 {
		cols := int(math.Ceil(math.Sqrt(float64(n))))
		cells := make([]fyne.CanvasObject, 0, n)
		for i := 0; i < n; i++ {
			b := newMeter()
			m.coreBars = append(m.coreBars, b)
			cells = append(cells, b)
		}
		body = append(body, container.NewGridWithColumns(cols, cells...))
	}
	cpuCard := widget.NewCard("CPU", "", container.NewVBox(body...))

	// Memory card: RAM + swap.
	m.memBar = newMeter()
	m.memLabel = widget.NewLabel("")
	m.swapBar = newMeter()
	m.swapLabel = widget.NewLabel("")
	memCard := widget.NewCard("Memory", "",
		container.NewVBox(m.memBar, m.memLabel, m.swapBar, m.swapLabel))

	// One card per physical partition.
	var diskCards []fyne.CanvasObject
	for _, p := range physicalPartitions() {
		row := diskRow{
			mount: p.Mountpoint,
			bar:   newMeter(),
			label: widget.NewLabel(""),
		}
		m.disks = append(m.disks, row)
		diskCards = append(diskCards, widget.NewCard(
			fmt.Sprintf("Disk %s", p.Mountpoint), p.Device+" · "+p.Fstype,
			container.NewVBox(row.bar, row.label)))
	}

	// Network card.
	m.netDown = widget.NewLabel("")
	m.netUp = widget.NewLabel("")
	m.netTotal = widget.NewLabel("")
	netCard := widget.NewCard("Network", "",
		container.NewVBox(m.netDown, m.netUp, m.netTotal))

	rows := []fyne.CanvasObject{header, cpuCard, memCard}
	rows = append(rows, diskCards...)
	rows = append(rows, netCard)
	m.root = container.NewVBox(rows...)
	return m
}

func newMeter() *widget.ProgressBar {
	b := widget.NewProgressBar()
	b.Max = 100
	return b
}

// physicalPartitions lists real block-device filesystems once, skipping
// virtual mounts (tmpfs, squashfs snaps, /proc, …) and duplicate devices.
func physicalPartitions() []disk.PartitionStat {
	parts, err := disk.Partitions(false)
	if err != nil {
		return nil
	}
	virtual := map[string]bool{
		"squashfs": true, "tmpfs": true, "devtmpfs": true, "ramfs": true,
		"proc": true, "sysfs": true, "overlay": true, "iso9660": true,
	}
	seen := map[string]bool{}
	out := []disk.PartitionStat{}
	for _, p := range parts {
		if virtual[p.Fstype] || !strings.HasPrefix(p.Device, "/dev/") || seen[p.Device] {
			continue
		}
		seen[p.Device] = true
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { // shortest mount first, so "/" leads
		if len(out[i].Mountpoint) != len(out[j].Mountpoint) {
			return len(out[i].Mountpoint) < len(out[j].Mountpoint)
		}
		return out[i].Mountpoint < out[j].Mountpoint
	})
	return out
}

// ------------------------------------------------------------- sampling

type snapshot struct {
	cpuPct    float64
	corePct   []float64
	tempC     float64
	tempName  string
	mem       *mem.VirtualMemoryStat
	rxRate    float64
	txRate    float64
	rxTot     uint64
	txTot     uint64
	procs     int
	upSeconds uint64
}

// sample gathers one round of readings. Rates are derived from the
// previous network counter snapshot held by the monitor.
func (m *monitor) sample() snapshot {
	s := snapshot{tempC: -1}

	if v, err := cpu.Percent(0, false); err == nil && len(v) == 1 {
		s.cpuPct = v[0]
	}
	s.corePct, _ = cpu.Percent(0, true)

	if v, err := mem.VirtualMemory(); err == nil {
		s.mem = v
	}

	if ts, err := sensors.SensorsTemperatures(); err == nil {
		for _, t := range ts {
			if t.Temperature > s.tempC && t.Temperature > 0 && t.Temperature < 120 {
				s.tempC, s.tempName = t.Temperature, t.SensorKey
			}
		}
	}

	if io, err := net.IOCounters(false); err == nil && len(io) > 0 {
		now := time.Now()
		s.rxTot, s.txTot = io[0].BytesRecv, io[0].BytesSent
		if !m.prevAt.IsZero() {
			if dt := now.Sub(m.prevAt).Seconds(); dt > 0 {
				s.rxRate = float64(s.rxTot-m.prevRx) / dt
				s.txRate = float64(s.txTot-m.prevTx) / dt
			}
		}
		m.prevRx, m.prevTx, m.prevAt = s.rxTot, s.txTot, now
	}

	if pids, err := process.Pids(); err == nil {
		s.procs = len(pids)
	}
	s.upSeconds, _ = host.Uptime()
	return s
}

// ------------------------------------------------------------- rendering

func (m *monitor) render(s snapshot) {
	// Header.
	m.hostLine.SetText(fmt.Sprintf("Uptime %s · %d processes",
		humanDuration(time.Duration(s.upSeconds)*time.Second), s.procs))

	// CPU.
	m.cpuBar.SetValue(s.cpuPct)
	detail := fmt.Sprintf("%.1f%% across %d logical cores", s.cpuPct, len(m.coreBars))
	if s.tempC >= 0 {
		detail += fmt.Sprintf(" · %s %.0f °C", s.tempName, s.tempC)
	}
	m.cpuDetail.SetText(detail)
	for i, b := range m.coreBars {
		if i < len(s.corePct) {
			b.SetValue(s.corePct[i])
		}
	}

	// Memory + swap.
	if s.mem != nil {
		m.memBar.SetValue(s.mem.UsedPercent)
		m.memLabel.SetText(fmt.Sprintf("%s used of %s · %s free",
			humanBytes(float64(s.mem.Used)), humanBytes(float64(s.mem.Total)),
			humanBytes(float64(s.mem.Available))))
		if s.mem.SwapTotal > 0 {
			pct := float64(s.mem.SwapTotal-s.mem.SwapFree) / float64(s.mem.SwapTotal) * 100
			m.swapBar.SetValue(pct)
			m.swapLabel.SetText(fmt.Sprintf("Swap: %s used of %s",
				humanBytes(float64(s.mem.SwapTotal-s.mem.SwapFree)),
				humanBytes(float64(s.mem.SwapTotal))))
		} else {
			m.swapBar.SetValue(0)
			m.swapLabel.SetText("No swap configured")
		}
	}

	// Disks.
	for _, d := range m.disks {
		u, err := disk.Usage(d.mount)
		if err != nil {
			d.bar.SetValue(0)
			d.label.SetText("usage unavailable")
			continue
		}
		d.bar.SetValue(u.UsedPercent)
		d.label.SetText(fmt.Sprintf("%s used of %s · %s free",
			humanBytes(float64(u.Used)), humanBytes(float64(u.Total)),
			humanBytes(float64(u.Free))))
	}

	// Network.
	m.netDown.SetText(fmt.Sprintf("↓ %s/s", humanBytes(s.rxRate)))
	m.netUp.SetText(fmt.Sprintf("↑ %s/s", humanBytes(s.txRate)))
	m.netTotal.SetText(fmt.Sprintf("Session totals · received %s · sent %s",
		humanBytes(float64(s.rxTot)), humanBytes(float64(s.txTot))))
}

// ------------------------------------------------------------ formatting

func humanBytes(b float64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%.0f B", b)
	}
	div, exp := float64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", b/div, "KMGTPE"[exp])
}

func humanDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, mins)
	default:
		return fmt.Sprintf("%dm", mins)
	}
}
