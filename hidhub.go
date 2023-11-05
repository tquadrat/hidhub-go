package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sstallion/go-hid" // package hid
)

/**
 * A safe trigger: once the internal counter reaches 0, the given function will
 * be executed and the counter will be reset to the start value.
 */
type SafeTrigger struct {
	m_Mutex   sync.Mutex
	m_Counter int
	m_Start   uint
	m_Stop    bool
	m_Trigger func() error
}

func NewSafeTrigger(start uint, trigger func() error) (*SafeTrigger, error) {
	retValue := &SafeTrigger{
		m_Start:   start,
		m_Trigger: trigger}
	retValue.m_Counter = int(start)
	retValue.m_Stop = false

	//---* Done *--------------------------------------------------------------
	return retValue, nil
} //	NewSafeTrigger()

func (t *SafeTrigger) Increment() error {
	t.m_Mutex.Lock()
	defer t.m_Mutex.Unlock()

	t.m_Counter--

	var retValue error
	retValue = nil

	if t.m_Counter <= 0 {
		retValue = t.m_Trigger()
		t.m_Counter = int(t.m_Start)
	}

	//---* Done *--------------------------------------------------------------
	return retValue
} //	Increment()

func (t *SafeTrigger) Proceed() bool {
	t.m_Mutex.Lock()
	defer t.m_Mutex.Unlock()

	//---* Done *--------------------------------------------------------------
	return !t.m_Stop
} //	Proceed()

func (t *SafeTrigger) Reset() error {
	t.m_Mutex.Lock()
	defer t.m_Mutex.Unlock()

	t.m_Counter = int(t.m_Start)

	//---* Done *--------------------------------------------------------------
	return nil
} //	Reset()

func (t *SafeTrigger) Stop() {
	t.m_Mutex.Lock()
	defer t.m_Mutex.Unlock()

	t.m_Stop = true
} //	Stop()
//-----------------------------------------------------------------------------

func Heartbeat(trigger *SafeTrigger) {
	for trigger.Proceed() {
		time.Sleep(time.Second)
		trigger.Increment()
	}
} //	Heartbeat()
//-----------------------------------------------------------------------------

/**----------------------------------------------------------------------------
 * Displays the device info (all we can get).
 *
 * @param	device The pointer to the device.
 */
func ShowDeviceInfo(device *hid.Device) error {
	var status error = nil

	var deviceInfo *hid.DeviceInfo
	deviceInfo, status = device.GetDeviceInfo()
	if status == nil {
		fmt.Printf("Path .....................: %s\n", deviceInfo.Path)             // Platform-Specific Device Path
		fmt.Printf("VendorId .................: %04x\n", deviceInfo.VendorID)       // Device Vendor ID
		fmt.Printf("ProductId ................: %04x\n", deviceInfo.ProductID)      // Device Product ID
		fmt.Printf("Serial Number ............: %s\n", deviceInfo.SerialNbr)        // Serial Number
		fmt.Printf("Device Version Number ....: %04x\n", deviceInfo.ReleaseNbr)     // Device Version Number
		fmt.Printf("Manufacturer String ......: %s\n", deviceInfo.MfrStr)           // Manufacturer String
		fmt.Printf("Product String ...........: %s\n", deviceInfo.ProductStr)       // Product String
		fmt.Printf("Usage Page ...............: %04x\n", deviceInfo.UsagePage)      // Usage Page for Device/Interface
		fmt.Printf("Usage for Device/Interface: %04x\n", deviceInfo.Usage)          // Usage for Device/Interface
		fmt.Printf("USB Interface Number .....: %d\n", deviceInfo.InterfaceNbr)     // USB Interface Number
		fmt.Printf("Bus Type .................: %s\n", deviceInfo.BusType.String()) // Underlying Bus Type
	}

	if status == nil {
		for i := 0; true; i++ {
			s, status = device.GetIndexedStr(i)
			if status != nil {
				fmt.Printf("Readng Indexed String % 2d failed: %s\n", i, status.Error())
				break
			}
			fmt.Printf("Indexed String % 2d: %s\n", i, s)
		}
	}

	//---* Done *--------------------------------------------------------------
	return status
} //	ShowDeviceInfo()

/**----------------------------------------------------------------------------
 * The program's entry point.
 */
func main() {
	var status error

	//---* The usage text and other output *-----------------------------------
	usage :=
		`
When running on a Raspberry Pi computer that is connected as a USB Gagdet 
to a host computer, this program forwards the input from the captured
keyboard to that host.

The captured keyboard is identified by its VendorId and ProductId; if not
known, it can be obtained through the tool 'lsusb'. Hex numbers can be
given with the '0x' prefix.

If this program does not run as 'root', you may have to create a udev rule to
get access to the device. See for details.
`

	//---* Define the command line options *-----------------------------------
	var (
		vendorId           uint
		productId          uint
		heartbeatFrequency uint // The heartbeat frequency in seconds; 0 for no heartbeat
	)

	flag.UintVar(&vendorId, "vendorId", 0, "The VendorId of the keyboard to capture")
	flag.UintVar(&productId, "productId", 0, "The ProductId of the keyboard to capture")
	flag.UintVar(&heartbeatFrequency, "heartbeat", 30, "The heartbeat frequency; set to 0 to switch it off")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "\nUsage of %s:\n\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(flag.CommandLine.Output(), usage)
	}

	//---* Read the commandline *----------------------------------------------
	flag.Parse()

	if vendorId > 0 && productId > 0 {
		fmt.Printf("VendorId  = 0x%04x\n", vendorId)
		fmt.Printf("ProductId = 0x%04x\n", productId)
	} else {
		fmt.Fprintf(flag.CommandLine.Output(), "You need to provide both VendorId and ProductId for the keyboard to capture.\n")
		flag.Usage()
		os.Exit(2)
	}

	//---* Initialise HIDAPI *-------------------------------------------------
	fmt.Println("HIDAPI Version: ", hid.GetVersionStr())
	hid.Init()
	defer hid.Exit()

	//---* Dump the connected HID devices *------------------------------------
	hid.Enumerate(uint16(vendorId), uint16(productId), func(info *hid.DeviceInfo) error {
		fmt.Printf("%s: ID %04x:%04x '%s' '%s'\n",
			info.Path,
			info.VendorID,
			info.ProductID,
			info.MfrStr,
			info.ProductStr)
		return nil
	})

	//---* Start the Heartbeat *-----------------------------------------------
	if heartbeatFrequency > 0 {
		message := "Heartbeat Now!"

		/**
		 * The heartbeat trigger.
		 */
		heartbeatTrigger, _ := NewSafeTrigger(3, func() error {
			fmt.Println(message)
			return nil
		})

		go Heartbeat(heartbeatTrigger)
		defer heartbeatTrigger.Stop()
	}

	//---* Open the captured keyboard *----------------------------------------
	var device *hid.Device
	device, status = hid.OpenFirst(uint16(vendorId), uint16(productId))
	if status != nil {
		fmt.Printf("Error: %s\nAborted!\n", status.Error())
		os.Exit(2)
	}
	defer device.Close()

	//status = ShowDeviceInfo(device)
	if status := ShowDeviceInfo(device); status != nil {
		fmt.Printf("Error: %s\nAborted!\n", status.Error())
		os.Exit(2)
	}

	b := make([]byte, 65)

	//---* Toggle LED (cmd 0x80). The first byte is the report number (0x0) ---
	for i := 1; i < 20; i++ {
		b[0] = 0x0
		b[1] = 0x80
		if _, status := device.Write(b); status != nil {
			fmt.Printf("Error on blinking keyboard: %s\nAborted!\n", status.Error())
			os.Exit(2)
		}
	}
	if heartbeatFrequency > 0 {
		time.Sleep(time.Minute)
	}

	//---* Done *--------------------------------------------------------------
	fmt.Println("Done!")
} //  main()
