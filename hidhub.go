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

func (t *SafeTrigger) IncrementTrigger() error {
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
} //	IncrementTrigger()

func (t *SafeTrigger) ProceedTrigger() bool {
	t.m_Mutex.Lock()
	defer t.m_Mutex.Unlock()

	//---* Done *--------------------------------------------------------------
	return !t.m_Stop
}

func (t *SafeTrigger) ResetTrigger() error {
	t.m_Mutex.Lock()
	defer t.m_Mutex.Unlock()

	t.m_Counter = int(t.m_Start)

	//---* Done *--------------------------------------------------------------
	return nil
} //	ResetTrigger()

func (t *SafeTrigger) StopTrigger() {
	t.m_Mutex.Lock()
	defer t.m_Mutex.Unlock()

	t.m_Stop = true
} //	StopTrigger()
//-----------------------------------------------------------------------------

/**
 * The trigger function for the heartbeat.
 */
func HeartbeatFunc() error {
	fmt.Println("Heartbeat!")

	//---* Done *--------------------------------------------------------------
	return nil
} //	HeartbeatFunc()

func Heartbeat(trigger *SafeTrigger) {
	for true {
		time.Sleep(time.Second)
		trigger.IncrementTrigger()
	}
} //	Heartbeat()
//-----------------------------------------------------------------------------

func main() {
	//---* The usage text and other output *-----------------------------------
	usage :=
		`
When running on a Raspberry Pi computer that is connected as a USB Gagdet 
to a host computer, this program forwards the input from the captured
keyboard to that host.

The captured keyboard is identified by its VendorId and ProductId; if not
known, it can be obtained through the tool 'lsusb'. Hex numbers can be
given with the '0x' prefix.
`

	//---* Define the command line options *-----------------------------------
	var (
		vendorId  uint
		productId uint
	)

	flag.UintVar(&vendorId, "vendorId", 0, "The VendorId of the keyboard to capture")
	flag.UintVar(&productId, "productId", 0, "The ProductId of the keyboard to capture")

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

	//---* Dump the connected HID deviced *------------------------------------
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
	message := "Heartbeat!"

	/**
	 * The heartbeat trigger.
	 */
	heartbeatTrigger, _ := NewSafeTrigger(3, func() error {
		fmt.Println(message)
		return nil
	})

	go Heartbeat(heartbeatTrigger)
	defer heartbeatTrigger.StopTrigger()

	time.Sleep(time.Minute)

	//---* Done *--------------------------------------------------------------
	fmt.Println("Done!")
} //  main()
