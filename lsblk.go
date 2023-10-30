package lsblk

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
)

// Lsblk main JSON struct to capture the output of `lsblk`
type Lsblk struct {
	Blockdevices []Blockdevice
}

// Blockdevice JSON strruct with details for every device
type Blockdevice struct {
	Name         string      // device name
	Kname        string      // internal kernel device name
	Pkname       string      // internal parent kernel device name
	Path         string      // path to the device node
	MajMin       string      `json:"maj:min"` // major:minor device number
	Fsavail      Num         // filesystem size available
	Fssize       Num         // filesystem size in bytes
	Fstype       string      // filesystem type
	Fsused       string      // filesystem size used
	Fsusep       string      `json:"fsuse%"` // filesystem use percentage
	Fsver        string      // filesystem version
	Mountpoint   string      // path where the device is mounted
	Label        string      // filesystem LABEL
	UUID         string      // filesystem UUID
	Ptuuid       string      // partition table identifier (usually UUID)
	Pttype       string      // partition table type
	Parttype     string      // partition type code or UUID
	Parttypename string      // partition type name
	Partlabel    string      // partition LABEL
	Partuuid     string      // partition UUID
	Partflags    string      // partition flags
	Ra           json.Number // read-ahead of the devic
	Ro           Bool        // read-only device
	Rm           Bool        // removable device
	Hotplug      Bool        // removable or hotplug device (usb, pcmcia, ...)
	Rota         Bool        // rotational device
	Rand         Bool        // adds randomness
	Model        string      // device identifier
	Serial       string      // disk serial number
	Size         Num         // size of the device in bytes
	State        string      // state of the device e.g. suspended, running, live
	Owner        string      // user name
	Group        string      // group name
	Mode         string      // device node permissions e.g. brw-rw----
	Alignment    json.Number // alignment offset
	Minio        json.Number `json:"min-io"`  // minimum I/O size
	Optio        json.Number `json:"opt-io"`  // optimal I/O size
	Physec       json.Number `json:"phy-sec"` // physical sector size
	Logsec       json.Number `json:"log-sec"` // logical sector size
	Sched        string      // I/O scheduler name e.g. mq-deadline
	Rqsize       json.Number `json:"rq-size"` // request queue size
	Type         string      // device type e.g. loop, disk, part, crypt, lvm
	Discaln      json.Number `json:"disc-aln"`  // discard alignment offset
	Discgran     json.Number `json:"disc-gran"` // discard granularity
	Discmax      json.Number `json:"disc-max"`  // discard max bytes
	Disczero     Bool        `json:"disc-zero"` // discard zeroes data
	Wsame        json.Number // write same max bytes
	Wwn          string      // unique storage identifier
	Hctl         string      // Host:Channel:Target:Lun for SCSI
	Tran         string      // device transport type e.g. usb, nvme
	Subsystems   string      // de-duplicated chain of subsystems e.g. block, block:scsi:usb:pci, block:nvme:pci
	Rev          string      // device revision
	Vendor       string      // device vendor
	Zoned        string      // zone model
	Dax          Bool        // dax-capable device
	Children     []Blockdevice
	// HumanReadableSize func(j *json.Number)
}

// Num custom field with string, int64 & HumanReadable K,M,Gb conversion of bytes
type Num struct {
	Int64         int64
	String        string
	HumanReadable string
}

// UnmarshalJSON custom marshaling of the JSON fields
func (b *Blockdevice) UnmarshalJSON(data []byte) error {
	type Alias Blockdevice
	aux := &struct {
		Fsavail json.Number
		Fssize  json.Number
		Size    json.Number
		*Alias
	}{
		Alias: (*Alias)(b),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	var a Num
	reflA := reflect.ValueOf(&aux).Elem().Elem()
	reflB := reflect.ValueOf(&b).Elem().Elem()
	for _, fname := range []string{"Size", "Fssize", "Fsavail"} {
		if reflA.FieldByName(fname).Interface().(json.Number).String() != "" {
			i, _ := reflA.FieldByName(fname).Interface().(json.Number).Int64()
			s := reflA.FieldByName(fname).Interface().(json.Number).String()
			a.HumanReadable = ByteCountSI(i)
			a.Int64 = i
			a.String = s
			reflB.FieldByName(fname).Set(reflect.ValueOf(a))
		}
	}

	// if aux.Size != "" {
	// 	i, err := aux.Size.Int64()
	// 	check(err)
	// 	a.HumanReadable = ByteCountSI(i)
	// 	a.Int64 = i
	// 	s := aux.Size.String()
	// 	a.String = s
	// 	b.Size = a
	// }
	// if aux.Fssize != "" {
	// 	i, err := aux.Fssize.Int64()
	// 	check(err)
	// 	a.HumanReadable = ByteCountSI(i)
	// 	a.Int64 = i
	// 	s := aux.Fssize.String()
	// 	a.String = s
	// 	b.Fssize = a
	// }

	// if aux.Fsavail != "" {
	// 	i, err := aux.Fsavail.Int64()
	// 	check(err)
	// 	a.HumanReadable = ByteCountSI(i)
	// 	a.Int64 = i
	// 	s := aux.Fsavail.String()
	// 	a.String = s
	// 	b.Fsavail = a
	// }
	return nil
}

// Bool custom field deal with string, int and bool value
type Bool bool

// UnmarshalJSON custom marshaling of the JSON fields
func (b *Bool) UnmarshalJSON(data []byte) (err error) {
	switch str := strings.ToLower(strings.Trim(string(data), `"`)); str {
	case "true":
		*b = true
	case "false":
		*b = false
	default:
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return err
		}
		*b = val > 0
	}
	return nil
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Utilities: START

func compareJSONNumbers(o, n Num) bool {
	if n.Int64 > o.Int64 {
		return true
	}
	return false
}

// ByteCountSI .
func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}

// ByteCountIEC .
func ByteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}

// Utilities: END

// JSONNumber extended
// type extended JSONNumber struct {
// json.Number
// }

// HasPartitions checks if blockdevice has partitions
func (b Blockdevice) HasPartitions() bool {
	if len(b.Children) > 0 {
		return true
	}
	return false
}

// IsRunning checks if blockdevice is running
func (b Blockdevice) IsRunning() bool {
	if b.State == "running" {
		return true
	}
	return false
}

// IsMounted checks if blockdevice has mountpoint
func (b Blockdevice) IsMounted() bool {
	if b.Mountpoint != "" {
		return true
	}
	return false
}

// IsUsbTran checks if blockdevices has USB as its transport
func (b Blockdevice) IsUsbTran() bool {
	if b.Tran == "usb" {
		return true
	}
	return false
}

// USBMountedPartitionWithLargestAvailableSpace .
// func USBMountedPartitionWithLargestAvailableSpace() Blockdevice {
// 	var partition Blockdevice
// 	partition.Fsavail = "0"
// 	for _, m := range USBMountedPartitions() {
// 		if m.Fsavail != "" {
// 			if compareJSONNumbers(partition.Fsavail, m.Fsavail) {
// 				partition = m
// 			}
// 		}
// 	}
// 	return partition
// }

// USBNotMountedPartitionOfLargestSize .
func USBNotMountedPartitionOfLargestSize() Blockdevice {
	var partition Blockdevice
	partition.Size.Int64 = 0
	for _, m := range USBNotMountedPartitions() {
		if m.Size.Int64 != 0 {
			if compareJSONNumbers(partition.Size, m.Size) {
				partition = m
			}
		}
	}
	return partition
}

// USBNotMountedPartitions returns all USB disk NOT mounted partitions
func USBNotMountedPartitions() []Blockdevice {
	lsblk := GetLsblk()
	var partitions []Blockdevice

	for _, b := range lsblk.Blockdevices {
		if b.IsUsbTran() {
			if b.HasPartitions() {
				for _, p := range b.Children {
					if !p.IsMounted() {
						partitions = append(partitions, p)
					}
				}
			}
		}
	}
	return partitions
}

// USBMountedPartitions returns all USB disk mounted partitions
func USBMountedPartitions() []Blockdevice {
	lsblk := GetLsblk()
	var partitions []Blockdevice

	for _, b := range lsblk.Blockdevices {
		if b.IsUsbTran() {
			if b.HasPartitions() {
				for _, p := range b.Children {
					if p.IsMounted() {
						partitions = append(partitions, p)
					}
				}
			}
		}
	}
	return partitions
}

// GetLsblk returns lsblk output as Lsblk.Blockdevices struct
func GetLsblk() Lsblk {
	var lsblk Lsblk
	lsblkj := GetLsblkOutput()
	err := json.Unmarshal(lsblkj, &lsblk)
	check(err)
	return lsblk
}

// GetLsblkOutput []byte output from `lbslk -pabOJ`
func GetLsblkOutput() []byte {
	lsblkj, err := exec.Command("lsblk", "-pabOJ").Output()
	check(err)
	return lsblkj
}
