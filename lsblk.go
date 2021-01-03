package lsblk

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/tidwall/gjson"
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
	Subsystems   string // de-dulicated chain of subsystems e.g. block, block:scsi:usb:pci, block:nvme:pci
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

// if it makes sense to change the JSON item into something else
// in `type Blockdevice struct` the custom item should get its
// respective type. in this case it changes from string to bool
//
// func (u *Blockdevice) UnmarshalJSON(data []byte) error {
// 	type Alias Blockdevice
// 	aux := &struct {
// 		State string
// 		*Alias
// 	}{
// 		Alias: (*Alias)(u),
// 	}

// 	if err := json.Unmarshal(data, &aux); err != nil {
// 		return err
// 	}

// 	if aux.State == "running" {
// 		u.State = true
// 	} else {
// 		u.State = false
// 	}
// 	return nil
// }

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

// MountedPartitions .
func MountedPartitions() []Blockdevice {
	lsblk := GetLsblk()
	var partitions []Blockdevice
	// var mPartitions []Blockdevice

	// using go-funk utilities
	// blockdevices := funk.Filter(lsblk.Blockdevices, func(b Blockdevice) bool { return b.HasPartitions() }).([]Blockdevice)

	// funk.ForEach(blockdevices, func(b Blockdevice) {
	// 	for _, c := range b.Children {
	// 		if c.IsMounted() {
	// 			partitions = append(partitions, c)
	// 		}
	// 	}
	// })

	// mPartitions = funk.Filter(partitions, func(p Blockdevice) bool { return p.IsMounted() }).([]Blockdevice)

	// using koazee streams
	// koazee.StreamOf(lsblk.Blockdevices).
	// 	Filter(func(b Blockdevice) bool { return b.HasPartitions() }).
	// 	ForEach(func(p Blockdevice) {
	// 		for _, c := range p.Children {
	// 			if c.IsMounted() {
	// 				partitions = append(partitions, c)
	// 			}
	// 		}
	// 	}).Do()

	// using plain go's for/if
	for _, b := range lsblk.Blockdevices {
		if b.HasPartitions() {
			for _, p := range b.Children {
				if p.IsMounted() {
					partitions = append(partitions, p)
				}
			}
		}
	}

	return partitions
}

// HasChildren .
func HasChildren(g []byte, el string) bool {
	if len(gjson.GetBytes(g, el+".#.name").Array()) > 0 {
		return true
	}
	return false
}

// GmountedPartitions .
func GmountedPartitions() {
	g := GetLsblkOutput()
	el := ".children|@flatten"
	n := 0
	for {
		if HasChildren(g, fmt.Sprintf("blockdevices.#%s", el)) {
			el = el + ".#.children|@flatten"
		} else {
			fmt.Printf("Final el: %s\n", strings.TrimSuffix(el, ".#.children|@flatten"))
			n = strings.Count(el, ".#.children|@flatten")
			break
		}
	}

	if n == 0 {
		fmt.Println("What if no children?")
	}

	fmt.Println("~~~~~~~~~~~~~~~~~~~~~\n")
	el = ".children.#"
	for n > -1 {
		ele := fmt.Sprintf("blockdevices.#%s.children", strings.Repeat(el, n))
		r := gjson.GetBytes(g, ele)
		fmt.Println(n, "**********************************************\n")
		rel := strings.Repeat("#.", n+2)
		fmt.Println(r.Get(fmt.Sprintf("%sname|@flatten", rel)))
		fmt.Println(r.Get(fmt.Sprintf("%smountpoint|@flatten", rel)))
		fmt.Println("**********************************************\n")
		n = n - 1
	}
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
