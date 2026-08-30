# sysmon (go-desktop)

A simple desktop GUI app written in Go that shows live system resource
usage: CPU (overall + per-core + temperature), RAM, swap, disk usage per
partition, and network throughput.

Built with [Fyne](https://fyne.io) for the GUI and
[gopsutil](https://github.com/shirou/gopsutil) for the metrics.
Stats refresh every second.

## Run

```sh
go run -tags migrated_fynedo .
```

## Build a binary

```sh
go build -tags migrated_fynedo -o sysmon .
./sysmon
```

The `migrated_fynedo` tag tells Fyne this app uses the `fyne.Do`
threading model (all UI updates from background goroutines are wrapped
in `fyne.Do`), which silences its startup migration notice.

## Requirements

- Go 1.21+
- A C compiler and OpenGL/X11 dev packages (already present on most Linux
  desktops). On Debian/Ubuntu, if missing:
  `sudo apt install gcc libgl1-mesa-dev xorg-dev`
