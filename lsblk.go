package lsblk

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

// Lsblk main JSON struct to capture the output of `lsblk`
type Lsblk struct {
	Blockdevices []Blockdevice
}

// Blockdevice JSON strruct with details for every device
type Blockdevice struct {
	Name         string // device name
	Kname        string // internal kernel device name
	Pkname       string // internal parent kernel device name
	Path         string // path to the device node
	MajMin       string `json:"maj:min"` // major:minor device number
	Fsavail      string // filesystem size available
	Fssize       string // filesystem size in bytes
	Fstype       string // filesystem type
	Fsused       string // filesystem size used
	Fsusep       string `json:"fsuse%"` // filesystem use percentage
	Fsver        string // filesystem version
	Mountpoint   string // path where the device is mounted
	Label        string // filesystem LABEL
	UUID         string // filesystem UUID
	Ptuuid       string // partition table identifier (usually UUID)
	Pttype       string // partition table type
	Parttype     string // partition type code or UUID
	Parttypename string // partition type name
	Partlabel    string // partition LABEL
	Partuuid     string // partition UUID
	Partflags    string // partition flags
	Ra           int    // read-ahead of the device
	Ro           bool   // read-only device
	Rm           bool   // removable device
	Hotplug      bool   // removable or hotplug device (usb, pcmcia, ...)
	Rota         bool   // rotational device
	Rand         bool   // adds randomness
	Model        string // device identifier
	Serial       string // disk serial number
	Size         int    // size of the device in bytes
	State        string // state of the device e.g. suspended, running, live
	Owner        string // user name
	Group        string // group name
	Mode         string // device node permissions e.g. brw-rw----
	Alignment    int    // alignment offset
	Minio        int    `json:"min-io"`  // minimum I/O size
	Optio        int    `json:"opt-io"`  // optimal I/O size
	Physec       int    `json:"phy-sec"` // physical sector size
	Logsec       int    `json:"log-sec"` // logical sector size
	Sched        string // I/O scheduler name e.g. mq-deadline
	Rqsize       int    `json:"rq-size"` // request queue size
	Type         string // device type e.g. loop, disk, part, crypt, lvm
	Discaln      int    `json:"disc-aln"`  // discard alignment offset
	Discgran     int    `json:"disc-gran"` // discard granularity
	Discmax      int    `json:"disc-max"`  // discard max bytes
	Disczero     bool   `json:"disc-zero"` // discard zeroes data
	Wsame        int    // write same max bytes
	Wwn          string // unique storage identifier
	Hctl         string // Host:Channel:Target:Lun for SCSI
	Tran         string // device transport type e.g. usb, nvme
	Subsystems   string // de-duplicated chain of subsystems e.g. block, block:scsi:usb:pci, block:nvme:pci
	Rev          string // device revision
	Vendor       string // device vendor
	Zoned        string // zone model
	Dax          bool   // dax-capable device
	Children     []Blockdevice
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// ByteCountSI .
func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

// ByteCountIEC .
func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

// HasPartitions .
func (b Blockdevice) HasPartitions() bool {
	if len(b.Children) > 0 {
		return true
	}
	return false
}

// IsRunning .
func (b Blockdevice) IsRunning() bool {
	if b.State == "running" {
		return true
	}
	return false
}

// IsMounted .
func (b Blockdevice) IsMounted() bool {
	if b.Mountpoint != "" {
		return true
	}
	return false
}

// IsUsbTran .
func (b Blockdevice) IsUsbTran() bool {
	if b.Tran == "usb" {
		return true
	}
	return false
}

// MountedPartitions .
func MountedPartitions() []Blockdevice {
	lsblk := GetLsblk()
	var partitions []Blockdevice

	for _, b := range lsblk.Blockdevices {
		if b.IsUsbTran() {
			fmt.Println(b.Name, b.Fstype, b.Rm, b.Hotplug)
			if b.HasPartitions() {
				for _, p := range b.Children {
					partitions = append(partitions, p)
				}
			}
		}
	}
	return partitions
}

// GetLsblk .
func GetLsblk() Lsblk {
	var lsblk Lsblk
	lsblkj := GetLsblkOutput()
	err := json.Unmarshal(lsblkj, &lsblk)
	check(err)
	return lsblk
}

// GetLsblkOutput .
func GetLsblkOutput() []byte {
	lsblkj, err := exec.Command("lsblk", "-pabOJ").Output()
	check(err)
	return lsblkj
}
