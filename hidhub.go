package main

import "fmt"
import "rsc.io/quote"
import "github.com/sstallion/go-hid" // package hid

func main() {
    fmt.Println( "Hello, World!" )
    fmt.Println( quote.Go() )
    
    //---* Initialise HIDAPI *-------------------------------------------------
    fmt.Println( "HIDAPI Version: ", hid.GetVersionStr() )
    hid.Init()
    
    //---* Dump the connected HID deviced *------------------------------------
    hid.Enumerate( hid.VendorIDAny, hid.ProductIDAny, func( info *hid.DeviceInfo ) error {
	fmt.Printf( "%s: ID %04x:%04x '%s' '%s'\n",
		info.Path,
		info.VendorID,
		info.ProductID,
		info.MfrStr,
		info.ProductStr )
	return nil
    })
    
    //---* Shutdown the HIDAPI *-----------------------------------------------
    hid.Exit()
}   //  main()
